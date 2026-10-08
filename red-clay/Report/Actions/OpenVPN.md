So yeah.
A real system would read VPN users from AD. but due to time restrictions, we are skipping that part and using local users.
the local users and the actual setup of VPN server is happening all on OPNSense itself.
I chose OpenVPN over OpenConnect over simplicity and faster setup time. At the moment im developing this, time is very short.

OpenVPN server is a simple community plugin on OPNSense.

The OpenVPN does need a server side certificate. So i defined a CA on OPNSense.
and a simple local user called `vpnuser`.

when the client wants to connect, it will need to verify the certificate. i don't have a way to handle that :). So i wrote the fingerprint in deploy.env.example. the openvpn client now can say that i trust the certificate with this fingerprint.

I used nc to send data between my machine. fairly easy.
the openvpn command to connect is something like this:
```bash
sudo openvpn --client --auth-user-pass --peer-fingerprint D0:1D:0E:93:A1:B3:07:F8:EB:5B:9F:02:AB:F5:98:BC:93:42:B0:34:8E:4B:35:FC:0A:76:91:61:8B:BC:98:9A --remote heimdall.clayface 1194 --dev tun
```
from OpenVpn server side configuration, I had to configure the OpenVPN to push default gatway
needed to add a firewall rule to pass any packet from OpenVPN subnet. one funny thing to know is that openvpn created it's own interface without me doing it for it.

and like that, you can connect to the subnet!