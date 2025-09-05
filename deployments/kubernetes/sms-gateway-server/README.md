# SMS Gateway Server Helm Chart

A Helm chart for deploying the SMS Gateway Server on Kubernetes.

## Installation

```bash
helm repo add android-sms-gateway https://android-sms-gateway.github.io/server
helm repo update

helm upgrade --install sms-gateway-server  \
  --namespace sms-gateway-system \
  --create-namespace \
  -f your-values.yaml \
  # optionally pin a chart version for reproducible installs:
  # --version X.Y.Z \
  android-sms-gateway/sms-gateway-server
```

## Configuration

This chart supports three configuration methods:

1. **External Secret** (Recommended for production)
2. **Unsecure ConfigMap** (Development/testing)
3. **Environment Variables** (Override specific values)

### Basic Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `ghcr.io/android-sms-gateway/server` |
| `image.tag`        | Container image tag | `""` (uses Chart.AppVersion) |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |

### Security Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext` | Pod security context | `{}` |
| `securityContext` | Container security context | `{}` |
| `imagePullSecrets` | Image pull secrets | `[]` |

### Autoscaling Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable horizontal pod autoscaler | `false` |
| `autoscaling.minReplicas` | Minimum number of replicas | `1` |
| `autoscaling.maxReplicas` | Maximum number of replicas | `100` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization | `80` |
| `autoscaling.targetMemoryUtilizationPercentage` | Target memory utilization | `null` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts[0].host` | Hostname | `chart-example.local` |
| `ingress.hosts[0].paths[0].path` | Path | `/` |
| `ingress.hosts[0].paths[0].pathType` | Path type | `ImplementationSpecific` |
| `ingress.tls` | TLS configuration | `[]` |

### Resource Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `null` |
| `resources.limits.memory` | Memory limit | `null` |
| `resources.requests.cpu` | CPU request | `null` |
| `resources.requests.memory` | Memory request | `null` |

### Node Selection

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nodeSelector` | Node selector | `{}` |
| `tolerations` | Tolerations | `[]` |
| `affinity` | Affinity rules | `{}` |

## Application Configuration

The SMS Gateway Server can be configured using three methods:

### Method 1: External Secret (Recommended)

Create an external secret containing your configuration:

```yaml
config:
  secret:
    enabled: true
    secretName: "my-sms-gateway-config"
```

Your secret should contain a `config.yml` key with the full configuration.

### Method 2: Unsecure ConfigMap

For development or non-sensitive environments:

```yaml
config:
  unsecure:
    enabled: true
    configYml: |
      gateway:
        mode: private
        private_token: "your-secret-token"
      http:
        listen: 0.0.0.0:8080
        proxies:
          - "127.0.0.1"
      database:
        dialect: mysql
        host: mysql-service
        port: 3306
        user: sms_user
        password: sms_password
        database: sms
        timezone: UTC
        max_open_conns: 4
        max_idle_conns: 2
      fcm:
        credentials_json: "{}"
        timeout_seconds: 1
        debounce_seconds: 5
      tasks:
        hashing:
          interval_seconds: 15
```

### Method 3: Environment Variables

Override specific configuration values using environment variables:

```yaml
extraEnv:
  - name: GATEWAY__MODE
    value: "public"
  - name: DATABASE__HOST
    value: "mysql.example.com"
  - name: DATABASE__PASSWORD
    valueFrom:
      secretKeyRef:
        name: mysql-secret
        key: password
```

## Configuration Reference

Based on the example configuration file, here are all available configuration options:

### Gateway Configuration

| Parameter | Environment Variable | Description | Default |
|-----------|---------------------|-------------|---------|
| `config.unsecure.configYml` (gateway.mode) | `GATEWAY__MODE` | Gateway mode: `public` (anonymous registration) or `private` (protected) | `private` |
| `config.unsecure.configYml` (gateway.private_token) | `GATEWAY__PRIVATE_TOKEN` | Access token for device registration in private mode | `123456789` |

### HTTP Server Configuration

| Parameter | Environment Variable | Description | Default |
|-----------|---------------------|-------------|---------|
| `config.unsecure.configYml` (http.listen) | `HTTP__LISTEN` | HTTP server listen address | `0.0.0.0:8080` |
| `config.unsecure.configYml` (http.proxies) | `HTTP__PROXIES` | Trusted proxy addresses | `["127.0.0.1"]` |

### Database Configuration

| Parameter | Environment Variable | Description | Default |
|-----------|---------------------|-------------|---------|
| `config.unsecure.configYml` (database.dialect) | `DATABASE__DIALECT` | Database dialect (only mysql supported) | `mysql` |
| `config.unsecure.configYml` (database.host) | `DATABASE__HOST` | Database host | `localhost` |
| `config.unsecure.configYml` (database.port) | `DATABASE__PORT` | Database port | `3306` |
| `config.unsecure.configYml` (database.user) | `DATABASE__USER` | Database user | `root` |
| `config.unsecure.configYml` (database.password) | `DATABASE__PASSWORD` | Database password | `root` |
| `config.unsecure.configYml` (database.database) | `DATABASE__DATABASE` | Database name | `sms` |
| `config.unsecure.configYml` (database.timezone) | `DATABASE__TIMEZONE` | Database timezone (important for message TTL) | `UTC` |
| `config.unsecure.configYml` (database.max_open_conns) | `DATABASE__MAX_OPEN_CONNS` | Maximum open database connections | `4` |
| `config.unsecure.configYml` (database.max_idle_conns) | `DATABASE__MAX_IDLE_CONNS` | Maximum idle database connections | `2` |

### Firebase Cloud Messaging (FCM) Configuration

| Parameter | Environment Variable | Description | Default |
|-----------|---------------------|-------------|---------|
| `config.unsecure.configYml` (fcm.credentials_json) | `FCM__CREDENTIALS_JSON` | Firebase credentials JSON (for public mode only) | `"{}"` |
| `config.unsecure.configYml` (fcm.timeout_seconds) | `FCM__TIMEOUT_SECONDS` | Push notification send timeout | `1` |
| `config.unsecure.configYml` (fcm.debounce_seconds) | `FCM__DEBOUNCE_SECONDS` | Push notification debounce (>= 5s) | `5` |

### Tasks Configuration

| Parameter | Environment Variable | Description | Default |
|-----------|---------------------|-------------|---------|
| `config.unsecure.configYml` (tasks.hashing.interval_seconds) | `TASKS__HASHING__INTERVAL_SECONDS` | Hashing interval for privacy purposes | `15` |

### HTTP API Configuration

| Parameter | Environment Variable | Description | Default |
|-----------|---------------------|-------------|---------|
| N/A | `HTTP__API__HOST` | External API host URL for the gateway | N/A |
| N/A | `HTTP__OPENAPI__ENABLED` | Enable OpenAPI documentation endpoint | `false` |
| N/A | `HTTP__API__PATH` | API base path prefix | `/` |

### Additional Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.debug` | Enable debug mode | `false` |
| `extraEnv` | Additional environment variables | `{}` |

## Health Checks

The chart includes health checks that can be customized:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Volume Configuration

Add additional volumes and volume mounts:

```yaml
volumes:
  - name: custom-volume
    configMap:
      name: my-configmap

volumeMounts:
  - name: custom-volume
    mountPath: /custom/path
    readOnly: true
```

## Examples

### Basic Deployment with External Database

```yaml
# values.yaml
config:
  unsecure:
    enabled: true
    configYml: |
      gateway:
        mode: private
        private_token: "my-secure-token"
      database:
        host: mysql.default.svc.cluster.local
        user: sms_user
        password: sms_password
        database: sms_gateway
```

> For production: run your database as a managed/external service with backups
> and durable storage. Avoid in-cluster ephemeral databases unless explicitly
> configured with PersistentVolumeClaims and a proper HA/backup strategy.

### Production Deployment with External Secret

```yaml
# values.yaml
config:
  secret:
    enabled: true
    secretName: "sms-gateway-config"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

### Environment Variable Override

```yaml
# values.yaml
config:
  unsecure:
    enabled: true
    configYml: |
      # Basic config here

# Override specific values with env vars
extraEnv:
  - name: DATABASE__HOST
    value: "production-mysql.example.com"
  - name: DATABASE__PASSWORD
    valueFrom:
      secretKeyRef:
        name: mysql-credentials
        key: password
  - name: FCM__CREDENTIALS_JSON
    valueFrom:
      secretKeyRef:
        name: fcm-credentials
        key: credentials.json
```

## Security Considerations

**Always use external secrets in production** for sensitive configuration values.

## Troubleshooting

### Common Issues

1. **Pod fails to start**: Check configuration format and required fields
2. **Database connection errors**: Verify database connectivity and credentials
3. **Health check failures**: Ensure the `/health` endpoint is accessible on port 8080

### Debug Mode

Enable debug logging:

```yaml
config:
  debug: true
```

## Additional Resources

### Example Configuration Files

For complete examples, see:

- **[examples/values.yaml](examples/values.yaml)** - Example Helm values configuration with external secret setup

- **[configs/config.example.yml](../../../configs/config.example.yml)** - Complete application configuration reference
