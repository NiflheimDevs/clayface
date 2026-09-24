locals {
  # Same deterministic MAC scheme as the dc and client modules: the VM name is
  # the only identifier terraform knows before the VM exists and the guest
  # cannot change, which is what lets OPNsense pin a DHCP reservation and a DNS
  # record for it (playbooks/opnsense.yml). An Ubuntu guest does keep the
  # hostname baked into its image, so this is not strictly required here the
  # way it is for a sysprepped Windows guest — but the lab has one rule for
  # every VM, and a reservation is what gives app01 its static address.
  mac_seed = sha256(var.name)
  vm_mac = format("52:54:00:%s:%s:%s",
    substr(local.mac_seed, 0, 2),
    substr(local.mac_seed, 2, 2),
    substr(local.mac_seed, 4, 2),
  )
}

resource "libvirt_volume" "app" {
  name = "${upper(var.name)}.qcow2"
  pool = "default"

  # Must be at least the base image's virtual size or qemu refuses to open the
  # overlay: the base was built on a 25 GiB disk, so 40 leaves the containers,
  # the images tarball and the PostgreSQL data directory real room.
  capacity = 40 * 1024 * 1024 * 1024 # 40 GiB

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

resource "libvirt_domain" "app" {
  name        = var.name
  memory      = 4096
  memory_unit = "MiB"
  vcpu        = 2
  type        = "kvm"

  # Plain BIOS, like the domain controller module and unlike the client module:
  # the Ubuntu image is a BIOS install. Booting it under UEFI finds no ESP and
  # the firmware stops at "no bootable device", so the firmware here has to
  # match the one the base image was installed with.
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
            volume = libvirt_volume.app.name
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
            bridge = var.bridge
          }
        }

        mac = {
          address = local.vm_mac
        }
      }
    ]

    # No graphics device on purpose. This VM is managed over SSH and its whole
    # workload is containers, so a SPICE console would be an unused attack
    # surface on a host that is supposed to be the one exposed to the
    # attacker. If a console is ever needed, add one temporarily rather than
    # leaving it in the module.
  }
}
