# Terraform Course --- Portable Homelab Automation

This course teaches Terraform as a general Infrastructure-as-Code tool, then applies each concept to a virtualized homelab. It uses QEMU/KVM with libvirt as the first concrete implementation because that is the current lab, but the core lessons deliberately remain applicable to Proxmox, VMware, and cloud providers.

## Course Path

1. [What Is Terraform?](Lesson-01-What-Is-Terraform.md)
2. [Terraform Architecture](Lesson-02-Terraform-Architecture.md)
3. [HCL Basics](Lesson-03-HCL-Basics.md)
4. [Providers](Lesson-04-Providers.md)
5. [Resources](Lesson-05-Resources.md)
6. [Terraform State](Lesson-06-State.md)
7. [Input Variables](Lesson-07-Input-Variables.md)
8. [Outputs](Lesson-08-Outputs.md)
9. [Locals](Lesson-09-Locals.md)
10. [Expressions and Iteration](Lesson-10-Expressions-and-Iteration.md)
11. [Functions and Console](Lesson-11-Functions-and-Console.md)
12. [Dependency Graph](Lesson-12-Dependency-Graph.md)
13. [Lifecycle and Change Safety](Lesson-13-Lifecycle-and-Change-Safety.md)
14. [Modules](Lesson-14-Modules.md)
15. [Remote State and Backends](Lesson-15-Remote-State-and-Backends.md)
16. [Provisioners and Bootstrapping](Lesson-16-Provisioners-and-Bootstrapping.md)
17. [Project Design and Best Practices](Lesson-17-Project-Design-and-Best-Practices.md)
18. [Virtualization Providers and Lab Concepts](Lesson-18-Virtualization-Providers-and-Lab-Concepts.md)
19. [Building a Multi-Host Lab](Lesson-19-Building-a-Multi-Host-Lab.md)
20. [Ansible Integration](Lesson-20-Ansible-Integration.md)
21. [Capstone: A Portable Multi-Host Lab](Lesson-21-Capstone-Project.md)

## How To Use It

Read and do the exercises in order. Keep all early experiments disposable, use Git, and do not run `apply` or `destroy` against irreplaceable lab resources until you understand the state and plan. Each platform-specific section is an example, not a requirement: substitute the provider and module implementation after verifying its current documentation.
