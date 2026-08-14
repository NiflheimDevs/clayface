# Ansible Course --- Enterprise Digital Twin Automation

This course teaches Ansible as a configuration-management tool, then applies it to a reproducible Enterprise Digital Twin. The practical environment uses Arch Linux control hosts, KVM/QEMU with libvirt, Terraform-created virtual machines, and Alpine or Arch Linux managed nodes. The underlying Ansible concepts also apply to Debian/Ubuntu, RHEL-derived systems, network devices, and Windows.

Terraform and Ansible are complementary: Terraform creates and connects infrastructure; Ansible configures the operating systems and services inside it. Each lesson explains where that boundary matters.

## Fast Track

If your immediate goal is configuring Terraform/libvirt-created Digital Twin VMs, follow the focused [20-Lesson Fast Track](Fast-Track-Digital-Twin.md) before using the full reference curriculum.

## Course Path

### Part 1 --- Fundamentals

1. [What Is Ansible?](Lesson-01-What-Is-Ansible.md)
2. [Ansible Architecture](Lesson-02-Ansible-Architecture.md)
3. [Installing Ansible](Lesson-03-Installing-Ansible.md)
4. [Inventory](Lesson-04-Inventory.md)
5. [Ad-Hoc Commands](Lesson-05-Ad-Hoc-Commands.md)
6. [Modules](Lesson-06-Modules.md)
7. [Playbooks](Lesson-07-Playbooks.md)
8. [Plays and Tasks](Lesson-08-Plays-and-Tasks.md)
9. [Idempotency](Lesson-09-Idempotency.md)

### Part 2 --- Core Ansible

10. [Variables](Lesson-10-Variables.md)
11. [Facts](Lesson-11-Facts.md)
12. [Registering Command Results](Lesson-12-Registering-Command-Results.md)
13. [Conditionals](Lesson-13-Conditionals.md)
14. [Loops](Lesson-14-Loops.md)
15. [Handlers](Lesson-15-Handlers.md)
16. [Tags](Lesson-16-Tags.md)
17. [Error Handling](Lesson-17-Error-Handling.md)
18. [Check Mode and Diff Mode](Lesson-18-Check-Mode-and-Diff-Mode.md)
19. [Ansible Configuration](Lesson-19-Ansible-Configuration.md)

### Part 3 --- Managing Linux

20. [Package Management](Lesson-20-Package-Management.md)
21. [Users and Groups](Lesson-21-Users-and-Groups.md)
22. [SSH Configuration](Lesson-22-SSH-Configuration.md)
23. [Files and Directories](Lesson-23-Files-and-Directories.md)
24. [Permissions and Ownership](Lesson-24-Permissions-and-Ownership.md)
25. [Services](Lesson-25-Services.md)
26. [System Configuration](Lesson-26-System-Configuration.md)
27. [Environment Variables](Lesson-27-Environment-Variables.md)
28. [Deploying Configuration Files](Lesson-28-Deploying-Configuration-Files.md)
29. [Templates and Jinja2](Lesson-29-Templates-and-Jinja2.md)
30. [Firewall Configuration](Lesson-30-Firewall-Configuration.md)

### Part 4 --- Organizing Ansible

31. [Project Structure](Lesson-31-Project-Structure.md)
32. [Roles](Lesson-32-Roles.md)
33. [Role Variables and Defaults](Lesson-33-Role-Variables-and-Defaults.md)
34. [Role Dependencies](Lesson-34-Role-Dependencies.md)
35. [Collections](Lesson-35-Collections.md)
36. [Reusable Roles](Lesson-36-Reusable-Roles.md)
37. [Inventory Organization](Lesson-37-Inventory-Organization.md)
38. [Group Variables and Host Variables](Lesson-38-Group-Variables-and-Host-Variables.md)

### Part 5 --- Secrets and Security

39. [Ansible Vault](Lesson-39-Ansible-Vault.md)
40. [Managing SSH Credentials](Lesson-40-Managing-SSH-Credentials.md)
41. [Secrets in Variables](Lesson-41-Secrets-in-Variables.md)
42. [Avoiding Secret Leakage](Lesson-42-Avoiding-Secret-Leakage.md)
43. [Security Best Practices](Lesson-43-Security-Best-Practices.md)

### Part 6 --- Advanced Ansible

44. [Delegation](Lesson-44-Delegation.md)
45. [run_once](Lesson-45-Run-Once.md)
46. [become](Lesson-46-Become.md)
47. [Privilege Escalation](Lesson-47-Privilege-Escalation.md)
48. [Blocks](Lesson-48-Blocks.md)
49. [Includes and Imports](Lesson-49-Includes-and-Imports.md)
50. [Dynamic and Static Behavior](Lesson-50-Dynamic-and-Static-Behavior.md)
51. [Advanced Jinja2](Lesson-51-Advanced-Jinja2.md)
52. [Custom Facts](Lesson-52-Custom-Facts.md)
53. [Dynamic Inventories](Lesson-53-Dynamic-Inventories.md)
54. [Advanced Collections](Lesson-54-Advanced-Collections.md)
55. [Debugging Ansible](Lesson-55-Debugging-Ansible.md)

### Part 7 --- Terraform + Ansible

56. [Terraform vs Ansible](Lesson-56-Terraform-vs-Ansible.md)
57. [Infrastructure vs Configuration Management](Lesson-57-Infrastructure-vs-Configuration-Management.md)
58. [Terraform Outputs as Ansible Inputs](Lesson-58-Terraform-Outputs-as-Ansible-Inputs.md)
59. [Generating Inventories from Terraform](Lesson-59-Generating-Inventories-from-Terraform.md)
60. [Managing Terraform-Created VMs](Lesson-60-Managing-Terraform-Created-VMs.md)
61. [Separate Creation and Configuration Stages](Lesson-61-Separate-Creation-and-Configuration-Stages.md)
62. [Repository Design](Lesson-62-Repository-Design.md)

### Part 8 --- Enterprise Digital Twin

63. [Digital Twin Ansible Architecture](Lesson-63-Digital-Twin-Ansible-Architecture.md)
64. [Base Linux Machines](Lesson-64-Base-Linux-Machines.md)
65. [Reusable Server Roles](Lesson-65-Reusable-Server-Roles.md)
66. [Web Servers](Lesson-66-Web-Servers.md)
67. [Database Servers](Lesson-67-Database-Servers.md)
68. [Internal Services](Lesson-68-Internal-Services.md)
69. [Users and SSH](Lesson-69-Users-and-SSH.md)
70. [Network and Service Configuration](Lesson-70-Network-and-Service-Configuration.md)
71. [Preparing for Active Directory](Lesson-71-Preparing-for-Active-Directory.md)
72. [Windows Automation](Lesson-72-Windows-Automation.md)
73. [Complete Lab Deployment Workflow](Lesson-73-Complete-Lab-Deployment-Workflow.md)
74. [Rebuilds and Reproducibility](Lesson-74-Rebuilds-and-Reproducibility.md)
75. [Capstone: Enterprise-Scale Digital Twin](Lesson-75-Capstone-Project.md)

## How To Use It

Read and perform the lessons in order. Start with one disposable Linux VM and keep all course work in Git. Do not apply a playbook to a valuable host until you understand its target inventory, privileges, and expected changes. Expand the fictional inventory gradually; no early lesson requires every machine in the final lab.

Use `next lesson` to continue, `exercise` for additional practice, `solution` for the current exercise solution, `explain this` to pause progression for a concept, and `quiz` to test the material covered so far.
