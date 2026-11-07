data "docker_network" "proxy" {
  name = "proxy"
}

data "docker_network" "internal" {
  name = "internal"
}

resource "docker_image" "app" {
  name         = "ghcr.io/android-sms-gateway/server:${var.app-version}"
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
        value = "backend"
      }
      #endregion
    }

    networks_advanced {
      name = data.docker_network.proxy.id
    }

    networks_advanced {
      name = data.docker_network.internal.id
    }

    placement {
      constraints = ["node.labels.traefik!=true"]
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

  #region Rate limit Middlewares
  labels {
    label = "traefik.http.middlewares.rate-limit_5-per-1m.ratelimit.average"
    value = "5"
  }

  labels {
    label = "traefik.http.middlewares.rate-limit_5-per-1m.ratelimit.period"
    value = "1m"
  }

  labels {
    label = "traefik.http.middlewares.rate-limit_5-per-1m.ratelimit.sourcecriterion.ipstrategy.depth"
    value = "1"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-3rdparty-ratelimit.ratelimit.average"
    value = "10"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-3rdparty-ratelimit.ratelimit.period"
    value = "1s"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-3rdparty-ratelimit.ratelimit.burst"
    value = "30"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-3rdparty-ratelimit.ratelimit.sourcecriterion.ipstrategy.depth"
    value = "1"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-3rdparty-inflight.inflightreq.amount"
    value = "20"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-3rdparty-inflight.inflightreq.sourcecriterion.ipstrategy.depth"
    value = "1"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-upstream-ratelimit.ratelimit.average"
    value = "5"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-upstream-ratelimit.ratelimit.period"
    value = "1m"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-upstream-ratelimit.ratelimit.sourcecriterion.ipstrategy.depth"
    value = "1"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-mobile-ratelimit.ratelimit.average"
    value = "5"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-mobile-ratelimit.ratelimit.period"
    value = "1s"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-mobile-ratelimit.ratelimit.burst"
    value = "100"
  }

  labels {
    label = "traefik.http.middlewares.${var.app-name}-mobile-ratelimit.ratelimit.sourcecriterion.ipstrategy.depth"
    value = "1"
  }
  #endregion

  #region Add Prefix Middleware
  labels {
    label = "traefik.http.middlewares.${var.app-name}-new-addprefix.addprefix.prefix"
    value = "/api"
  }
  #endregion

  #region Deprecated
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
    value = "cloudflare"
  }
  #endregion

  #region Deprecated Limited
  labels {
    label = "traefik.http.routers.${var.app-name}_limited.rule"
    value = "Host(`${var.app-host}`) && PathPrefix(`/api/mobile/v1/device`) && Method(`POST`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_limited.middlewares"
    value = "rate-limit_5-per-1m"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_limited.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_limited.tls.certresolver"
    value = "cloudflare"
  }
  #endregion

  #region Deprecated Rate-limited
  labels {
    label = "traefik.http.routers.${var.app-name}_3rdparty.rule"
    value = "Host(`${var.app-host}`) && PathPrefix(`/api/3rdparty`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_3rdparty.middlewares"
    value = "${var.app-name}-3rdparty-inflight,${var.app-name}-3rdparty-ratelimit"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_3rdparty.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_3rdparty.tls.certresolver"
    value = "cloudflare"
  }

  labels {
    label = "traefik.http.routers.${var.app-name}_upstream.rule"
    value = "Host(`${var.app-host}`) && PathPrefix(`/api/upstream`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_upstream.middlewares"
    value = "${var.app-name}-upstream-ratelimit"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_upstream.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_upstream.tls.certresolver"
    value = "cloudflare"
  }

  labels {
    label = "traefik.http.routers.${var.app-name}_mobile.rule"
    value = "Host(`${var.app-host}`) && PathPrefix(`/api/mobile`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_mobile.middlewares"
    value = "${var.app-name}-mobile-ratelimit"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_mobile.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_mobile.tls.certresolver"
    value = "cloudflare"
  }

  labels {
    label = "traefik.http.routers.${var.app-name}_limited-user.rule"
    value = "Host(`${var.app-host}`) && PathPrefix(`/api/mobile/v1/user`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_limited-user.middlewares"
    value = "rate-limit_5-per-1m"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_limited-user.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}_limited-user.tls.certresolver"
    value = "cloudflare"
  }
  #endregion

  #region Primary
  labels {
    label = "traefik.http.routers.${var.app-name}-new.rule"
    value = "Host(`api.sms-gate.app`)"
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
    label = "traefik.http.routers.${var.app-name}-new.tls.certresolver"
    value = "cloudflare"
  }

  labels {
    label = "traefik.http.routers.${var.app-name}-wellknown.rule"
    value = "Host(`api.sms-gate.app`) && PathPrefix(`/.well-known`) && Method(`GET`,`HEAD`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-wellknown.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-wellknown.tls.certresolver"
    value = "cloudflare"
  }
  #endregion

  #region Primary Rate-limited
  labels {
    label = "traefik.http.routers.${var.app-name}-new-3rdparty.rule"
    value = "Host(`api.sms-gate.app`) && PathPrefix(`/3rdparty`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-3rdparty.middlewares"
    value = "${var.app-name}-3rdparty-inflight,${var.app-name}-3rdparty-ratelimit,${var.app-name}-new-addprefix"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-3rdparty.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-3rdparty.tls.certresolver"
    value = "cloudflare"
  }

  labels {
    label = "traefik.http.routers.${var.app-name}-new-upstream.rule"
    value = "Host(`api.sms-gate.app`) && PathPrefix(`/upstream`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-upstream.middlewares"
    value = "${var.app-name}-upstream-ratelimit,${var.app-name}-new-addprefix"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-upstream.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-upstream.tls.certresolver"
    value = "cloudflare"
  }

  labels {
    label = "traefik.http.routers.${var.app-name}-new-mobile.rule"
    value = "Host(`api.sms-gate.app`) && PathPrefix(`/mobile`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-mobile.middlewares"
    value = "${var.app-name}-mobile-ratelimit,${var.app-name}-new-addprefix"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-mobile.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new-mobile.tls.certresolver"
    value = "cloudflare"
  }
  #endregion

  #region Primary Limited
  #region Device Registration
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited.rule"
    value = "Host(`api.sms-gate.app`) && PathPrefix(`/mobile/v1/device`) && Method(`POST`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited.middlewares"
    value = "rate-limit_5-per-1m,${var.app-name}-new-addprefix"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited.tls.certresolver"
    value = "cloudflare"
  }
  #endregion
  #region User Operations
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited-user.rule"
    value = "Host(`api.sms-gate.app`) && PathPrefix(`/mobile/v1/user`)"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited-user.entrypoints"
    value = "https"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited-user.middlewares"
    value = "rate-limit_5-per-1m,${var.app-name}-new-addprefix"
  }
  labels {
    label = "traefik.http.routers.${var.app-name}-new_limited-user.tls.certresolver"
    value = "cloudflare"
  }
  #endregion
  #endregion

  labels {
    label = "traefik.http.services.${var.app-name}.loadbalancer.server.port"
    value = 3000
  }

  #region Prometheus support
  labels {
    label = "prometheus.enabled"
    value = true
  }

  labels {
    label = "prometheus.io/scrape"
    value = true
  }
  labels {
    label = "prometheus.io/port"
    value = 3000
  }
  labels {
    label = "prometheus.io/job"
    value = "backend"
  }
  #endregion

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
