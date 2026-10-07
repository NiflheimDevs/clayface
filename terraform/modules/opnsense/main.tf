
locals {
  # The DMZ NIC's MAC, derived from the VM name the same way every other
  # module derives its VM's. A distinct seed suffix keeps it out of the space
  # the other modules hash into, so it cannot collide with a VM called
  # "opnsense01-dmz" or with opnsense01's own LAN MAC.
  #
  # It is pinned rather than left to libvirt for one reason: NIC order in this
  # resource is device order, and the guest names its interfaces by position
  # (vtnet0, vtnet1, vtnet2). If a future edit ever renumbers them, a pinned
  # MAC turns "the firewall rules now point at the wrong interface" into a
  # named assertion failure in ansible/playbooks/opnsense_dmz.yml.
  dmz_mac_seed = sha256("${var.name}-dmz")
  dmz_mac = format("52:54:00:%s:%s:%s",
    substr(local.dmz_mac_seed, 0, 2),
    substr(local.dmz_mac_seed, 2, 2),
    substr(local.dmz_mac_seed, 4, 2),
  )

  # The uplink NIC's MAC, pinned for the DMZ's reason and one more of its own.
  #
  # The DMZ's is pinned so a renumbering becomes a named assertion failure
  # rather than a firewall ruleset pointing at the wrong interface. The same
  # applies here. The extra reason is that OPNsense remembers which interface
  # an assignment belongs to partly by MAC: this NIC must be *assigned* by hand
  # in the UI before it does anything, and a MAC that changed on every apply
  # would make that assignment go stale without saying so.
  #
  # LAN and WAN are unpinned, and that asymmetry is deliberate rather than an
  # oversight — but it is also why redefining this domain deserves care. See
  # the note on this resource in docs/network-design.md section 8.2.
  uplink_mac_seed = sha256("${var.name}-uplink")
  uplink_mac = format("52:54:00:%s:%s:%s",
    substr(local.uplink_mac_seed, 0, 2),
    substr(local.uplink_mac_seed, 2, 2),
    substr(local.uplink_mac_seed, 4, 2),
  )
}

resource "libvirt_domain" "opnsense" {
  name        = var.name
  memory      = 2500
  memory_unit = "MiB"
  vcpu        = 3
  type        = "kvm"

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  features = {
    acpi = true
    apic = {}
  }

  devices = {
    disks = [
      {
        driver = {
          type = "qcow2"
          name = "qemu"
        }

        source = {
          file = {
            file = "/var/lib/libvirt/images/opnsense.qcow2"
          }
        }

        target = {
          dev = "vda"
          bus = "virtio"
        }
      }
    ]

    # NIC order is device order, and the guest names its interfaces by it
    # (vtnet0, vtnet1, ...): disks and interfaces share the ordering space, so
    # anything added here must be APPENDED, never inserted ahead of an
    # existing NIC. Inserting one renumbers the WAN and quietly points the
    # firewall rules at the wrong interface. The DMZ NIC is third for exactly
    # that reason: appending gives it vtnet2 and leaves LAN/WAN alone.
    #
    # The uplink is fourth for the same reason again — vtnet3, LAN/WAN/DMZ
    # untouched. It must stay last: it is the NIC OPNsense has to be told about
    # by hand, and a hand-made assignment is the thing a renumbering breaks
    # most quietly.
    interfaces = [
      {
        type = "bridge"

        model = {
          type = "virtio"
        }

        source = {
          bridge = {
            bridge = var.bridge
          }
        }
      },
      {
        type = "bridge"

        model = {
          type = "virtio"
        }

        source = {
          bridge = {
            bridge = var.wan_bridge
          }
        }
      },
      {
        type = "bridge"

        model = {
          type = "virtio"
        }

        mac = {
          address = local.dmz_mac
        }

        source = {
          bridge = {
            bridge = var.dmz_bridge
          }
        }
      },
      {
        type = "bridge"

        model = {
          type = "virtio"
        }

        mac = {
          address = local.uplink_mac
        }

        source = {
          bridge = {
            bridge = var.uplink_bridge
          }
        }
      },
    ]
  }
}
