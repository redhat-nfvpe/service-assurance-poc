#!/usr/bin/env bash

FORK=redhat-nfvpe

read -r -d '' HELPMSG << EOM
Usage:
    bash deploy.sh -r <remote_hostname_or_ip> -k <ssh_keyname> [-f <github_org_or_user>]
        -r      Remote hostname or IP address to connect as SSH proxy
        -k      SSH keyname to use. Will be stored in $HOME/.ssh/<ssh_keyname>
        -f      GitHub organization or username, e.g. https://github.com/<username>/sa-inventory. (Optional, default: redhat-nfvpe)
        -h      This text.
EOM

while getopts ":hr:k:f:" opt; do
    case $opt in
        h)
            echo -e "$HELPMSG" >&2
            exit 0
            ;;
        r)
            REMOTE_HOST=$OPTARG
            ;;
        k)
            KEYNAME=$OPTARG
            ;;
        f)
            FORK=$OPTARG
            ;;
        /?)
            echo "Invalid option: -$OPTARG" >&2
            exit 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument" >&2
            exit 1
            ;;
    esac
done

if [ -z "$REMOTE_HOST" ]; then
    echo "Remote host is required. Use -r."
    exit 1
fi

if [ -z "$KEYNAME" ]; then
    echo "SSH key name is required. Use -k."
    exit 1
fi

# TODO: clean this up so we more gracefully handle existing directories
mkdir -p ~/src/sa-development
cd ~/src/sa-development
git clone https://github.com/redhat-nfvpe/kube-ansible
cd kube-ansible
git checkout develop
ansible-galaxy install -r requirements.yml
git clone https://github.com/$FORK/sa-inventory ./inventory/sa

# playbooks/vm-teardown.yml not necessary, but here for completeness
ansible-playbook -i inventory/sa/virthost.local \
    -e "@./inventory/sa/vars.yml" \
    -e "ansible_host=$REMOTE_HOST" \
    -e "ssh_proxy_host=$REMOTE_HOST" \
    -e "vm_ssh_key_path=$HOME/.ssh/$KEYNAME" \
    playbooks/vm-teardown.yml \
    playbooks/virthost-setup.yml

ansible-playbook -i inventory/vms.local.generated \
    -e "@./inventory/sa/vars.yml" \
    -e "ssh_proxy_host=$REMOTE_HOST" \
    -e "vm_ssh_key_path=$HOME/.ssh/$KEYNAME" \
    playbooks/kube-install.yml \
    playbooks/ka-gluster-install/gluster-install.yml \
    playbooks/ka-monitoring/config.yml
