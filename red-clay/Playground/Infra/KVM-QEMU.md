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