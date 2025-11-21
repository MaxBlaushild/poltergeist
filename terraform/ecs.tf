module "ecs" {
  source = "terraform-aws-modules/ecs/aws"

  cluster_name = local.name

  # Capacity provider
  fargate_capacity_providers = {
    FARGATE = {
      default_capacity_provider_strategy = {
        weight = 50
        base   = 20
      }
    }
    FARGATE_SPOT = {
      default_capacity_provider_strategy = {
        weight = 50
      }
    }
  }

  services = {
    sonar_core = {
      cpu = 2048
      memory = 4096

      container_definitions = {
        "core" = {
          cpu       = 256
          memory    = 512
          essential = true
          image     = "${aws_ecr_repository.core.repository_url}:latest"
          port_mappings = [
            {
              name          = local.core_container_name
              containerPort = local.core_container_port
              hostPort      = local.core_container_port
              protocol      = "tcp"
            }
          ]
        }

        "authenticator" = {
          cpu       = 256
          memory    = 512
          essential = true
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
          }, {
            name      = "AUTH_PRIVATE_KEY",
            valueFrom = "${aws_secretsmanager_secret.auth_private_key.arn}"
          }]
          image = "${aws_ecr_repository.authenticator.repository_url}:latest"
          port_mappings = [
            {
              name          = "authenticator"
              containerPort = 8089
              hostPort      = 8089
              protocol      = "tcp"
            }
          ]
        }

        "fount-of-erebos" = {
          cpu       = 256
          memory    = 512
          essential = true
          secrets = [{
            name      = "OPEN_AI_KEY",
            valueFrom = "${aws_secretsmanager_secret.open_ai_key.arn}"
          }]
          image = "${aws_ecr_repository.fount_of_erebos.repository_url}:latest"
          port_mappings = [
            {
              name          = "fount-of-erebos"
              containerPort = 8081
              hostPort      = 8081
              protocol      = "tcp"
            }
          ]
        }

        "texter" = {
          cpu       = 256
          memory    = 512
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
          port_mappings = [
            {
              name          = "texter"
              containerPort = 8084
              hostPort      = 8084
              protocol      = "tcp"
            }
          ]
        }

        "job-runner" = {
          cpu       = 256
          memory    = 512
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
          port_mappings = [
            {
              name          = "job-runner"
              containerPort = 9013
              hostPort      = 9013
              protocol      = "tcp"
            }
          ]
        }

        "travel-angels" = {
          cpu       = 256
          memory    = 512
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
          }]
          port_mappings = [
            {
              name          = "travel-angels"
              containerPort = 8083
              hostPort      = 8083
              protocol      = "tcp"
            }
          ]
        }

        "sonar" = {
          cpu       = 256
          memory    = 512
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
          port_mappings = [
            {
              name          = "sonar"
              containerPort = 8042
              hostPort      = 8042
              protocol      = "tcp"
            }
          ]
        }
      }

      service_connect_configuration = {
          namespace = aws_service_discovery_http_namespace.sonar_namespace.arn
          service = {
            client_alias = {
              port     = local.core_container_port
              dns_name = local.core_container_name
            }
            port_name      = local.core_container_name
            discovery_name = local.core_container_name
          }
        }

        load_balancer = {
          service = {
            target_group_arn = element(module.sonar_alb.target_group_arns, 0)
            container_name   = local.core_container_name
            container_port   = local.core_container_port
          }
        }

        subnet_ids = module.vpc.private_subnets
        security_group_rules = {
          alb_ingress_8080 = {
            type                     = "ingress"
            from_port                = 0
            to_port                  = local.core_container_port
            protocol                 = "tcp"
            description              = "Service port"
            source_security_group_id = module.sonar_alb_sg.security_group_id
          }
          egress_all = {
            type        = "egress"
            from_port   = 0
            to_port     = 0
            protocol    = "-1"
            cidr_blocks = ["0.0.0.0/0"]
          }
        }
    }

    # poltergeist_core = {
    #   cpu    = 1024
    #   memory = 2048

    #   # Container definition(s)
    #   container_definitions = {
    #     "core" = {
    #       cpu       = 256
    #       memory    = 512
    #       essential = true
    #       image     = "${aws_ecr_repository.core.repository_url}:latest"
    #       port_mappings = [
    #         {
    #           name          = local.core_container_name
    #           containerPort = local.core_container_port
    #           hostPort      = local.core_container_port
    #           protocol      = "tcp"
    #         }
    #       ]
    #     }

    #     "fount-of-erebos" = {
    #       cpu       = 256
    #       memory    = 512
    #       essential = true
    #       secrets = [{
    #         name      = "OPEN_AI_KEY",
    #         valueFrom = "${aws_secretsmanager_secret.open_ai_key.arn}"
    #       }]
    #       image = "${aws_ecr_repository.fount_of_erebos.repository_url}:latest"
    #       port_mappings = [
    #         {
    #           name          = "fount-of-erebos"
    #           containerPort = 8081
    #           hostPort      = 8081
    #           protocol      = "tcp"
        #     }
        #   ]
        # }

        # "trivai" = {
        #   cpu       = 256
        #   memory    = 512
        #   essential = true
        #   secrets = [{
        #     name      = "DB_PASSWORD",
        #     valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
        #     }, {
        #     name      = "SENDGRID_API_KEY",
        #     valueFrom = "${aws_secretsmanager_secret.sendgrid_api_key.arn}"
        #     }, {
        #     name      = "GUESS_HOW_MANY_PHONE_NUMBER",
        #     valueFrom = "${aws_secretsmanager_secret.twilio_phone_number.arn}"
        #   }]
      #     image = "${aws_ecr_repository.trivai.repository_url}:latest"
      #     port_mappings = [
      #       {
      #         name          = "trivai"
      #         containerPort = 8082
      #         hostPort      = 8082
      #         protocol      = "tcp"
      #       }
      #     ]
      #   }
      # }

      # service_connect_configuration = {
      #   namespace = aws_service_discovery_http_namespace.this.arn
      #   service = {
      #     client_alias = {
      #       port     = local.core_container_port
      #       dns_name = local.core_container_name
      #     }
      #     port_name      = local.core_container_name
      #     discovery_name = local.core_container_name
      #   }
      # }

    #   load_balancer = {
    #     service = {
    #       target_group_arn = element(module.alb.target_group_arns, 0)
    #       container_name   = local.core_container_name
    #       container_port   = local.core_container_port
    #     }
    #   }

    #   subnet_ids = module.vpc.private_subnets
    #   security_group_rules = {
    #     alb_ingress_8080 = {
    #       type                     = "ingress"
    #       from_port                = 0
    #       to_port                  = local.core_container_port
    #       protocol                 = "tcp"
    #       description              = "Service port"
    #       source_security_group_id = module.alb_sg.security_group_id
    #     }
    #     egress_all = {
    #       type        = "egress"
    #       from_port   = 0
    #       to_port     = 0
    #       protocol    = "-1"
    #       cidr_blocks = ["0.0.0.0/0"]
    #     }
    #   }
    # }
  }
}

