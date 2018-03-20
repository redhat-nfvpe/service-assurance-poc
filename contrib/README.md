# Automated Deployment

Deployment of virtual infrastructure on a single baremetal node is done with
[kube-ansible](https://github.com/redhat-nfvpe/kube-ansible), which also
leverages [Prometheus Operator](https://github.com/coreos/prometheus-operator).

In this directory we provide a simple `deploy.sh` script that makes it easier
to setup a development environment without memorizing Ansible commands.

## Prerequisites

You have a baremetal node running CentOS 7.4, accessible via SSH as the root
user with SSH keys (no password).

## Deployment

Run the following command. Substitute the IP address and SSH key name for your
own.

    bash deploy.sh -r 192.168.100.100 -k virthost-lab

Your deployment will be run out of `~/src/sa-development`.
