data "docker_network" "proxy" {
  name = "proxy"
}

data "docker_network" "internal" {
  name = "internal"
}

resource "docker_image" "app" {
  name         = "capcom6/${var.app-name}:${var.app-version}"
  keep_locally = true
}

resource "docker_secret" "config" {
  name = "${var.app-name}-config.yml-${replace(timestamp(), ":", ".")}"
  data = var.app-config-b64

  lifecycle {
    ignore_changes        = [name]
    create_before_destroy = true
  }
}

resource "docker_service" "app" {
  name = var.app-name

  task_spec {
    container_spec {
      image = docker_image.app.name

      env = jsondecode(base64decode(var.app-env-json-b64))

      secrets {
        secret_id   = docker_secret.config.id
        secret_name = docker_secret.config.name
        file_name   = "/app/config.yml"
        file_mode   = 384
        file_uid    = 405
        file_gid    = 100
      }
    }

    networks_advanced {
      name = data.docker_network.proxy.id
    }

    networks_advanced {
      name = data.docker_network.internal.id
    }

    resources {
      limits {
        memory_bytes = var.memory-limit
      }

      reservation {
        memory_bytes = 32 * 1024 * 1024
      }
    }
  }

  # Traefik support
  labels {
    label = "traefik.enable"
    value = true
  }
  labels {
    label = "traefik.docker.network"
    value = data.docker_network.proxy.name
  }
  labels {
    label = "traefik.http.routers.${var.app-name}.rule"
    value = "Host(`${var.app-host}`) && PathPrefix(`/api`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}.tls.certresolver"
    value = "le"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-new-addprefix.addprefix.prefix"
    value = "/api"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new.middlewares"
    value = "${var.app-name}-new-addprefix"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new.rule"
    value = "Host(`api.sms-gate.app`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new.tls.certresolver"
    value = "le"
  }

  labels {
    label = "traefik.http.services.${var.app-name}.loadbalancer.server.port"
    value = 3000
  }

  # Prometheus support
  labels {
    label = "prometheus.enabled"
    value = true
  }

  rollback_config {
    order   = "start-first"
    monitor = "5s"
  }

  update_config {
    order          = "start-first"
    failure_action = "rollback"
    monitor        = "5s"
  }
}
