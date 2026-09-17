
Before doing anything, we need to define exactly what this project is and what it is going to do so that we can choose our technology and path correctly.

What we will try to do in this project is creating a **dynamic environment** which acts as a playground for an attacker to attack and sharpen his skills and try new things.

There are several challenges we need to address before going into project:
## 1. Real Simulation
In order for us to create a real environment for an attacker to attack, we have to design and mock a real world corporation network. 
A real world corporation network consists of many elements such as:
1. VPN server
2. Active directory
3. Web Servers and Databases 
4. etc
Some of these elements should be fully isolated from each other and some should be highly integrated.
This demands for spinning up multiple VMs and them being able to talk to each other.
Handling life cycles and deployment of these VMs is the challenge here.

Side note: multiple VMs can be very expensive to handle (consumes a lot of resources). In order to make it more friendly, we have to support spinning VMs in multiple machines. This introduces a networking challenge.
## 2. Dynamic Environment
In order for us to achieve our goal of *dynamic environment*, we need to make the containers and apps in our VM highly configurable.
For example we need to allow the deploy-er to be able to choose the MangoDB version and also config the configurations.


Now that our challenges are clear, we need to choose the best fitting tools to tackle 'em.
[[Tools]]

### Actions:
1. [[Bare Metal VM]]
2. [[Report/Actions/OPNSense|OPNSense]]
3. [[Terraform]]
4. [[Ansible]]

Now that we have done basic Ansible playbooks, we come into an important dilemma.
### What to do with shared knowledge between terraform and ansible?

Creating a single source of truth is proving to be very right now difficult right now.
Actually more ugly than difficult.

**Hypervisor's ip addresses are considered FACTS**. they don't belong to either terraform or ansible. a general yaml could possibly hold the values for us.

VM creation is the job of terraform. terraform needs to know **where to create the VM** (what host). **Host Addresses** and **Map of VM To Host**.
There is an idea of having a *scipt deciding where each vm should go using the spec of every available host/hypervisor*. This will require the script to know **Each Host Specs** and **Each VM Required Specs**. we will out scope this for now.

Creating network configuration on each host is the job of Ansible. For that, it needs to know **Host Addresses**, **Which Host is Holding OPNSense (Who is Edge) OR Map of VM to Host**

Starting VM's is the job of Ansible. For that it will need to know **Host Addresses**, **Map of VM To Host**

### Another problem: More than two hosts?

More than two hosts will create a need to L2 forwarding network. it doesn't look that hard but im scared.


For the first problem, we could use a source of knowledge that terraform and ansible read from. (or asible reads from terrafrom while terraform reads from that).

currently implemented. 
the implementation raised an important issue and that was akward configurations that looked bad. 
for those to get fixed, we may need to generate some terraform configs with a script! (**TODO**)

For the second problem, We can have a mesh or L2 forwarding on one of the hosts and that being connected to everyone (like a brain)

### Update: project knowledge base + naming

The knowledge kept growing but had no home. We now separate two kinds of knowledge:

1. **Facts** — hypervisor IPs, VM placements, network settings. these live in `lab.yaml` and terraform/ansible read it directly. unchanged.
2. **Intent** — what the project is, the target topology, architecture decisions and their reasons. this now lives in a knowledge base: `docs/redclay.yaml`. nothing reads it automatically; it is for humans and agents. if it disagrees with `lab.yaml` or the code, lab.yaml/code wins.

Writing it down forced some floating decisions to become real:

- the "MangoDB version" idea from the dynamic-environment section → resolved as [[PostgreSQL]] in a container on APP01, with versions and configuration controllable. SQL Server stays optional for later (windows-integrated auth attack paths that postgres doesn't reproduce naturally).
- applications run as containers on APP01, not one VM per service.
- a modern identity provider (IDP01: OIDC/OAuth2/SAML/SSO) is part of the core, next to AD.
- the scenario / programmable attack graph idea is explicitly **deferred**. core infrastructure and AD first, scenario engine later.

Naming: the project is now called **[[Clayface]]**. the knowledge base originally called it RedClay; it has been renamed to match.

Current phase: [[Attacker/Active Directory]] — DC01, CLIENT01, users/groups/permissions, Kerberos and LDAP, then Linux AD integration, IDP01, and the first attack paths.

### Actions:
1. [[AD]]
2. 