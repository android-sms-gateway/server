resource "docker_service" "worker" {
  name = "${var.app-name}-worker"

  task_spec {
    container_spec {
      image   = docker_image.app.name
      command = ["/app/app", "worker"]

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

  #region Prometheus support
  labels {
    label = "prometheus.io/scrape"
    value = "true"
  }
  labels {
    label = "prometheus.io/port"
    value = "3000"
  }
  labels {
    label = "prometheus.io/job"
    value = "worker"
  }
  #endregion

  rollback_config {
    order   = "stop-first"
    monitor = "5s"
  }

  update_config {
    order          = "stop-first"
    failure_action = "rollback"
    monitor        = "5s"
  }
}
