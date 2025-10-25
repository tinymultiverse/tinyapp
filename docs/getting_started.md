# Getting Started

Getting started with Tiny App is easy.

## Prerequisites
- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)
file (default location is `~/.kube/config`).
- Installed [helm](https://helm.sh/docs/intro/install/).
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

### Configure Domain (If running on local machine)

If your cluster is running on local machine (Docker Desktop k8s, Minikube etc.), add the following entry to /etc/hosts (Linux or Mac) or C:\Windows\System32\drivers\etc\hosts (Windows):

```bash
127.0.0.1  tinymultiverse.local
```

## Install Tiny App Components

If you want to enable TLS and/or metrics, read further before executing:

```bash
helm install tinyapp ./helm/tinyapp --namespace tinyapp --create-namespace --set server.appIngressDomain=tinymultiverse.local
```

If you have a different domain set up for your cluster, you should use that instead of tinymultiverse.local.

#### With TLS

To enable TLS for TinyApp URLs, add the following options:

```bash
--set server.appIngressTlsEnabled=true \
--set controller.tlsSecretName=<SECRET_NAME>
```

TLS secret should exist in the same namespace as TinyApp controller.

#### Metrics/Prometheus

If you don't already have Prometheus set up for your cluster, check out the
[Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator).

By default, app containers expose metrics at port '9090' and path '/metrics'. To customize, you can override:

```bash
--set controller.metricsEnabled=true \
--set controller.gatewayMetricsPort=9090 \
--set controller.gatewayMetricsPath=/metrics \
--set controller.gatewayMetricsTlsEnabled=true \
--set controller.tlsSecretName=<SECRET_NAME>
```

## Deploy TinyApp

#### Using JupyterLab Extension

Create a Persistent Volume Claim if you don't have one configured already.

```bash
kubectl apply -f manifests/pvc.yaml
```

If you get an error, you may need to update storageClassName in manifests/pvc.yaml.

```bash
kubectl get storageclass
```

Start a JupyterLab container by running:

```bash
helm install my-jupyterlab ./helm/jupyterlab-with-tinyapp \
  --set jupyter.appPreviewUrl=http://tinymultiverse.local \
  --set jupyter.aiEnabled=false \
  --set ingress.host=tinymultiverse.local
```

You should now be able to access jupyterlab at http://tinymultiverse.local/jupyter.

Refer to the [extension user guide](https://github.com/tinymultiverse/jupyterlab-tinyapp/blob/main/docs/USER_GUIDE.md) for how to preview app, view logs, deploy app, etc.