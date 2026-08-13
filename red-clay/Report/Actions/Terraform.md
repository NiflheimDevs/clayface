We want to use this so that i don't have to manually define VM's on every machine and automate the vm definition across the hosts available.

I generated a quick terraform course using my beloved chatgpt and learned terraform concepts in a swift 21 lesson small chapters [[README]]

after learning, i created a simple VM on my own machine.
some problems occurred. I found a popular plugin for kvm/qemu (libvert) and used the latest version of that.
apparently this plugin is going through a heavy rewrite and big changes and is a bit unstable (0.9.x is new version and 0.8.x and before is the old version). instead of rolling to a stable release, i decided to keep going.

there were not any major problems, but one. i couldn't configure the vm to have a tty and monitor. The terraform logs literally mentioned that this failure must be a bug in the plugin.
also the documents weren't so reach so i couldn't figure out how to fix it so i left it without it. 
The only thing that i validated from this little experiment was the fact that the VM specs were created on local machine.

the next step is to do the same on multiple machine + opnsense

