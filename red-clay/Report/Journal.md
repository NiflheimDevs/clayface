
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

Actions:
1. [[Bare Metal VM]]
2. 
