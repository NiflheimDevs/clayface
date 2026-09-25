
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
    ]
  }
}
