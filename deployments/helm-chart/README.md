# SMSGate Server Helm Chart

This Helm chart deploys the SMSGate Server to a Kubernetes cluster. The server acts as the backend component for the [SMSGate app](https://github.com/capcom6/android-sms-gateway), facilitating SMS messaging through connected Android devices.

## Prerequisites

- Kubernetes 1.23+
- Helm 3.2.0+
- PV provisioner support in the underlying infrastructure (if using persistent storage for database)

## Installation

To install the chart with the release name `my-release`:

```bash
helm repo add sms-gate https://s3.sms-gate.app/charts
helm repo update
helm install my-release sms-gate/server
```

## Uninstallation

To uninstall/delete the `my-release` deployment:

```bash
helm delete my-release
```

## Configuration

The following table lists the configurable parameters of the SMSGate chart and their default values.

| Parameter                                       | Description                           | Default                                                                             |
| ----------------------------------------------- | ------------------------------------- | ----------------------------------------------------------------------------------- |
| `replicaCount`                                  | Number of replicas                    | `1`                                                                                 |
| `image.repository`                              | Container image repository            | `ghcr.io/android-sms-gateway/server`                                                |
| `image.tag`                                     | Container image tag                   | `latest`                                                                            |
| `image.pullPolicy`                              | Container image pull policy           | `IfNotPresent`                                                                      |
| `service.type`                                  | Kubernetes service type               | `ClusterIP`                                                                         |
| `service.port`                                  | Service port                          | `3000`                                                                              |
| `ingress.enabled`                               | Enable ingress                        | `false`                                                                             |
| `ingress.className`                             | Ingress class name                    | `""`                                                                                |
| `ingress.hosts`                                 | Ingress hosts configuration           | `[{host: sms-gateway.local, paths: [{path: /, pathType: ImplementationSpecific}]}]` |
| `resources`                                     | Resource requests/limits              | `{requests: {cpu: 100m, memory: 128Mi}, limits: {cpu: 500m, memory: 512Mi}}`        |
| `autoscaling.enabled`                           | Enable autoscaling                    | `false`                                                                             |
| `autoscaling.minReplicas`                       | Minimum replicas for autoscaling      | `1`                                                                                 |
| `autoscaling.maxReplicas`                       | Maximum replicas for autoscaling      | `5`                                                                                 |
| `autoscaling.targetCPUUtilizationPercentage`    | Target CPU utilization percentage     | `80`                                                                                |
| `autoscaling.targetMemoryUtilizationPercentage` | Target memory utilization percentage  | `80`                                                                                |
| `database.host`                                 | Database host                         | `db`                                                                                |
| `database.port`                                 | Database port                         | `3306`                                                                              |
| `database.user`                                 | Database user                         | `sms`                                                                               |
| `database.password`                             | Database password                     | `""`                                                                                |
| `database.name`                                 | Database name                         | `sms`                                                                               |
| `database.deployInternal`                       | Deploy internal MariaDB               | `true`                                                                              |
| `gateway.privateToken`                          | Private token for device registration | `""`                                                                                |
| `gateway.fcmCredentials`                        | Firebase Cloud Messaging credentials  | `""`                                                                                |
| `gateway.config.enabled`                        | Enable config file mounting           | `false`                                                                             |
| `gateway.config.existingConfigMap`              | Existing ConfigMap to use             | `""`                                                                                |
| `gateway.config.data`                           | Inline YAML config content            | `""`                                                                                |
| `env`                                           | Additional environment variables      | `{}`                                                                                |

## Custom Configuration

### Using an External Database

To use an external database instead of the built-in MariaDB:

```yaml
database:
  deployInternal: false
  host: external-db-host
  port: 3306
  user: external-user
  password: "secure-password"
  name: external-db
```

### Configuring Ingress

To enable ingress with TLS:

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: sms-gateway.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - hosts:
        - sms-gateway.example.com
      secretName: sms-gateway-tls
```

### Setting Resource Limits

To adjust resource requests and limits:

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi
```

### Configuring Config File Mounting

#### Using an Existing ConfigMap

To use an existing ConfigMap for configuration:

```yaml
gateway:
  config:
    enabled: true
    existingConfigMap: my-existing-configmap
```

#### Using Inline YAML Configuration

To provide configuration using inline YAML:

```yaml
gateway:
  config:
    enabled: true
    data: |
      tasks:
        hashing:
          interval_seconds: 900
```

## Notes

- The application health endpoint is available at `/health`
- When using private mode, you must set `gateway.privateToken`
- For production use, always set secure passwords and enable TLS
- The chart supports persistent storage for the internal MariaDB database
- The config file mounting feature is optional and can be enabled by setting `gateway.config.enabled` to `true`
- When using config file mounting, you can either specify an existing ConfigMap with `gateway.config.existingConfigMap` or provide inline YAML configuration with `gateway.config.data`
- Environment variables remain an alternative configuration method and can be set using the `env` parameter

## License

This Helm chart is licensed under the Apache-2.0 license. See [LICENSE](LICENSE) for more information.

## Legal Notice

Android is a trademark of Google LLC.