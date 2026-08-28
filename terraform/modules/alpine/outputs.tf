output "name" {
  description = "Name of the libvirt domain created by this module"
  value       = libvirt_domain.alpine.name
}
