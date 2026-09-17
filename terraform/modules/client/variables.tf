variable "name" {
  type = string
}

variable "base_image_path" {
  description = "Absolute path of the qcow2 base image on the hypervisor (built per base-image/README.md). Must exist on every host that runs a client VM."
  type        = string
  default     = "/var/lib/libvirt/images/client-base.qcow2"
}

# UEFI firmware paths as variables rather than literals: the layout differs
# between distributions (Arch ships /usr/share/edk2/x64/, Debian/Ubuntu
# /usr/share/OVMF/), so a host with a different firmware package can point
# at its own without editing the module.
variable "uefi_loader_path" {
  description = "Path to the secure-boot UEFI firmware (OVMF code) on the hypervisor."
  type        = string
  default     = "/usr/share/edk2/x64/OVMF_CODE.secboot.4m.fd"
}

variable "uefi_nvram_template_path" {
  description = "Path to the UEFI variable-store template (OVMF vars) on the hypervisor. Copied once per VM into libvirt's nvram directory."
  type        = string
  default     = "/usr/share/edk2/x64/OVMF_VARS.4m.fd"
}
