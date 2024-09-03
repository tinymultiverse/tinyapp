# Getting Started

Getting started with Tiny App is easy.

## Prerequisites
- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)
file (default location is `~/.kube/config`).

## Install Tiny App Components

Clone this repository and fill in value for APP_INGRESS_DOMAIN environment variable in manifests/install.yaml.

```bash
kubectl create namespace tinyapp
kubectl apply -n tinyapp -f manifests/install.yaml
```

#### Prometheus

If you don't already have Prometheus set up for your cluster, check out the
[Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator).

By default, app containers expose metrics at port '9090' and path '/metrics'. To customize, you can set
GATEWAY_METRICS_PORT & GATEWAY_METRICS_PATH environment variables for tinyapp-controller deployment.

To enable TLS, set GATEWAY_METRICS_TLS_ENABLED and TLS_SECRET_NAME environment variables for tinyapp-controller
deployment.

#### Notes
- To configure TLS for app ingress, set APP_INGRESS_TLS_ENABLED env var for tinyapp-server and TLS_SECRET_NAME for
tinyapp-controller.

## Deploy Tiny App Instance

#### 1. Deploy Tiny App from JupyterLab extension
