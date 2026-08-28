#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
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

step "Ansible: start VMs created by terraform (start_vms.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/start_vms.yml" 

step "Done. Infrastructure provisioned."
