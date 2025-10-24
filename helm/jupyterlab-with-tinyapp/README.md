# JupyterLab with TinyApp Helm Chart

This Helm chart deploys JupyterLab integrated with TinyApp on a Kubernetes cluster.

## Installation

To install the chart with the release name `my-jupyterlab`:

```bash
helm install my-jupyterlab ./helm/jupyterlab-with-tinyapp
```

## Upgrading

To upgrade the release:

```bash
helm upgrade my-jupyterlab ./helm/jupyterlab-with-tinyapp
```

## Uninstallation

To uninstall/delete the `my-jupyterlab` release:

```bash
helm uninstall my-jupyterlab
```

## Configuration

### Overriding Values

You can override these values in several ways:

#### 1. Using `--set` flag:

```bash
helm install my-jupyterlab ./helm/jupyterlab-with-tinyapp \
  --set jupyter.appPreviewUrl=http://myapp.example.com \
  --set jupyter.aiEnabled=true \
  --set ingress.host=myapp.example.com
```

#### 2. Using a custom values file:

Create a file `my-values.yaml`:

```yaml
jupyter:
  appPreviewUrl: http://myapp.example.com
  aiEnabled: "true"
  tinyAppImage: ghcr.io/tinymultiverse/jupyterlab-tinyapp:v1.2.3

ingress:
  host: myapp.example.com
```

Then install with:

```bash
helm install my-jupyterlab ./helm/jupyterlab-with-tinyapp -f my-values.yaml
```

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- A PersistentVolumeClaim named `my-data` (or override `volume.claimName`)

## Components

This chart deploys:

- **Deployment**: Contains two containers:
  - `jupyter`: JupyterLab with TinyApp integration
  - `gateway`: TinyApp gateway for routing
- **Service**: Exposes the deployment on port 80
- **Ingress**: Routes external traffic to the service (can be disabled)
