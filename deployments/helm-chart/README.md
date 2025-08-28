# SMS Gateway for Android™ Server Helm Chart

This Helm chart deploys the SMS Gateway for Android™ Server to a Kubernetes cluster. The server acts as the backend component for the [SMS Gateway for Android](https://github.com/capcom6/android-sms-gateway), facilitating SMS messaging through connected Android devices.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- PV provisioner support in the underlying infrastructure (if using persistent storage for database)

## Installation

To install the chart with the release name `my-release`:

```bash
helm install my-release ./deployments/helm-chart
```

## Uninstallation

To uninstall/delete the `my-release` deployment:

```bash
helm delete my-release
```

## Configuration

The following table lists the configurable parameters of the SMS Gateway chart and their default values.

| Parameter                                       | Description                           | Default                                                                             |
| ----------------------------------------------- | ------------------------------------- | ----------------------------------------------------------------------------------- |
| `replicaCount`                                  | Number of replicas                    | `1`                                                                                 |
| `image.repository`                              | Container image repository            | `capcom6/sms-gateway`                                                               |
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
| `gateway.mode`                                  | Gateway mode (`public` or `private`)  | `private`                                                                           |
| `gateway.privateToken`                          | Private token for device registration | `""`                                                                                |
| `gateway.fcmKey`                                | Firebase Cloud Messaging key          | `""`                                                                                |
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

### Setting Gateway Mode

To run in public mode:

```yaml
gateway:
  mode: "public"
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

## Notes

- The application health endpoint is available at `/health`
- When using private mode, you must set `gateway.privateToken`
- For production use, always set secure passwords and enable TLS
- The chart supports persistent storage for the internal MariaDB database

## License

This Helm chart is licensed under the Apache-2.0 license. See [LICENSE](LICENSE) for more information.

## Legal Notice

Android is a trademark of Google LLC.