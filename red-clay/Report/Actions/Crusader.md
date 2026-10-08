Preparing the Kali image was rather easy since it was linux. 
I decided to not ip pin the crusader just for good measures. maybe a little bit lookahead for maybe someone wanting to add blue teaming to this lab. they get random ip so you can't tell which is which!
So, the Crusader VM is connected to OPNsense via the WAN interface.
because i wanted the crusader to have access to internet, i had to configure OPNSense as a router for Crusader to forward it's request to outside world from a FOURTH interface which is NAT and connected to the default vnet of libvirt.
No worries though. all 10.0.0.0/8 are blocked from WAN with some exceptions for public service and VPN.

So OPNsense now has become a very important peace and does multiple jobs for different sides. It will be hard to make sure new functionalities and rules won't break the integrity of the lab.
Whatever :-)

I first created Crusader image with cloud init for some network configuration and vpn CA configuration but that worked weirdly and i didn't feel good about it.
So i am moving to playbook instead of cloud init.
I don't think it's better, i just think im more comfortable this way. 

didn't add host ip route table entry for the WAN because im doing my best to not leave any possible hole in the project (no cheating!).