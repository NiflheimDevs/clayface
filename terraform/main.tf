terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.9.8"
    }
  }
}
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "vmtest" {
  name        = "vmtest"
  memory      = "1024"
  memory_unit = "MiB"
  vcpu        = 1
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
            file = "/var/lib/libvirt/images/alpinelinux3.23"
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
      }
    ]
  }

}

output "vm-id" {
  value = libvirt_domain.vmtest.id
}
