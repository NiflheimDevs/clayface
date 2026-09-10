# Windows base image (`ws-base.qcow2`) — build guide

Build recipe for the Windows Server 2025 base image that
`terraform/modules/domaincontroller` uses as its qcow2 backing store (the
module's `base_image_path` variable). The unattend file makes every VM built
from this base Ansible-reachable on first boot, so `deploy.sh` →
`playbooks/dc.yml` installs AD DS, promotes the DC and wires up DNS with zero
console interaction.

**This image deliberately contains no server roles.** The AD DS role is
installed by `dc.yml` during promotion — see "Keep the image role-free" below.

## Keep the image role-free

Microsoft's *Sysprep support for server roles* table lists **AD DS as not
supported** under `/generalize`:

> However, if you run the Sysprep command together with the /generalize option
> against an installation of a server, and you are using an unsupported server
> role, those roles may not function after the imaging and deployment process
> is completed. Therefore, you must enable and configure any server roles that
> do not support Sysprep after you have performed the imaging and deployment
> process.

That applies to the role being *installed*, not just to a promoted domain
controller — generalizing a server with `AD-Domain-Services` installed is
outside the supported statement, and it is what broke the first build attempt
(see `red-clay/Report/Actions/AD.md`). Role-free base image plus
`microsoft.ad.domain` in `dc.yml` is the supported path, and `dc.yml` installs
the role itself, so nothing is lost.

## Why sysprep with `/unattend:`

A plain `sysprep.exe /generalize /oobe /shutdown` leaves the image at the
OOBE "set the Administrator password" screen on next boot — only the
console (SPICE) can get past it, which breaks automation. The unattend
file answers OOBE, sets the known lab password, and enables WinRM +
firewall on all profiles during specialize (before any logon). See the
header of `unattend-dc.xml` for exactly what it does.

## One-time build (on any Windows-capable hypervisor)

1. **Create a builder VM** with the Server 2025 ISO. Match the DC module's
   hardware so drivers generalize cleanly: q35, UEFI-less BIOS boot, SATA
   disk, e1000e NIC (no virtio — keeps the image free of guest-agent
   dependencies).

2. **Install Windows Server 2025** (Desktop Experience) manually, once.
   Any password works here — the unattend resets it to `Admin@123`.

3. **Install NO server roles.** No AD DS, no DHCP, no DNS — nothing. AD DS is
   added by `dc.yml` at promotion time.

4. **Patch Windows before capturing.** Update to the latest cumulative update;
   at minimum `26100.4946` (KB5063878), which is the only Microsoft-documented
   Server 2025 sysprep fix ("the boot file configuration is not properly
   updated, resulting in push-button reset options not working"). Capturing an
   unpatched 26100 image is where most sysprep failures come from.

5. **Copy the unattend file** into place on the guest:
   ```bat
   mkdir C:\Windows\System32\sysprep 2>nul
   copy unattend-dc.xml C:\Windows\System32\sysprep\unattend.xml
   ```
   The location is a convenience, not a requirement — see the troubleshooting
   section.

6. **Generalize and shut down** (as Administrator):
   ```bat
   C:\Windows\System32\Sysprep\sysprep.exe ^
       /generalize /oobe /shutdown ^
       /unattend:C:\Windows\System32\sysprep\unattend.xml
   ```
   The VM shuts down when done. Do **not** boot it again — sysprep'd
   images must stay untouched (booting runs OOBE and dirties the state).

7. **Publish the base** to every hypervisor that builds DC VMs, exactly
   at the path the terraform module expects:
   ```bash
   # from wherever the builder VM's disk lives:
   qemu-img convert -O qcow2 <builder-disk> \
       /var/lib/libvirt/images/ws-base.qcow2
   # backing stores are opened read-only by qemu; keep it that way and
   # never boot the base directly
   ```
   If a VM already exists on top of an older base image, keep the old file
   until that VM's disk has been recreated — `qemu-img info --backing-chain`
   on the VM's disk names the backing file it needs, and a missing backing
   file means the domain will not start.

8. **Deploy**: `./deploy.sh` (or terraform apply + the ansible playbooks).
   First boot of `dc01` runs specialize/OOBE unattended (~2–5 min), then
   `dc.yml` waits for WinRM, renames to `dc01`, enforces `Admin@123`,
   installs AD DS, creates the `clayface.local` forest, reboots, points the
   DC's DNS at itself and forwards the rest to OPNsense, and verifies the
   domain's A and DC-locator SRV records resolve.

## Sysprep troubleshooting (Server 2025 / build 26100)

Sysprep failures log to `C:\Windows\System32\Sysprep\Panther\setupact.log`
and `setuperr.log`. Read `setuperr.log` first — it names the offending package
or component exactly. The three known failure modes:

- **AppX validation** — `SYSPRP Package <name> was installed for a user, but
  not provisioned for all users`, or `Sysprep ... 0x80073cf2`. Sysprep has a
  provider that cleans up AppX packages, and it only works when a package is
  per-user or provisioned for all users. Fix: list with
  `Get-AppxPackage -AllUsers`, remove with `Remove-AppxPackage`, then
  `Remove-AppxProvisionedPackage -Online`; keep the Store offline and let no
  update land mid-generalize. This is the most common 26100 failure.
- **`AppXSvc` disabled** — `SYSPRP ActionPlatform::LaunchModule: Failure
  occurred while executing 'SysprepGeneralizeValidate' from
  C:\Windows\System32\AppxSysprep.dll; dwRet = 0x422`. Sysprep requires the
  AppX Deployment Service to be Manual or Automatic. Check and fix with:
  ```bat
  reg add "HKLM\SYSTEM\CurrentControlSet\Services\AppXSvc" /v Start /t REG_DWORD /d 3 /f
  ```
- **Unpatched 26100** — see step 4. Patch before capturing.

Two facts worth knowing when debugging:

- **The `/unattend:<path>` location does not matter.** Windows caches a copy to
  `%WINDIR%\Panther\unattend.xml`, and the next boot reads *that*, not your
  original path. The copy into `System32\sysprep` is one implicit search
  location among others; keeping it is harmless.
- **The random hostname is expected.** `<ComputerName>*</ComputerName>` yields
  a random 15-character name, and generalize does not touch the NIC MAC — so
  the pinned-MAC → OPNsense DHCP reservation → `dc01.clayface` chain survives
  the rebuild untouched. `dc.yml` renames the host before promotion.

## Notes & caveats

- **Password in cleartext**: `Admin@123` is committed in
  `unattend-dc.xml` (PlainText=true). Lab-only, deliberately weak,
  consistent with `LAB_WIN_ADMIN_PASS` default in the Ansible inventory.
  Change both together if you rotate.
- **Timezone/locale**: the unattend leaves them at install defaults. Add
  `<TimeZone>`, `<InputLocale>`, `<UILanguage>` to the oobeSystem
  component if the lab needs specific ones.
- **Multiple DCs later**: `/generalize` gives every VM from this base a
  unique machine SID — this is the supported path if you ever add dc02.
  The terraform dc module names each VM's volume after the VM
  (`upper(var.name).qcow2`), so a second DC does not collide with dc01's disk.
- **Validation**: the XML here was only checked for well-formedness. If
  you have Windows ADK installed, validate with Windows System Image
  Manager (WSIM) before building.
