#!/usr/bin/env bash

GITORG=redhat-nfvpe
CLEAN=false
FAKE=false


read -r -d '' HELPMSG << EOM
Usage:
    bash deploy.sh -r <remote_hostname_or_ip> -k <ssh_keyname> [-f <github_org_or_user>]
        -r      Remote hostname or IP address to connect as SSH proxy
        -k      SSH keyname to use. Will be stored in $HOME/.ssh/<ssh_keyname>
        -g      GitHub organization or username, e.g. https://github.com/<username>/sa-inventory. (Optional, default: redhat-nfvpe)
        -c      Clean the working directory prior to deployment.
        -f      Fake deployment. Only shows you what would have been run after cloning.
        -h      This text.
EOM

while getopts ":hr:k:g:cf" opt; do
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
        g)
            GITORG=$OPTARG
            ;;
        c)
            CLEAN=true
            ;;
        f)
            FAKE=true
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

if [ "$CLEAN" = true ]; then
    rm -rf ./working/kube-ansible/
fi
git clone https://github.com/redhat-nfvpe/kube-ansible.git ./working/kube-ansible
pushd ./working/kube-ansible/
ansible-galaxy install -r requirements.yml
#git clone https://github.com/$FORK/sa-inventory.git ./inventory/sa
git clone https://github.com/redhat-nfvpe/sa-inventory.git ./inventory/sa/
cp -r ../../../ansible/playbooks/* ./playbooks/
cp -r ../../../ansible/roles/* ./roles/
cp -r ../../../ansible/vars/ ./vars/
popd

# playbooks/vm-teardown.yml not necessary, but here for completeness
VIRT_SETUP="ansible-playbook -i inventory/sa/virthost.local \
    -e \"@./inventory/sa/vars.yml\" \
    -e \"ansible_host=$REMOTE_HOST\" \
    -e \"ssh_proxy_host=$REMOTE_HOST\" \
    -e \"vm_ssh_key_path=$HOME/.ssh/$KEYNAME\" \
    -e \"@./vars/telemetry_vars.yml\" \
    playbooks/vm-teardown.yml \
    playbooks/virthost-setup.yml \
    playbooks/cloud-monitor.yml"

CLUSTER_SETUP="ansible-playbook -i inventory/vms.local.generated \
    -e \"@./inventory/sa/vars.yml\" \
    -e \"ssh_proxy_host=$REMOTE_HOST\" \
    -e \"vm_ssh_key_path=$HOME/.ssh/$KEYNAME\" \
    -e \"@./vars/cloud_vars.yml\" \
    playbooks/cloud-monitor.yml \
    playbooks/kube-install.yml \
    playbooks/ka-gluster-install/gluster-install.yml \
    playbooks/ka-monitoring/config.yml"

TELEMETRY_SETUP="ansible-playbook -i inventory/vms.local.generated \
    -e \"ssh_proxy_host=$REMOTE_HOST\" \
    -e \"vm_ssh_key_path=$HOME/.ssh/$KEYNAME\" \
    playbooks/telemetry-monitoring/config.yml"

if [ "$FAKE" = true ]; then
    echo "--- Virtual Host Setup ---"
    echo "$VIRT_SETUP"
    echo ""
    echo "--- Cluster Setup ---"
    echo "$CLUSTER_SETUP"
    echo ""
    echo "--- Telemetry Setup ---"
    echo "$TELEMETRY_SETUP"
else
    pushd ./working/kube-ansible
    eval $VIRT_SETUP
    eval $CLUSTER_SETUP
    eval $TELEMETRY_SETUP
    popd
fi
