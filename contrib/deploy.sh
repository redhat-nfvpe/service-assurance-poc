#!/usr/bin/env bash
mkdir -p ~/src/sa-development
cd ~/src/sa-development
git clone https://github.com/redhat-nfvpe/kube-ansible
cd kube-ansible
git checkout develop
ansible-galaxy install -r requirements.yml
git clone https://github.com/redhat-nfvpe/sa-inventory ./inventory/sa

# playbooks/vm-teardown.yml not necessary, but here for completeness
ansible-playbook -i inventory/sa/virthost.local \
    -e "@./inventory/sa/vars.yml" \
    playbooks/vm-teardown.yml \
    playbooks/virthost-setup.yml

ansible-playbook -i inventory/vms.local.generated \
    -e “@./inventory/sa/vars.yml” \
    playbooks/kube-install.yml \
    playbooks/ka-gluster-install/gluster-install.yml \
    playbooks/ka-monitoring/config.yml
