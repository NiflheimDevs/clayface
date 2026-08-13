I deployed two alpine VMs on two different machines on the same subnet
I used QEMU and KVM
Configured host kernels to use VXLAN for the networking of these VMs

now an app in one VM can talk to an app on another VM like they are on the same subnet even if they really weren't.

I used libvirt to manage the VMs for now.

first create 
```bash
sudo ip link add vm-br0 type bridge && sudo ip link set vm-br0 up
```

for vxlan:
```bash
sudo ip link add vxlan0 type vxlan id 100 remote [dst-ip] local [your-ip] dev wlan0 dstport [port] && sudo ip link set vxlan0 up
```

enslave vxlan0 to vm-br0
```bash
sudo ip link set vxlan0 master vm-br0
```


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