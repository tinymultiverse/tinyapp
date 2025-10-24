# TinyApp Helm Chart

This Helm chart deploys the TinyApp controller and server components on a Kubernetes cluster.

## Components

This chart deploys:

- **Namespace**: `tinyapp` namespace for all resources
- **CustomResourceDefinition**: TinyApp CRD for defining TinyApp resources
- **ServiceAccounts**: Separate service accounts for controller and server
- **RBAC**: Roles and RoleBindings for controller and server permissions
- **Controller Deployment**: Manages TinyApp custom resources
- **Server Deployment**: Provides the TinyApp server API
- **Service**: Exposes the TinyApp server

## Installation

To install the chart with the release name `tinyapp`:

```bash
helm install tinyapp ./helm/tinyapp
```

## Upgrading

To upgrade the release:

```bash
helm upgrade tinyapp ./helm/tinyapp
```

## Uninstallation

To uninstall/delete the `tinyapp` release:

```bash
helm uninstall tinyapp
```

**Note**: This will not delete the CRD or namespace by default. To remove them:

```bash
kubectl delete crd tinyapps.tinymultiverse.ai
kubectl delete namespace tinyapp
```

## Configuration

The following table lists the configurable parameters that can be overridden:

| Parameter                             | Description                    | Default                                 |
| ------------------------------------- | ------------------------------ | --------------------------------------- |
| `namespace`                           | Namespace for all resources    | `tinyapp`                               |
| `server.appIngressDomain`             | Domain for app ingress         | `host.docker.internal`                  |
| `server.appIngressTlsEnabled`         | Enable TLS for app ingress     | `"false"`                               |
| `controller.tlsSecretName`            | TLS secret name for controller | `""` (empty)                            |
| `controller.metricsEnabled`           | Enable metrics for gateway     | `"false"`                               |
| `controller.gatewayMetricsPort`       | Gateway metrics port           | `"9090"`                                |
| `controller.gatewayMetricsPath`       | Gateway metrics path           | `/metrics`                              |
| `controller.gatewayMetricsTlsEnabled` | Enable TLS for gateway metrics | `"false"`                               |
| `controller.image`                    | Controller Docker image        | `ghcr.io/tinymultiverse/tinyapp:latest` |
| `server.image`                        | Server Docker image            | `ghcr.io/tinymultiverse/tinyapp:latest` |
| `controller.replicaCount`             | Number of controller replicas  | `1`                                     |
| `server.replicaCount`                 | Number of server replicas      | `1`                                     |
| `rbac.create`                         | Create RBAC resources          | `true`                                  |

### Overriding Values

#### Using `--set` flag:

```bash
helm install tinyapp ./helm/tinyapp \
  --set server.appIngressDomain=myapp.example.com \
  --set server.replicaCount=2
```

#### With TLS enabled:

```bash
helm install tinyapp ./helm/tinyapp \
  --set server.appIngressDomain=myapp.example.com \
  --set server.appIngressTlsEnabled=true \
  --set controller.tlsSecretName=my-tls-secret
```

#### With metrics enabled:

```bash
helm install tinyapp ./helm/tinyapp \
  --set controller.metricsEnabled=true \
  --set controller.gatewayMetricsPort=9090 \
  --set controller.gatewayMetricsPath=/metrics
```

#### With metrics and TLS enabled:

```bash
helm install tinyapp ./helm/tinyapp \
  --set controller.metricsEnabled=true \
  --set controller.gatewayMetricsTlsEnabled=true \
  --set controller.tlsSecretName=my-tls-secret
```

#### Using a custom values file:

Create a file `my-values.yaml`:

```yaml
namespace: tinyapp

server:
  appIngressDomain: myapp.example.com
  replicaCount: 2

controller:
  replicaCount: 2
```

Then install with:

```bash
helm install tinyapp ./helm/tinyapp -f my-values.yaml
```

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Resource Requirements

Default resource allocations:

**Controller**:

- Requests: 50m CPU, 64Mi memory
- Limits: 100m CPU, 1Gi memory

**Server**:

- Requests: 50m CPU, 64Mi memory
- Limits: 100m CPU, 1Gi memory

You can override these in your custom values file.
