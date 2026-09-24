#!/bin/bash
set -euo pipefail

# Lab secrets live in deploy.env. Create it from
# deploy.env.example: the OPNsense API key/secret come from the OPNsense UI
# (System -> Access -> Users).
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/deploy.env" ]; then
    set -a
    . "$SCRIPT_DIR/deploy.env"
    set +a
else
    echo "!! $SCRIPT_DIR/deploy.env not found." >&2
    echo "   Copy deploy.env.example to deploy.env and fill in the OPNsense" >&2
    echo "   API credentials (OPNsense UI -> System -> Access -> Users)." >&2
    exit 1
fi

export no_proxy="clayface,10.0.0.1,localhost,127.0.0.1${no_proxy:+,$no_proxy}"
export NO_PROXY="$no_proxy"

TERRAFORM_DIR="$SCRIPT_DIR/terraform"
ANSIBLE_DIR="$SCRIPT_DIR/ansible"

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Provision infrastructure end-to-end."
    echo ""
    echo "Options:"
    echo "  --destroy     Teardown infrastructure instead of deploying"
    echo "  --plan        Show what terraform will do without applying"
    echo "  --skip-ansible Only run terraform"
    echo "  -h, --help    Show this help"
    exit 0
}

DESTROY=false
PLAN_ONLY=false
SKIP_ANSIBLE=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --destroy)    DESTROY=true; shift ;;
        --plan)       PLAN_ONLY=true; shift ;;
        --skip-ansible) SKIP_ANSIBLE=true; shift ;;
        -h|--help)    usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

step() {
    echo ""
    echo "==> $1"
}

fail() {
    echo "!! $1" >&2
    exit 1
}

# ---------- terraform ----------

step "Terraform init"
terraform -chdir="$TERRAFORM_DIR" init -input=false -upgrade

if [ "$DESTROY" = true ]; then
    step "Terraform destroy"
    terraform -chdir="$TERRAFORM_DIR" destroy -auto-approve -input=false
    step "Done. Infrastructure destroyed."
    exit 0
fi

if [ "$PLAN_ONLY" = true ]; then
    step "Terraform plan"
    terraform -chdir="$TERRAFORM_DIR" plan -input=false
    step "Plan complete. Review above."
    exit 0
fi

step "Terraform apply"
terraform -chdir="$TERRAFORM_DIR" apply -auto-approve -input=false

# ---------- ansible ----------

if [ "$SKIP_ANSIBLE" = true ]; then
    step "Done. Skipping ansible."
    exit 0
fi

step "Ansible: configure hypervisors (hosts.yml)"
ansible-playbook -i "$ANSIBLE_DIR/inventory/lab_inventory.py" "$ANSIBLE_DIR/playbooks/hosts.yml" --ask-become-pass

step "Ansible: configure edge (edge.yml)"
ansible-playbook -i "$ANSIBLE_DIR/inventory/lab_inventory.py" "$ANSIBLE_DIR/playbooks/edge.yml" --ask-become-pass

step "Ansible: start the edge VM (start_vms.yml --tags gateway)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    --tags gateway \
    "$ANSIBLE_DIR/playbooks/start_vms.yml"

# Windows-only: pins every VM that has a terraform MAC and an `ip:` in
# lab.yaml as a DHCP reservation + DNS record in OPNsense, via its REST API,
# so <name>.clayface resolves from the guest's first boot. Needs
# OPNSENSE_API_KEY / OPNSENSE_API_SECRET (from deploy.env; one-time API key:
# OPNsense UI -> System -> Access -> Users).
#
# Runs BEFORE the remaining VMs boot, deliberately: a guest that boots first
# takes a pool lease, and <name>.clayface then resolves to its reserved
# address, which nothing answers on until that lease renews.
step "Ansible: pin VMs with a static address in OPNsense DHCP/DNS (opnsense.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/opnsense.yml"

step "Ansible: start the remaining VMs (start_vms.yml --tags members)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    --tags members \
    "$ANSIBLE_DIR/playbooks/start_vms.yml"

# Windows-only: WinRM, no become pass needed. Idempotent — skips everything
# already done (hostname set, password enforced, domain installed).
step "Ansible: promote domain controller(s) (dc.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/dc.yml"

# Builds the identity layer inside the forest: OUs, groups, users, the domain
# account policy, the LAPS schema extension and its OU permissions, and the
# IDP01 read-only delegation. Needs LAB_USER_PASS (optional; defaults).
step "Ansible: build the AD identity layer (ad.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/ad.yml"

# Before client.yml: the GPOs must be linked to OU=Workstations before a
# workstation is moved into it, or the first policy refresh finds an empty OU
# and the machine comes up unmanaged until the next cycle.
step "Ansible: build the workstation GPOs (ad_gpo.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/ad_gpo.yml"

step "Ansible: join workstation(s) to the domain (client.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/client.yml"

# Builds the application tier: the container images are built here on the
# control node, shipped to app01 as a docker-save tarball, loaded and brought
# up with docker compose. Needs a running docker daemon on this machine, and
# LAB_APP_SSH_PASS / LAB_APP_DB_PASS / LAB_SVC_APP_PASS (optional; defaults).
# Runs after client.yml only because it is independent of the domain — app01
# is not domain-joined, it carries a service account as a credential.
step "Ansible: deploy the application tier (app.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/app.yml"

# Read-only. Runs each planted weakness over HTTPS against the running stack
# and asserts it works, keyed to app.weaknesses in lab.yaml; a toggle that
# is off is asserted to fail instead. No lab mutation, no restarts.
step "Ansible: validate the application tier (app_validate.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/app_validate.yml"

# Read-only. Asserts the identity layer is what docs/ad-identity-design.md
# says it is, including the negative cases — a workstation that did not
# receive policy, a tier-0 account that can administer one, a delegation
# that is broader than its two OUs. Runs last: it asserts workstation state,
# which only exists after client.yml.
step "Ansible: validate the AD identity layer (ad_validate.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/ad_validate.yml"

step "Done. Infrastructure provisioned."
