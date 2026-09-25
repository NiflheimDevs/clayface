Ok first we install OPNSense.
the OPNSense VM will need atleast two V NICs to function. 1 act as a LAN and 2nd one act as WAN.

*fun fact: two vcpu cores was not enough and it kept stalling(evidence: dashboard didn't load)*

we set the LAN interface to vm-lan0 and the WAN interface as a NAT interface managed by virtio. this way WAN can connect to internet and the devices on the LAN can do so as well.

one good thing about opnsense is that it has gui. for accessing the gui, you need to give an ip address to the vm-lan0 interface.

## The DMZ interface (added later, for the segmentation wave)

a third NIC went on the VM, appended **after** the LAN and WAN ones. the order
matters: the guest names its interfaces by position (vtnet0, vtnet1, vtnet2 in
that order), and the firewall rules name interfaces, so inserting a NIC ahead of
an existing one renumbers everything after it and points the rules at the wrong
device. appended third, it lands on vtnet2 and the first two are untouched.

terraform pins a MAC for it (derived from the VM name with a `-dmz` suffix, so
it cannot collide with opnsense01's own LAN MAC) and
`ansible/playbooks/opnsense_dmz.yml` finds the device by that MAC rather than by
name. that is what turns a renumbering into a loud failure instead of a boundary
that quietly stopped existing.

the assignment itself is API-settable — `interfaces/assignment/addItem` +
`reconfigure` — and the playbook does it. **the interface's IPv4 address is
not.** on 26.7 the assignment model covers assignment only: a `setItem` that
carries the address fields answers `{"result":"saved"}` and writes none of
them, and `interfaces/assignment/pending` 404s. the address lives in
`config.xml`, written by the legacy `interfaces.php` form, which authenticates
by GUI session plus CSRF and has no API-key route in. only 27.1 makes it
settable. so it is a one-time hand step, and the playbook asserts it and stops
with the UI steps if it is missing:

```
Interfaces -> Assignments -> the DMZ row
  IPv4 Configuration Type : Static IPv4
  IPv4 address            : 10.0.10.1/24
  Block private networks  : OFF   (the DMZ is RFC1918)
  Block bogon networks    : OFF
  Enabled                 : ON
```

one thing that made deferring it safe: OPNsense's generated ruleset has **no
pass-any-per-interface rule**. the automatic rules are narrow and named (DHCP,
sshlockout, virusprot, IPv6 ICMP, a LAN-only anti-lockout). an interface with no
pass rule written for it is denied, so the DMZ is fail-closed from the moment it
comes up — an unaddressed DMZ interface is unreachable, not unfiltered.

see `docs/network-design.md` and `docs/opnsense-image.md` for the long version.

At first, I didn't configure the OPNsense with dhcp and connected everything with static ip and no dns.

the two alpine machines set their own ip address and put the opnsense ip address as their gateway and default route
and WALLAH! everything works super simply

the next step is configuring DHCP server and DNS for machines receiving ip from DHCP server.

i have skipped firewall rules for now. will come back to it whenever we needed to fix everything.

configuring dhcp on opnsense is too simple tbh. no need to document anything. (i love gui. if it was cli i would have gone crazy configuring this. jk, i'd die for cli)

## DNS for leases

Ok this one is a bit confusing so i have to document it.
OPNSense has a tab for dnsmasq and dhcp. in there, it has a configuration: DHCP. in there, configure DHCP default domain to for example "lab" and enable DHCP local domain

now dnsmasq will record hostname.lab to their ip address

now we have to configure unbound dns to forward the query. well, just do that! for "lab" domain, forward queries to dnsmasq.

**NOTE**: with active directory, windows clients will register their domain names from there and not from here.