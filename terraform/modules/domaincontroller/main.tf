locals {
  mac_seed = sha256(var.name)
  vm_mac = format("52:54:00:%s:%s:%s",
    substr(local.mac_seed, 0, 2),
    substr(local.mac_seed, 2, 2),
    substr(local.mac_seed, 4, 2),
  )
}

resource "libvirt_volume" "dc01" {
  name     = "${upper(var.name)}.qcow2"
  pool     = "default"
  capacity = 25 * 1024 * 1024 * 1024 # 25 GiB

  target = {
    format = {
      type = "qcow2"
    }
  }

  backing_store = {
    path = var.base_image_path

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

  cpu = {
    mode = "host-model"
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

    interfaces = [
      {
        type = "bridge"

        model = {
          type = "e1000e"
        }

        source = {
          bridge = {
            bridge = var.bridge
          }
        }

        mac = {
          address = local.vm_mac
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
