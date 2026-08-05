a firewall, router and dhcp/dns server

in this project, it works as both router (edge of corporation network) and absolute isolation between simulated WAN from LAN

more modern version of PfSense (better ui)

GPT 5.6 vm setting recomm:

| Setting   | Recommendation      |
| --------- | ------------------- |
| CPU       | 4 vCPUs             |
| RAM       | 2 GB (1 GB minimum) |
| Disk      | 20 GB qcow2         |
| Firmware  | UEFI (OVMF)         |
| Machine   | q35                 |
| NIC model | VirtIO              |
| Disk bus  | VirtIO              |

