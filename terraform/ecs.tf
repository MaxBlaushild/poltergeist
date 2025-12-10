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
        aws_secretsmanager_secret.hue_bridge_hostname.arn,
        aws_secretsmanager_secret.hue_bridge_username.arn,
        aws_secretsmanager_secret.hue_client_id.arn,
        aws_secretsmanager_secret.hue_client_secret.arn,
        aws_secretsmanager_secret.hue_application_key.arn,
        aws_secretsmanager_secret.travel_angels_stripe_secret_key.arn,
      ]

      container_definitions = {
        "core" = {
          essential = true
          image     = "${aws_ecr_repository.core.repository_url}:latest"
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
          }]
          image = "${aws_ecr_repository.job_runner.repository_url}:latest"
          portMappings = [
            {
              name          = "job-runner"
              containerPort = 9013
              hostPort      = 9013
              protocol      = "tcp"
            }
          ]
        }

        "travel-angels" = {
          essential = true
          image = "${aws_ecr_repository.travel_angels.repository_url}:latest"
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "GOOGLE_DRIVE_CLIENT_ID",
            valueFrom = "${aws_secretsmanager_secret.google_drive_client_id.arn}"
          }, {
            name      = "GOOGLE_DRIVE_CLIENT_SECRET",
            valueFrom = "${aws_secretsmanager_secret.google_drive_client_secret.arn}"
          }, {
            name      = "MAPBOX_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.mapbox_api_key.arn}"
          }, {
            name      = "GOOGLE_MAPS_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.google_maps_api_key.arn}"
          }]
          portMappings = [
            {
              name          = "travel-angels"
              containerPort = 8083
              hostPort      = 8083
              protocol      = "tcp"
            }
          ]
        }

        "sonar" = {
          essential = true
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "IMAGINE_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.imagine_api_key.arn}"
          }, {
            name      = "USE_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.use_api_key.arn}"
          },
          {
            name      = "MAPBOX_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.mapbox_api_key.arn}"
          }, {
            name      = "GOOGLE_MAPS_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.google_maps_api_key.arn}"
          }]
          image = "${aws_ecr_repository.sonar.repository_url}:latest"
          portMappings = [
            {
              name          = "sonar"
              containerPort = 8042
              hostPort      = 8042
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
        
        "final-fete" = {
          essential = true
          image = "${aws_ecr_repository.final_fete.repository_url}:latest"
          secrets = [{
            name      = "HUE_BRIDGE_HOSTNAME",
            valueFrom = "${aws_secretsmanager_secret.hue_bridge_hostname.arn}"
          }, {
            name      = "HUE_BRIDGE_USERNAME",
            valueFrom = "${aws_secretsmanager_secret.hue_bridge_username.arn}"
          }, {
            name      = "HUE_CLIENT_ID",
            valueFrom = "${aws_secretsmanager_secret.hue_client_id.arn}"
          }, {
            name      = "HUE_CLIENT_SECRET",
            valueFrom = "${aws_secretsmanager_secret.hue_client_secret.arn}"
          }, {
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "HUE_APPLICATION_KEY",
            valueFrom = "${aws_secretsmanager_secret.hue_application_key.arn}"
          }]
          portMappings = [
            {
              name          = "final-fete"
              containerPort = 8085
              hostPort      = 8085
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

