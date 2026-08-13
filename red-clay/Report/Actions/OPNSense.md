Ok first we install OPNSense.
the OPNSense VM will need atleast two V NICs to function. 1 act as a LAN and 2nd one act as WAN.

*fun fact: two vcpu cores was not enough and it kept stalling(evidence: dashboard didn't load)*

we set the LAN interface to vm-br0 and the WAN interface as a NAT interface managed by virtio. this way WAN can connect to internet and the devices on the LAN can do so as well.

one good thing about opnsense is that it has gui. for accessing the gui, you need to give and ip address to the vm-br0 interface.

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