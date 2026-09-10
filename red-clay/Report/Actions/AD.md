I chose windows server 2025 as base image.
installed it and added the role AD DS to it for domain service.

since the AD must be reproducible (e.g. on next deployment, everything should be reset). we needed to have an immutable base image. Every vm that gets created uses that image as base and doesn't WRITE to it, just reads. the writes go to separate disk.

I created the base image until the point of adding just the role.
In order for the vm to boot everywhere and correctly, i needed to run *sysprep.exe*. this program brings the vm to a state before booting WITHOUT deleting anything.

So when you boot, it will be like you are booting for the first time (random hostname, some sid and etc) BUT the domain controller/service is still present.

I did that and i ran into a problem.
I wanted to do the promotion to ad, user creation and etc with ansible. but if the newly booted vm is in the state of *just booting for the first time*, i can reach it!. ansible needs WinRM process to be running on the vm in order to connect and run scripts.

so we tricked the vm into thinking this is not the first time you are getting booted so that other processes start normally. with an xml file doing some things. unattend.xml was the file's name.

So now that everything was set, still ansible couldn't work! the problem rose from reachability!
eveything is dhcp in this lab
with no knowledge of what the vm ip is, how ansible is supposed to connect to it? the answer that i had for this was predetermined hostname and dns from opnsense. but windows server generates hostname randomly so i don't know the domain name!

so i had to bind the mac to a static ip (for dc).
this way i know the ip and dns. mac is known on vm creation and outputted by terraform.

BUT THERE ARE PROBLEMS WITH SYSPREP. it somehow destroys things and nothing works

so i did it again and again and again until it worked
that's it! it worked
i didn't have time to learn ad api and ansible plugin for it so it was written by ai. i did tell it the flow and it implemented. can't wait for it to bite my ass later.
