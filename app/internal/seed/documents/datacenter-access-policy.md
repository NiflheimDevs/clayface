# Datacenter and hypervisor access policy (INTERNAL)

**Owner:** abed.nad
**Effective:** 2024-06-01
**Classification:** Internal

## Physical

The server room is on the second floor, behind the door with the badge reader.
Badge access is granted to IT staff and to the facilities contractor during
scheduled maintenance windows only. The keys to the rack cabinets are held by
IT; the spare set is in the lockbox in the IT office, and the lockbox code is
on the whiteboard in the same room, which is noted here so that the next
review can fix it.

## Hypervisor hosts

Both hypervisor hosts are reached over SSH from the management workstation.
The management interface is on the same L2 segment as the guests, which is
convenient and is on the list to be separated onto a management VLAN.

Rules that currently apply:

- The hypervisor hosts are administered directly over SSH, as an ordinary
  user with sudo. There is no separate management network yet.
- Guests are started and stopped from the host, not through any control plane.
  Recovery of a broken guest is done on the host console.
- Snapshots are taken before any change to a production guest. There is no
  automated snapshot schedule, so this relies on the operator remembering.

## Guests and their purpose

- The domain controller is the identity anchor. Its base image is treated as
  read-only: a VM runs an overlay, and booting a base image directly would
  corrupt everything stacked on it.
- The application host runs the customer portal and its database as containers.
  It is the only guest that publishes ports to the wider network.
- Workstations are rebuilt from the base image rather than repaired.

## Access requests

Requests go to IT, are approved by the requester's manager, and are recorded in
the ticket queue. Emergency access may be granted verbally and recorded
afterwards — a documented gap, kept because the alternative during an outage is
slower than the business will accept.

## Review

This policy is reviewed annually. The last review found two open items: the
lockbox code on the whiteboard, and the absence of a management VLAN. Both are
listed above rather than closed.
