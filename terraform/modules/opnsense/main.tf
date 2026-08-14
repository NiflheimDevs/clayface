
resource "libvirt_domain" "opnsense" {
  name        = var.name
  memory      = 3000
  memory_unit = "MiB"
  vcpu        = 3
  type        = "kvm"

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
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

    interfaces = [
      {
        type = "bridge"

        model = {
          type = "virtio"
        }

        source = {
          bridge = {
            bridge = "vm-br0"
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
            bridge = "vm-wan0"
          }
        }
      },
    ]
  }
}
