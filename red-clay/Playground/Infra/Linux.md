# LXC
???
does it concern me? 
seems like a fun concept

# Virtual CPU and etc
wrote it somewhere something about it (in KVM i think). didn't go deep enough on this. blacked boxed it.
# Network

## FLOW

in OSI levels

### Application
boring
### Transport
boring

### IP

in here,  there is a routing table  accessed by ```
``` bash
ip route
```
this command shows IP to Physical Layer. 
so where is Link Layer?
### Link Layer

in IP, we now know from routing table that where will the packet should go. but first we gotta do Link Layer stuff. the computer looks at another table accessed by
``` bash
ip neigh
```
this is the infamous *ARP* table. it shows IP to MAC and if it is reachable or not (expiration?).
either we have a reachable IP -> MAC or we don't. if we do then yay! if not, we need to get one! do an ARP request. but to where (by who)? THE WHERE WAS DISCOVERED IN THE PREVIOUS LAYER!
### Physical Layer
apparently some wifi drivers have some limitation. in 802.11 standard protocol, it is expected to each frame getting to NIC (getting sent) to have the **SAME** MAC address. 
so you can't have multiple src MACs and send them to your NIC like its nothing. 
## Bridge

### What is it?
practically a switch. it receives data and decides where to send them next.
bridge has MAC.
can have ip. for host identity.(a bit complicated and hard to understand)
### How does it work?
act exactly like a switch. forwards between slaves through mac address resolves.

### How does it forward?
for a switch to forward, you need ports
ports are created with `master` keyword
for example, 
``` bash
ip link set vnet0 master br0
```

## VXLAN

emulates LAN over IP! they think they are on the same LAN while they are not.

```bash

sudo ip link add vxlan0 type vxlan id 100 remote 192.168.1.22 local 192.168.1.11 dev wlan0 dstport 4789

```
any frame going into this VXLAN, will be wrapped in IP packet with dest ip 192.168.1.22 and src ip 192.168.1.11 with dest port 4789 and header ID 100
```
+----------------------+
| Outer Ethernet       |
+----------------------+
| Outer IP             |
+----------------------+
| Outer UDP            |
+----------------------+
| VXLAN Header         |
+----------------------+
| Original Ethernet    |
| Frame                |
+----------------------+
```

the same VXLAN should exist at destination to unwrap the packet on the same port with same ID. 
VXLANs can share the same port.