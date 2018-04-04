# Add Smart Agent Prometheus

Components used in the automation of the Smart Agent dashboard and Prometheus
Operator instance are contained in this directory.

It's expected that you have already setup your Kubernetes environment, and are
simply interested in creating the Smart Agent Prometheus Operator instance.

Currently things are static, and mostly intended for development and demo
purposes. The hope is that this work will eventually migrate to something
closer to production quality.

> **NOTE**
>
> These files are consumed during a deployment from the `contrib/deploy.sh`
> script, resulting in them being copied into a `kube-ansibe` clone within the
> `contrib/working/kube-ansible/` directory. Not intended to be consumed
> separately.
