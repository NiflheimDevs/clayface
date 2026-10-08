Windows Workstation is actually an employee's windows PC or laptop.
They get domain joined to AD.
more information in on how it works in [[Playground/Components/Active Directory|Active Directory]]

This was mostly vibe coded since there were a lot of special powershell and AD commands involved.

The pain for this part was the monitor configuration and motherboard emulation.
I won't go into the details cause it was mostly trial and error.
I caveat to know:
for TPM, install this:
```bash
sudo pacman -S swtpm
sudo systemctl restart libvirtd
```