I deployed two alpine VMs on two different machines on the same subnet
I used QEMU and KVM
Configured host kernels to use VXLAN for the networking of these VMs

now an app in one VM can talk to an app on another VM like they are on the same subnet even if they really weren't.

I used libvirt to manage the VMs for now.

first create
```bash
sudo ip link add vm-lan0 type bridge && sudo ip link set vm-lan0 up
```

there is more than one segment now, so one bridge each — the Lan segment, the
DMZ segment, and the NAT leg for the WAN:
```bash
sudo ip link add vm-dmz0 type bridge && sudo ip link set vm-dmz0 up
sudo ip link add vm-wan0 type bridge && sudo ip link set vm-wan0 up
```

the DMZ bridge carries the control node's address too, which is how the
playbooks reach anything in the DMZ:
```bash
sudo ip addr add 10.0.10.2/24 dev vm-dmz0
```

for vxlan, one per segment — the id is what pairs a VXLAN with a segment:
```bash
sudo ip link add vxlan100 type vxlan id 100 remote [dst-ip] local [your-ip] dev wlan0 dstport [port] && sudo ip link set vxlan100 up
sudo ip link add vxlan101 type vxlan id 101 remote [dst-ip] local [your-ip] dev wlan0 dstport [port] && sudo ip link set vxlan101 up
```

enslave each vxlan to its bridge — the id names the segment, so vxlan100 goes
to the Lan bridge and vxlan101 to the DMZ one:
```bash
sudo ip link set vxlan100 master vm-lan0
sudo ip link set vxlan101 master vm-dmz0
```

the WAN bridge gets no VXLAN: it is a NAT leg that never leaves the host.


make sure firewalld allows vxlan port:
```bash
sudo firewall-cmd --add-port=[port]udp [--permanent]
sudo firewall-cmd --reload
```

inside each vm, up the main link and set an ip addr like this:
```bash
sudo ip addr add 10.0.0.1/24 dev eth0
```
wallah! it should now work!