This thing has many things. 
# AD DS
It practically provides
1. Authentication
2. Access Control

It holds domain joined machines, users/groups and services.
For each thing it holds you can define access policies to machines and services.

*Important*: everyone mostly find everything via DNS. So it must be the DNS server of relevant components.
example of requests:
```
_ldap._tcp.dc._msdcs.clayface.local
_kerberos._tcp.clayface.local
```
## Flow
AD DS has public and protected APIs. To talk to protected API, you need to have a secure channel with it or some how authenticated.
To get it, you need to *Join the Domain*.
To join the domain, you send a request to AD DS and introduce yourself + a domain account credentials. AD DS then authenticates the account and lets you join (it creates entry in it self + DNS record (not related but still does it)).

Now we can talk to AD DS for important stuff.

Now for example, someone wants to login to the machine as a domain user. Now the fun starts. 
The computer uses the authenticated channel to ask AD if this domain account exists? is password correct and yada yada. After all that, if everything was ok, the AD returns a TGT(Ticket Granting Tickets). 
Know user logs in. everytime user wants to use a service, it uses the TGT to make a req to KDC (Key Distribution Center, sub system of AD) to get a service specific ticket.

This is called **Kerberos** service authenticaion.

In this flow, you need to give the ticket with your request to a service to get access.
There is another flow that you don't have the responsibility to get a ticket, the service does.

The protocol is called **LDAP** (Lightweight Directory Access Protocol). LDAP is used to query users and groups and policies and etc in AD. 
This is a Domain Controller. It practically provides
1. Authentication
2. Access Control

It holds domain joined machines, users/groups and services.
For each thing it holds you can define access policies to machines and services.

*Important*: it mostly finds everything via DNS. So it must be the DNS server of relevant components.

## Flow
AD DS has public and protected APIs. To talk to protected API, you need to have a secure channel with it.
To get it, you need to *Join the Domain*.
To join the domain, you send a request to AD DS and introduce yourself + a domain account credentials. AD DS then authenticates the account and lets you join (it creates entry in it self + DNS record).

Now we can talk to AD DS for important stuff.

Now for example, someone wants to login to the machine as a domain user. Now the fun starts. 
The computer uses the authenticated channel to ask AD if this domain account exists? is password correct and yada yada. After all that, if everything was ok, the AD returns a TGT(Ticket Granting Tickets). 
Know user logs in. everytime user wants to use a service, it uses the TGT to make a req to KDC (Key Distribution Center, sub system of AD) to get a service specific ticket.

This is called **Kerberos** service authenticaion.
```
Alice
 │
 │ credentials
 ▼
CLIENT01
 │
 │ Kerberos AS-REQ
 ▼
KDC on DC01
 │
 │ authenticate Alice
 │
 │ AS-REP
 ▼
CLIENT01
 │
 └── TGT
```

In this flow, you need to give the ticket with your request to a service to get access.
There is another flow that you don't have the responsibility to get a ticket, the service does.

The protocol is called **LDAP** (Lightweight Directory Access Protocol). LDAP is used to query users and groups and policies and etc in AD. 