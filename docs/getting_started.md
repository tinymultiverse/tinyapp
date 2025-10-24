# Getting Started

Getting started with Tiny App is easy.

## Prerequisites
- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)
file (default location is `~/.kube/config`).
- Default ingress controller running in your cluster (ingressclass with annotation ingressclass.kubernetes.io/is-default-class=true).

#### Install HAProxy Ingress Controller (Optional)

If you don't already have an ingress controller running in the cluster, you can install HAProxy by running:

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts

helm repo update

kubectl create namespace haproxy-controller

helm install haproxy-ingress haproxytech/kubernetes-ingress \
  --namespace haproxy-controller \
  --set controller.service.type=LoadBalancer \
  --set controller.ingressClass=haproxy \
  --set controller.config.reload-strategy=reusesocket

kubectl annotate ingressclass haproxy ingressclass.kubernetes.io/is-default-class=true
```

## Install Tiny App Components

Clone this repository and fill in value for APP_INGRESS_DOMAIN environment variable in manifests/install.yaml. If you're running Docker Desktop Kubernetes, you can set it to "host.docker.internal".

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

## Deploy TinyApp

#### Using JupyterLab Extension

Start a JupyterLab container by running:

```bash
kubectl apply -f manifests/pvc.yaml
kubectl apply -f manifests/jupyterlab.yaml
```

This assumes you're running Docker Desktop Kubernetes and exposes JupyterLab container at http://host.docker.internal.

Refer to the [extension user guide](https://github.com/tinymultiverse/jupyterlab-tinyapp/blob/main/docs/USER_GUIDE.md) for how to preview app, view logs, deploy app, etc.