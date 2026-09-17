# Windows base images — build guide

Build recipes for the two Windows base images the lab boots VMs from. Each
one is the qcow2 backing store for a terraform module's `base_image_path`
variable; the module stacks a thin per-VM overlay on top, so one base can
serve any number of VMs.

The unattend file in each case makes every VM built from that base
Ansible-reachable on first boot, so `deploy.sh` promotes the DC and joins the
client with zero console interaction.

|  | `ws-base.qcow2` | `client-base.qcow2` |
|---|---|---|
| OS | Windows Server 2025 (build 26100), Desktop Experience | Windows 11 26H1, business editions (Enterprise/Education/Pro) |
| Terraform module | `modules/domaincontroller` | `modules/client` |
| Unattend | `unattend-dc.xml` | `unattend-cl.xml` |
| Firmware | **BIOS** | **UEFI**, secure boot (`OVMF_CODE.secboot.4m.fd` + `OVMF_VARS.4m.fd`) |
| TPM | none | **`tpm-crb`, version 2.0** (host needs `swtpm`) |
| Machine / disk / NIC | q35 / SATA / e1000e | q35 / SATA / e1000e |
| Role in the lab | forest root `dc01` (`clayface.local`) | domain-joined workstation `client01` |
| Playbook | `playbooks/dc.yml` | `playbooks/client.yml` |

**The firmware row is the one that breaks the naive copy.** The two modules
are near-identical text, but a client image built under UEFI lands on a GPT
disk with an EFI System Partition, and the DC module's BIOS machine has no
MBR boot sector to find — it drops to the SeaBIOS "no bootable device"
prompt. Conversely, a BIOS-built Server image will not boot under the
client module's UEFI loader. **The builder VM's firmware must match the
terraform module's, always.**

Also read: `CLAUDE.md` (the DC section) and `docs/redclay.yaml` for where
these images sit in the lab.

## Keep the *server* image role-free

This section applies to `ws-base.qcow2` only. There is no equivalent rule for
the client image — it carries no roles either, but only because a workstation
does not have any.

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
OOBE "set the Administrator password" screen on next boot — only the console
(SPICE) can get past it, which breaks automation. The unattend file answers
OOBE, sets the known lab password, and enables WinRM + firewall on all
profiles during specialize (before any logon). See the header comment of each
unattend file for exactly what it does.

## One-time build

The steps are the same for both images; the parenthesised notes are where
they differ. Build each on any Windows-capable hypervisor.

1. **Create a builder VM** with the install ISO. Match the terraform module's
   hardware so drivers generalize cleanly: q35, SATA disk, e1000e NIC (no
   virtio — keeps the image free of guest-agent dependencies), and:
   - `ws-base`: **BIOS** boot, no TPM.
   - `client-base`: **UEFI** with secure boot, plus TPM 2.0. Windows 11 setup
     refuses to proceed without the TPM, so this is not optional.

2. **Install Windows** (Server 2025 Desktop Experience / Windows 11 26H1
   business edition) manually, once. Any password works here — the unattend
   resets it to `Admin@123`.

3. **Install nothing else.** No server roles on `ws-base`. On `client-base`,
   do not join a domain — `client.yml` does that per VM.

4. **Patch Windows before capturing.** Update to the latest cumulative update.
   For Server 2025 that means at minimum `26100.4946` (KB5063878), the only
   Microsoft-documented Server 2025 sysprep fix ("the boot file configuration
   is not properly updated, resulting in push-button reset options not
   working"). Capturing an unpatched 26100 image is where most sysprep
   failures come from. Patch the client image for the same reason — an
   unpatched generalize is the least-tested path.

5. **On the client image only: quiet the Store.** Windows 11 ships with far
   more provisioned AppX packages than Server, and the Store keeps updating
   them. AppX is the single most common sysprep failure mode (see
   troubleshooting). Disconnect the network, or pause Store updates, before
   generalizing — and re-check `Get-AppxPackage -AllUsers` immediately before
   the sysprep run.

6. **Check the base's virtual size.** The terraform module's overlay
   `capacity` must be **≥** this number, or the overlay is created smaller
   than its own backing file:
   ```bash
   sudo qemu-img info /var/lib/libvirt/images/<base>.qcow2
   ```
   Read the *virtual size*, not the disk usage — qcow2 is sparse. Both bases
   in this lab are 20 GiB virtual, under the 25 GiB the DC module declares
   and the 30 GiB the client module declares. Check rather than assume: a
   Windows 11 image built on the installer's default disk can be far larger
   (64 GB is common), which would need the module's `capacity` raised.

7. **Copy the unattend file** into place on the guest, named `unattend.xml`:
   ```bat
   mkdir C:\Windows\System32\sysprep 2>nul
   copy unattend-dc.xml C:\Windows\System32\sysprep\unattend.xml
   ```
   The location is a convenience, not a requirement — see the troubleshooting
   section.

8. **Generalize and shut down** (as Administrator):
   ```bat
   C:\Windows\System32\Sysprep\sysprep.exe ^
       /generalize /oobe /shutdown ^
       /unattend:C:\Windows\System32\sysprep\unattend.xml
   ```
   The VM shuts down when done. Do **not** boot it again — sysprep'd images
   must stay untouched (booting runs OOBE and dirties the state).

   `/generalize` rearms the activation grace period, but it has a **three-rearm
   limit**. If you hit it, build a fresh builder VM rather than fighting it.

9. **Publish the base** to every hypervisor that boots VMs from it, exactly at
   the path the terraform module expects:
   ```bash
   # ws-base.qcow2      -> modules/domaincontroller's base_image_path
   # client-base.qcow2  -> modules/client's base_image_path
   qemu-img convert -O qcow2 <builder-disk> \
       /var/lib/libvirt/images/<base>.qcow2
   # backing stores are opened read-only by qemu; keep it that way and
   # never boot the base directly
   ```
   If a VM already exists on top of an older base image, keep the old file
   until that VM's disk has been recreated — `qemu-img info --backing-chain`
   on the VM's disk names the backing file it needs, and a missing backing
   file means the domain will not start.

10. **Undefine the builder VM.** Once the base is a backing file it is
    read-only forever, and *booting it corrupts every overlay stacked on top,
    silently and later* — not immediately, which is what makes it dangerous:
    ```bash
    virsh -c qemu:///system undefine <builder-domain> --nvram
    ```
    Keep a pre-sysprep copy (`client-base-bck.qcow2` style) as the escape
    hatch. Do not confuse the backup with the base — only the base may be
    defined as a backing file.

11. **Deploy**: `./deploy.sh`. See `CLAUDE.md` for what each playbook does.

## Sysprep troubleshooting (Server 2025 / build 26100, Windows 11 24H2+)

Sysprep failures log to `C:\Windows\System32\Sysprep\Panther\setupact.log`
and `setuperr.log`. Read `setuperr.log` first — it names the offending package
or component exactly. The three known failure modes:

- **AppX validation** — `SYSPRP Package <name> was installed for a user, but
  not provisioned for all users`, or `Sysprep ... 0x80073cf2`. Sysprep has a
  provider that cleans up AppX packages, and it only works when a package is
  per-user or provisioned for all users. Fix: list with
  `Get-AppxPackage -AllUsers`, remove with `Remove-AppxPackage`, then
  `Remove-AppxProvisionedPackage -Online`; keep the Store offline and let no
  update land mid-generalize. This is the most common 26100 failure, and it is
  *more* likely on the Windows 11 client image — see build step 5.
- **`AppXSvc` disabled** — `SYSPRP ActionPlatform::LaunchModule: Failure
  occurred while executing 'SysprepGeneralizeValidate' from
  C:\Windows\System32\AppxSysprep.dll; dwRet = 0x422`. Sysprep requires the
  AppX Deployment Service to be Manual or Automatic. Check and fix with:
  ```bat
  reg add "HKLM\SYSTEM\CurrentControlSet\Services\AppXSvc" /v Start /t REG_DWORD /d 3 /f
  ```
- **Unpatched 26100** — see build step 4. Patch before capturing.

Two facts worth knowing when debugging:

- **The `/unattend:<path>` location does not matter.** Windows caches a copy to
  `%WINDIR%\Panther\unattend.xml`, and the next boot reads *that*, not your
  original path. The copy into `System32\sysprep` is one implicit search
  location among others; keeping it is harmless.
- **The random hostname is expected.** `<ComputerName>*</ComputerName>` yields
  a random 15-character name, and generalize does not touch the NIC MAC. The
  rename is then done from Ansible — `dc.yml` before promotion, `client.yml`
  as part of the domain join.

## Notes & caveats

- **The pinned-MAC chain.** `dc.yml` and `client.yml` cannot resolve
  `<name>.clayface` until the guest has been renamed, and the guest boots
  under a random name. The MAC is the one identifier terraform knows *before*
  the VM exists and the guest cannot change, so the DC and client modules
  derive a deterministic MAC from the VM name, and `playbooks/opnsense.yml`
  turns it into a DHCP reservation plus an A record in OPNsense. Adding a VM
  to `lab.yaml` therefore needs both `module:` and `ip:` — see `CLAUDE.md`.
- **Passwords in cleartext**: `Admin@123` is committed in both unattend files
  (PlainText=true). Lab-only, deliberately weak, consistent with
  `LAB_WIN_ADMIN_PASS` default in the Ansible inventory. It appears in three
  places — the inventory script, the playbooks, and the unattend files —
  change all of them together if you rotate. The client image additionally
  keeps a `builder` / `ChangeMe123!` account as a console fallback when WinRM
  is down; the Ansible inventory deliberately does not know about it.
- **Timezone/locale**: the unattends leave them at install defaults. Add
  `<TimeZone>`, `<InputLocale>`, `<UILanguage>` to the oobeSystem component if
  the lab needs specific ones.
- **Multiple VMs from one base**: `/generalize` gives every VM from a base a
  unique machine SID, so `dc02` or `client02` are fine. The terraform dc and
  client modules name each VM's overlay after the VM
  (`upper(var.name).qcow2`), so a second VM does not collide with the first's
  disk.
- **Validation**: the XML files here were only checked for well-formedness. If
  you have Windows ADK installed, validate with Windows System Image Manager
  (WSIM) before building.
