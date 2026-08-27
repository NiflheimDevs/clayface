## QEMU

Emulates hardware.

Can emulate:

* x86 on ARM
* ARM on x86
* old systems
* embedded systems

Pure emulation is slow.

---

## KVM

Kernel virtualization acceleration.

Turns Linux kernel into a hypervisor.

QEMU + KVM together:

* QEMU provides virtual hardware
* KVM provides near-native execution

---


*IMPORTANT*: KVM is a type 1 hypervisor even though it is within linux kernel(a host guest exists). It uses kernel features and hardware features to behave exactly like a type 1 hypervisor.