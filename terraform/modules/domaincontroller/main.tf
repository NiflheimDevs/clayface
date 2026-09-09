resource "libvirt_volume" "dc01" {
  name     = "DC01.qcow2"
  pool     = "default"
  capacity = 25 * 1024 * 1024 * 1024 # 25 GiB

  target = {
    format = {
      type = "qcow2"
    }
  }

  backing_store = {
    path = "/var/lib/libvirt/images/ws-ad-base.qcow2"

    format = {
      type = "qcow2"
    }
  }
}

resource "libvirt_domain" "dc" {
  name        = var.name
  memory      = 4096
  memory_unit = "MiB"
  vcpu        = 2
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
          volume = {
            pool   = "default"
            volume = libvirt_volume.dc01.name
          }
        }

        target = {
          dev = "sdb"
          bus = "sata"
        }
      }
    ]

    cpu = {
      mode = "host-model"
    }

    interfaces = [
      {
        type = "bridge"

        model = {
          type = "e1000e"
        }

        source = {
          bridge = {
            bridge = "vm-br0"
          }
        }
      }
    ]
    graphics = [
      {
        type = "spice"

        spice = {
          auto_port = true
        }
      }
    ]

    videos = [
      {
        model = {
          type    = "qxl"
          heads   = 1
          primary = "yes"
          ram     = 65536
          vram    = 65536
          vga_mem = 16384
        }
      }
    ]
  }
}
