# Networking

QEMU creates:
1. one virtual NIC with the name vnet0. also called TAP in linux.
2.  one virtual [bridge](Linux.md#Bridge) named virbr0.

## NAT
in NAT, guest kernel sends its packets to the vnet0 like a good boy. from there, virbr0 carries it. 
it sets the dest MAC to host MAC and delivers to host (switch)
host receives it and brings it to IP
in IP, libvert has done a configuration called MASQUERADE. with a rule that anything that comes from the vnet0 to change their ip to host ip and send it to network like a host packet.

but how does a packet reach guest kernel? well, there is a tracker called conntrack. this conntrack tracks network flows and can reverse NAT. for TCP UDP, uses ports to translate. other protocols like ICMP have unique packet ids and conntrack uses them to track them.

## Bridge

there is no IP translations here.
a packet reaches vnet0 as if it is just a normal kernel connected to somewhere.
in virt-manager, it needs a bridge to create the VM. because in this mode we don't use virbr0 and we are going to use our own bridge with our own rules and routes. we call our bridge br0.

we have two choices:
1. connect br0 to host NIC directly
2. we wrap our packets in IP packets

option 1 does work. with limitations. 
- if we are using wifi, there are driver limitations. [[Linux#Physical Layer]] 
- if we are using eth0, everything works. guest gets it's own ip from gateway router. the possible problem can be exactly That! getting an IP means different IPs and reaching the VM will force us to know the ip.

option 2 is perfect for our use case. it adds networking overhead in both packet sizes and computation but worth it.
how do we do it? you might ask
we have something called [[Linux#VXLAN | VXLAN]]
