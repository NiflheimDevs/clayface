
# NIC Types

| NIC                     | Speed | CPU Usage | Driver Needed | Best For                 |
| ----------------------- | ----- | --------- | ------------- | ------------------------ |
| VirtIO                  | ★★★★★ | Very Low  | Yes           | Almost everything        |
| VirtIO Non-transitional | ★★★★★ | Very Low  | Yes           | Modern Linux             |
| e1000                   | ★★★   | Medium    | No            | Compatibility            |
| e1000e                  | ★★★   | Medium    | No            | Compatibility            |
| vmxnet3                 | ★★★★  | Low       | Yes           | VMware migrations        |
| igb                     | ★★★   | Medium    | No            | Intel compatibility      |
| rtl8139                 | ★     | High      | No            | Legacy systems           |
| pcnet                   | ★     | High      | No            | Very old OSes            |
| ne2k                    | ☆     | Very High | No            | DOS                      |
| lance                   | ☆     | Very High | No            | Historical compatibility |

| Guest OS                            | Recommended NIC                                  |
| ----------------------------------- | ------------------------------------------------ |
| Modern Linux                        | VirtIO (or VirtIO Non-Transitional if supported) |
| OPNsense                            | VirtIO                                           |
| FreeBSD 13+                         | VirtIO                                           |
| Windows 10/11                       | VirtIO (install VirtIO drivers if needed)        |
| Windows installation before drivers | e1000 or e1000e temporarily                      |
| Windows XP                          | e1000                                            |
| Windows 98                          | rtl8139 or pcnet                                 |
| DOS                                 | ne2k                                             |
# Storage Types

| Bus             | Speed | CPU        | Drivers               | Best Use                                 |
| --------------- | ----- | ---------- | --------------------- | ---------------------------------------- |
| VirtIO (blk)    | ★★★★★ | Very Low   | Yes                   | General-purpose modern VMs               |
| VirtIO SCSI     | ★★★★★ | Very Low   | Yes                   | Many disks, hotplug, enterprise features |
| NVMe            | ★★★★☆ | Low-Medium | Native in modern OSes | NVMe testing, realistic hardware         |
| SATA            | ★★★   | Medium     | No                    | Compatibility                            |
| SCSI (emulated) | ★★★   | Medium     | No                    | Legacy enterprise compatibility          |
| IDE             | ★     | High       | No                    | Very old operating systems               |
| USB             | ★     | High       | No                    | Removable media testing                  |

| Guest                             | Recommended                     |
| --------------------------------- | ------------------------------- |
| OPNsense                          | VirtIO (blk) or VirtIO SCSI     |
| Alpine Linux                      | VirtIO (blk)                    |
| Ubuntu/Debian                     | VirtIO (blk)                    |
| Fedora/RHEL                       | VirtIO (blk)                    |
| Windows 11                        | VirtIO (install VirtIO drivers) |
| Windows installer without drivers | SATA                            |
| Windows XP                        | IDE or SATA                     |
| DOS                               | IDE                             |
