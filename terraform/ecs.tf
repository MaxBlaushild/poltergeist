module "ecs" {
  source = "terraform-aws-modules/ecs/aws"

  cluster_name = local.name


  services = {
    sonar_core = {
      cpu = 1024
      memory = 2048

      task_exec_secret_arns = [
        aws_secretsmanager_secret.db_password.arn,
        aws_secretsmanager_secret.auth_private_key.arn,
        aws_secretsmanager_secret.open_ai_key.arn,
        aws_secretsmanager_secret.twilio_account_sid.arn,
        aws_secretsmanager_secret.twilio_auth_token.arn,
        aws_secretsmanager_secret.imagine_api_key.arn,
        aws_secretsmanager_secret.use_api_key.arn,
        aws_secretsmanager_secret.google_maps_api_key.arn,
        aws_secretsmanager_secret.mapbox_api_key.arn,
        aws_secretsmanager_secret.google_drive_client_id.arn,
        aws_secretsmanager_secret.google_drive_client_secret.arn,
        aws_secretsmanager_secret.dropbox_client_id.arn,
        aws_secretsmanager_secret.dropbox_client_secret.arn,
        aws_secretsmanager_secret.hue_bridge_hostname.arn,
        aws_secretsmanager_secret.hue_bridge_username.arn,
        aws_secretsmanager_secret.hue_client_id.arn,
        aws_secretsmanager_secret.hue_client_secret.arn,
        aws_secretsmanager_secret.hue_application_key.arn,
        aws_secretsmanager_secret.travel_angels_stripe_secret_key.arn,
        aws_secretsmanager_secret.ethereum_private_key.arn,
        aws_secretsmanager_secret.ca_private_key.arn,
        aws_secretsmanager_secret.polymarket_api_key.arn,
        aws_secretsmanager_secret.polymarket_api_secret.arn,
        aws_secretsmanager_secret.polymarket_api_passphrase.arn,
        aws_secretsmanager_secret.polymarket_address.arn,
      ]

      tasks_iam_role_statements = [
        {
          sid       = "ReefSiteArtifactsObjectAccess"
          effect    = "Allow"
          actions   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
          resources = ["arn:aws:s3:::reef-site-artifacts/*"]
        },
        {
          sid       = "ReefSiteArtifactsBucketAccess"
          effect    = "Allow"
          actions   = ["s3:ListBucket"]
          resources = ["arn:aws:s3:::reef-site-artifacts"]
        }
      ]

      container_definitions = {
        "core" = {
          essential = true
          image     = "${aws_ecr_repository.core.repository_url}:latest"
          # readonlyRootFilesystem defaults to true in this module. reef-site's
          # OpenSCAD preview path (go/reef-site/internal/server/configure.go)
          # shells out and needs a writable scratch dir for each render
          # (go/pkg/reef/procexec.NewWorkDir, under os.TempDir()) — give it an
          # in-memory tmpfs at /tmp rather than disabling read-only-root
          # wholesale.
          linuxParameters = {
            tmpfs = [
              {
                containerPath = "/tmp"
                size          = 512
              }
            ]
          }
          environment = [
            {
              name  = "DB_HOST"
              value = aws_db_instance.poltergeist-db.address
            },
            {
              name  = "DB_USER"
              value = aws_db_instance.poltergeist-db.username
            },
            {
              name  = "DB_PORT"
              value = tostring(aws_db_instance.poltergeist-db.port)
            },
            {
              name  = "DB_NAME"
              value = aws_db_instance.poltergeist-db.db_name
            },
            {
              name  = "PHONE_NUMBER"
              value = var.twilio_phone_number
            },
            {
              name  = "REDIS_URL"
              value = "redis://${aws_elasticache_cluster.redis.cache_nodes[0].address}:6379"
            },
            {
              name  = "HUE_REDIRECT_URI"
              value = "https://api.poltergeist.gg/final-fete/hue-oauth/callback"
            },
            {
              name  = "GOOGLE_DRIVE_REDIRECT_URI"
              value = "https://api.poltergeist.gg/travel-angels/google-drive/callback"
            },
            {
              name  = "DROPBOX_REDIRECT_URI"
              value = "https://api.poltergeist.gg/travel-angels/dropbox/callback"
            },
            {
              name  = "BASE_URL"
              value = "https://api.poltergeist.gg"
            },
            {
              name = "ETHEREUM_TRANSACTOR_URL"
              value = "http://localhost:8088"
            },
            {
              name = "C2PA_CONTRACT_ADDRESS"
              value = "0x653d604fdaA2320DF90cc4e9dFd5aabe86BD91A9"
            },
            {
              name  = "GOOGLE_APPLICATION_CREDENTIALS"
              value = "/etc/core/vera-firebase.json"
            },
            {
              name  = "UNCLAIMED_STREETS_APPLICATION_CREDENTIALS"
              value = "/etc/core/fcm-service-account.json"
            },
            {
              name  = "REEF_S3_BUCKET"
              value = "reef-site-artifacts"
            },
            {
              name  = "REEF_AWS_REGION"
              value = local.region
            },
            {
              # Default (210mm) was narrower than what the schema itself
              # advertises as the widest configurable rack (widthMm max
              # 250mm) — any width over 210 always failed the print-envelope
              # check regardless of the actual printer. 250 matches the
              # printer's real 256x256x256mm bed with a small safety margin.
              name  = "REEF_MAX_BBOX_MM"
              value = "250"
            },
            {
              name  = "REEF_OPERATOR_EMAIL"
              value = "orders@reef.forteus.tech"
            },
            {
              name  = "EMAIL_FROM_ADDRESS"
              value = "no-reply@reef.forteus.tech"
            }
          ]
          secrets = [
            {
              name      = "DB_PASSWORD"
              valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
            },
            {
              name      = "TWILIO_ACCOUNT_SID"
              valueFrom = "${aws_secretsmanager_secret.twilio_account_sid.arn}"
            },
            {
              name      = "TWILIO_AUTH_TOKEN"
              valueFrom = "${aws_secretsmanager_secret.twilio_auth_token.arn}"
            },
            {
              name      = "HUE_CLIENT_ID"
              valueFrom = "${aws_secretsmanager_secret.hue_client_id.arn}"
            },
            {
              name      = "HUE_CLIENT_SECRET"
              valueFrom = "${aws_secretsmanager_secret.hue_client_secret.arn}"
            },
            {
              name      = "HUE_APPLICATION_KEY"
              valueFrom = "${aws_secretsmanager_secret.hue_application_key.arn}"
            },
            {
              name      = "HUE_BRIDGE_HOSTNAME"
              valueFrom = "${aws_secretsmanager_secret.hue_bridge_hostname.arn}"
            },
            {
              name      = "HUE_BRIDGE_USERNAME"
              valueFrom = "${aws_secretsmanager_secret.hue_bridge_username.arn}"
            },
            {
              name      = "IMAGINE_API_KEY"
              valueFrom = "${aws_secretsmanager_secret.imagine_api_key.arn}"
            },
            {
              name      = "USE_API_KEY"
              valueFrom = "${aws_secretsmanager_secret.use_api_key.arn}"
            },
            {
              name      = "MAPBOX_API_KEY"
              valueFrom = "${aws_secretsmanager_secret.mapbox_api_key.arn}"
            },
            {
              name      = "GOOGLE_MAPS_API_KEY"
              valueFrom = "${aws_secretsmanager_secret.google_maps_api_key.arn}"
            },
            {
              name      = "GOOGLE_DRIVE_CLIENT_ID"
              valueFrom = "${aws_secretsmanager_secret.google_drive_client_id.arn}"
            },
            {
              name      = "GOOGLE_DRIVE_CLIENT_SECRET"
              valueFrom = "${aws_secretsmanager_secret.google_drive_client_secret.arn}"
            },
            {
              name      = "DROPBOX_CLIENT_ID"
              valueFrom = "${aws_secretsmanager_secret.dropbox_client_id.arn}"
            },
            {
              name      = "DROPBOX_CLIENT_SECRET"
              valueFrom = "${aws_secretsmanager_secret.dropbox_client_secret.arn}"
            },
            {
              name      = "CA_PRIVATE_KEY"
              valueFrom = "${aws_secretsmanager_secret.ca_private_key.arn}"
            }
          ]
          portMappings = [
            {
              name          = local.core_container_name
              containerPort = local.core_container_port
              hostPort      = local.core_container_port
              protocol      = "tcp"
            }
          ]
        }

        "authenticator" = {
          essential = true
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "AUTH_PRIVATE_KEY",
            valueFrom = "${aws_secretsmanager_secret.auth_private_key.arn}"
          }]
          image = "${aws_ecr_repository.authenticator.repository_url}:latest"
          portMappings = [
            {
              name          = "authenticator"
              containerPort = 8089
              hostPort      = 8089
              protocol      = "tcp"
            }
          ]
        }

        "fount-of-erebos" = {
          # cpu       = 256
          # memory    = 512
          essential = true
          secrets = [{
            name      = "OPEN_AI_KEY",
            valueFrom = "${aws_secretsmanager_secret.open_ai_key.arn}"
          }]
          image = "${aws_ecr_repository.fount_of_erebos.repository_url}:latest"
          portMappings = [
            {
              name          = "fount-of-erebos"
              containerPort = 8081
              hostPort      = 8081
              protocol      = "tcp"
            }
          ]
        }

        "texter" = {
          essential = true
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "TWILIO_ACCOUNT_SID",
            valueFrom = "${aws_secretsmanager_secret.twilio_account_sid.arn}"
          }, {
            name = "TWILIO_AUTH_TOKEN",
            valueFrom = "${aws_secretsmanager_secret.twilio_auth_token.arn}"
          }]
          image = "${aws_ecr_repository.texter.repository_url}:latest"
          portMappings = [
            {
              name          = "texter"
              containerPort = 8084
              hostPort      = 8084
              protocol      = "tcp"
            }
          ]
        }

        "job-runner" = {
          essential = true
          # Same readonlyRootFilesystem-vs-writable-/tmp issue as the "core"
          # container above — job-runner's full generate+slice path
          # (go/job-runner/internal/processors/generate_reef_configuration.go)
          # shells out to OpenSCAD and PrusaSlicer via the same
          # procexec.NewWorkDir(os.TempDir()) helper.
          linuxParameters = {
            tmpfs = [
              {
                containerPath = "/tmp"
                size          = 1024
              }
            ]
          }
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "IMAGINE_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.imagine_api_key.arn}"
          }, {
            name      = "USE_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.use_api_key.arn}"
          }, {
            name      = "GOOGLE_MAPS_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.google_maps_api_key.arn}"
          }, {
            name      = "POLYMARKET_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.polymarket_api_key.arn}"
          }, {
            name      = "POLYMARKET_API_SECRET",
            valueFrom = "${aws_secretsmanager_secret.polymarket_api_secret.arn}"
          }, {
            name      = "POLYMARKET_API_PASSPHRASE",
            valueFrom = "${aws_secretsmanager_secret.polymarket_api_passphrase.arn}"
          }, {
            name      = "POLYMARKET_ADDRESS",
            valueFrom = "${aws_secretsmanager_secret.polymarket_address.arn}"
          }]
          image = "${aws_ecr_repository.job_runner.repository_url}:latest"
          environment = [
            {
              name  = "CHAIN_ID"
              value = "84532"
            },
            {
              name  = "RPC_URL"
              value = "https://sepolia.base.org"
            },
            {
              name  = "POLYMARKET_TRADES_URL"
              value = "https://clob.polymarket.com/data/trades"
            },
            {
              name  = "POLYMARKET_ALERT_TO_NUMBER"
              value = "+14407858475"
            },
            {
              name  = "POLYMARKET_ALERT_FROM_NUMBER"
              value = var.twilio_phone_number
            },
            {
              name  = "POLYMARKET_SUSPICIOUS_NOTIONAL_THRESHOLD"
              value = "1000"
            },
            {
              name  = "POLYMARKET_SUSPICIOUS_SIZE_THRESHOLD"
              value = "0"
            },
            {
              name  = "POLYMARKET_TRADES_LIMIT"
              value = "100"
            },
            {
              # Must match core's REEF_MAX_BBOX_MM — a Viper config-loading
              # bug meant job-runner silently ignored any REEF_* override
              # attempt here until it was fixed (see
              # go/job-runner/internal/config/config.go), so core's preview
              # and job-runner's real checkout validation could accept
              # different max sizes without anyone setting them differently
              # on purpose.
              name  = "REEF_MAX_BBOX_MM"
              value = "250"
            },
            {
              # Minor/localized support (e.g. frag_rack's small vent-channel
              # overhang, measured ~2.4% on a real print) still passes; only
              # a design needing scaffolding through a substantial fraction
              # of the print rejects. See go/pkg/reef/validate's
              # checkExcessiveSupport.
              name  = "REEF_MAX_SUPPORT_MATERIAL_PCT"
              value = "10"
            },
            {
              # Without this, PrusaSlicer reports every part's weight as
              # exactly 0.00g (confirmed against a real 2.7.4 binary),
              # zeroing out pricing.Price's material-cost component on every
              # order. 1.27 g/cm3 is standard PETG; update if the actual
              # filament differs meaningfully.
              name  = "REEF_FILAMENT_DENSITY_G_CM3"
              value = "1.27"
            }
          ]
          portMappings = [
            {
              name          = "job-runner"
              containerPort = 9013
              hostPort      = 9013
              protocol      = "tcp"
            }
          ]
        }

      "travel-angels-billing" = {
          essential = true
          secrets = [{
            name      = "STRIPE_SECRET_KEY",
            valueFrom = "${aws_secretsmanager_secret.travel_angels_stripe_secret_key.arn}"
          }, {
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }]
          image = "${aws_ecr_repository.billing.repository_url}:latest"
          portMappings = [
            {
              name          = "travel-angels-billing"
              containerPort = 8022
              hostPort      = 8022
              protocol      = "tcp"
            }
          ]
        }

        "ethereum-transactor" = {
          essential = true
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "PRIVATE_KEY",
            valueFrom = "${aws_secretsmanager_secret.ethereum_private_key.arn}"
          }]
          image = "${aws_ecr_repository.ethereum_transactor.repository_url}:latest"
          environment = [
            {
              name  = "DB_HOST"
              value = aws_db_instance.poltergeist-db.address
            },
            {
              name  = "DB_USER"
              value = aws_db_instance.poltergeist-db.username
            },
            {
              name  = "DB_PORT"
              value = tostring(aws_db_instance.poltergeist-db.port)
            },
            {
              name  = "DB_NAME"
              value = aws_db_instance.poltergeist-db.db_name
            },
            {
              name  = "CHAIN_ID"
              value = "84532"
            },
            {
              name  = "RPC_URL"
              value = "https://sepolia.base.org"
            }
          ]
          portMappings = [
            {
              name          = "ethereum-transactor"
              containerPort = 8088
              hostPort      = 8088
              protocol      = "tcp"
            }
          ]
        }
      }

      service_connect_configuration = {
          namespace = aws_service_discovery_http_namespace.sonar_namespace.arn
          service = [{
            client_alias = {
              port     = local.core_container_port
              dns_name = local.core_container_name
            }
            port_name      = local.core_container_name
            discovery_name = local.core_container_name
          }]
        }

        load_balancer = {
          service = {
            target_group_arn = module.sonar_alb.target_groups["core"].arn
            container_name   = local.core_container_name
            container_port   = local.core_container_port
          }
        }

        subnet_ids = module.vpc.private_subnets
        security_group_ingress_rules = {
          alb_ingress_8080 = {
            from_port                = 0
            to_port                  = local.core_container_port
            protocol                 = "tcp"
            description              = "Service port"
            referenced_security_group_id = module.sonar_alb_sg.security_group_id
          }
        }
              security_group_egress_rules = {
        all = {
          ip_protocol = "-1"
          cidr_ipv4   = "0.0.0.0/0"
        }
      }
    }
  }
}
