resource "libvirt_volume" "crusader" {
  name = "${upper(var.name)}.qcow2"
  pool = "default"

  capacity = 30 * 1024 * 1024 * 1024

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

# DHCP is intentionally provided by OPNsense. There is no cloud-init seed:
# the VM starts with the base image's normal NetworkManager DHCP behavior, and
# Ansible configures the guest once its DHCP/DNS name is available.
resource "libvirt_domain" "crusader" {
  name        = var.name
  memory      = 4096
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
            volume = libvirt_volume.crusader.name
          }
        }
        target = {
          dev = "vda"
          bus = "virtio"
        }
        boot = {
          order = 1
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
            bridge = var.bridge
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
