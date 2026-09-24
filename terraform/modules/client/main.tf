locals {
  mac_seed = sha256(var.name)
  vm_mac = format("52:54:00:%s:%s:%s",
    substr(local.mac_seed, 0, 2),
    substr(local.mac_seed, 2, 2),
    substr(local.mac_seed, 4, 2),
  )
}

resource "libvirt_volume" "client" {
  name     = "${upper(var.name)}.qcow2"
  pool     = "default"
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

resource "libvirt_domain" "client" {
  name        = var.name
  memory      = 4096
  memory_unit = "MiB"
  vcpu        = 2
  type        = "kvm"

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"

    firmware = "efi"

    firmware_info = {
      features = [
        { name = "enrolled-keys", enabled = "no" },
        { name = "secure-boot", enabled = "yes" },
      ]
    }

    loader          = var.uefi_loader_path
    loader_readonly = "yes"
    loader_secure   = "yes"
    loader_type     = "pflash"
    loader_format   = "raw"

    nv_ram = {
      template        = var.uefi_nvram_template_path
      template_format = "raw"
      format          = "raw"
      nv_ram          = "/var/lib/libvirt/qemu/nvram/${var.name}_VARS.fd"
    }
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
            volume = libvirt_volume.client.name
          }
        }

        target = {
          dev = "sda"
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

    tpms = [
      {
        model = "tpm-crb"
        backend = {
          emulator = {
            version = "2.0"
          }
        }
      }
    ]
  }
}
