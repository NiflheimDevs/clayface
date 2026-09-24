
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
    # firewall rules at the wrong interface.
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
    ]
  }
}
