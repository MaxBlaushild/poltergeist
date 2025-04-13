provider "aws" {
  region = local.region
}

data "aws_availability_zones" "available" {}

locals {
  region = "us-east-1"
  name   = "poltergeist"

  vpc_cidr = "10.0.0.0/16"
  azs      = slice(data.aws_availability_zones.available.names, 0, 3)

  core_container_name = "core"
  core_container_port = 8080

  tags = {
    Name       = local.name
    Example    = local.name
    Repository = "https://github.com/terraform-aws-modules/terraform-aws-ecs"
  }
}

resource "aws_secretsmanager_secret" "auth_private_key" {
  name = "AUTH_PRIVATE_KEY"
}

variable "auth_private_key" {
  description = "Auth private key"
  type        = string
}

resource "aws_secretsmanager_secret_version" "auth_private_key" {
  secret_id     = aws_secretsmanager_secret.auth_private_key.id
  secret_string = var.auth_private_key
}

resource "aws_secretsmanager_secret" "open_ai_key" {
  name = "OPEN_AI_KEY"
}

resource "aws_secretsmanager_secret" "slack_scorekeeper_token" {
  name = "SLACK_SCOREKEEPER_TOKEN"
}

variable "slack_scorekeeper_token" {
  description = "Slack Scorekeeper Token"
  type        = string
}

resource "aws_secretsmanager_secret_version" "slack_scorekeeper_token" {
  secret_id     = aws_secretsmanager_secret.slack_scorekeeper_token.id
  secret_string = var.slack_scorekeeper_token
}

resource "aws_secretsmanager_secret" "slack_scorekeeper_webhook_url" {
  name = "SLACK_SCOREKEEPER_WEBHOOK_URL"
}

variable "slack_scorekeeper_webhook_url" {
  description = "Slack Scorekeeper Webhook Url"
  type        = string
}

resource "aws_secretsmanager_secret_version" "slack_scorekeeper_webhook_url" {
  secret_id     = aws_secretsmanager_secret.slack_scorekeeper_webhook_url.id
  secret_string = var.slack_scorekeeper_webhook_url
}

resource "aws_secretsmanager_secret" "twilio_phone_number" {
  name = "TWILIO_PHONE_NUMBER"
}

variable "twilio_phone_number" {
  description = "Twilio Phone Number"
  type        = string
}

resource "aws_secretsmanager_secret_version" "twilio_phone_number" {
  secret_id     = aws_secretsmanager_secret.twilio_phone_number.id
  secret_string = var.twilio_phone_number
}

resource "aws_secretsmanager_secret" "google_maps_api_key" {
  name = "GOOGLE_MAPS_API_KEY"
}

variable "google_maps_api_key" {
  description = "Google Maps API Key"
  type        = string
}

resource "aws_secretsmanager_secret_version" "google_maps_api_key" {
  secret_id     = aws_secretsmanager_secret.google_maps_api_key.id
  secret_string = var.google_maps_api_key
}

resource "aws_secretsmanager_secret" "twilio_auth_token" {
  name = "TWILIO_AUTH_TOKEN"
}

variable "twilio_auth_token" {
  description = "Twilio AuthToken"
  type        = string
}

resource "aws_secretsmanager_secret_version" "twilio_auth_token" {
  secret_id     = aws_secretsmanager_secret.twilio_auth_token.id
  secret_string = var.twilio_auth_token
}

resource "aws_secretsmanager_secret" "twilio_account_sid" {
  name = "TWILIO_ACCOUNT_SID"
}

variable "twilio_account_sid" {
  description = "Twilio Account SID"
  type        = string
}

resource "aws_secretsmanager_secret_version" "stripe_secret_key" {
  secret_id     = aws_secretsmanager_secret.stripe_secret_key.id
  secret_string = var.stripe_secret_key
}

resource "aws_secretsmanager_secret" "stripe_secret_key" {
  name = "STRIPE_SECRET_KEY"
}

variable "stripe_secret_key" {
  description = "Stripe Secret Key"
  type        = string
}

resource "aws_secretsmanager_secret_version" "use_api_key" {
  secret_id     = aws_secretsmanager_secret.use_api_key.id
  secret_string = var.use_api_key
}

resource "aws_secretsmanager_secret" "use_api_key" {
  name = "USE_API_KEY"
}

variable "use_api_key" {
  description = "Use API Key"
  type        = string
}
  
resource "aws_secretsmanager_secret_version" "mapbox_api_key" {
  secret_id     = aws_secretsmanager_secret.mapbox_api_key.id
  secret_string = var.mapbox_api_key
}

resource "aws_secretsmanager_secret" "mapbox_api_key" {
  name = "MAPBOX_API_KEY"
}

variable "mapbox_api_key" {
  description = "Mapbox API Key"
  type        = string
}

resource "aws_secretsmanager_secret_version" "imagine_api_key" {
  secret_id     = aws_secretsmanager_secret.imagine_api_key.id
  secret_string = var.imagine_api_key
}

resource "aws_secretsmanager_secret" "imagine_api_key" {
  name = "IMAGINE_API_KEY"
}

variable "imagine_api_key" {
  description = "Imagine API Key"
  type        = string
}

resource "aws_secretsmanager_secret_version" "twilio_account_sid" {
  secret_id     = aws_secretsmanager_secret.twilio_account_sid.id
  secret_string = var.twilio_account_sid
}

variable "open_ai_key" {
  description = "API key for Open AI"
  type        = string
}


resource "aws_secretsmanager_secret_version" "open_ai_key" {
  secret_id     = aws_secretsmanager_secret.open_ai_key.id
  secret_string = var.open_ai_key
}

resource "aws_ecr_repository" "core" {
  name                 = "core"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "authenticator" {
  name                 = "authenticator"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "admin" {
  name                 = "admin"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "billing" {
  name                 = "billing"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "texter" {
  name                 = "texter"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "scorekeeper" {
  name                 = "scorekeeper"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "fount_of_erebos" {
  name                 = "fount-of-erebos"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "job_runner" {
  name                 = "job-runner"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "crystal_crisis" {
  name                 = "crystal-crisis"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "ear" {
  name                 = "ear"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "trivai" {
  name                 = "trivai"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "sonar" {
  name                 = "sonar"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

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

    poltergeist_core = {
      cpu    = 1024
      memory = 2048

      # Container definition(s)
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

        "trivai" = {
          cpu       = 256
          memory    = 512
          essential = true
          secrets = [{
            name      = "DB_PASSWORD",
            valueFrom = "${aws_secretsmanager_secret.db_password.arn}"
            }, {
            name      = "SENDGRID_API_KEY",
            valueFrom = "${aws_secretsmanager_secret.sendgrid_api_key.arn}"
            }, {
            name      = "GUESS_HOW_MANY_PHONE_NUMBER",
            valueFrom = "${aws_secretsmanager_secret.twilio_phone_number.arn}"
          }]
          image = "${aws_ecr_repository.trivai.repository_url}:latest"
          port_mappings = [
            {
              name          = "trivai"
              containerPort = 8082
              hostPort      = 8082
              protocol      = "tcp"
            }
          ]
        }
      }

      service_connect_configuration = {
        namespace = aws_service_discovery_http_namespace.this.arn
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
          target_group_arn = element(module.alb.target_group_arns, 0)
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
          source_security_group_id = module.alb_sg.security_group_id
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
  }
}


################################################################################
# Supporting Resources
################################################################################

resource "aws_service_discovery_http_namespace" "this" {
  name        = local.name
  description = "CloudMap namespace for ${local.name}"
  tags        = local.tags
}

resource "aws_service_discovery_http_namespace" "sonar_namespace" {
  name        = "Sonar"
  description = "CloudMap namespace for sonar"
  tags        = local.tags
}

data "aws_acm_certificate" "guesswith_us_cert" {
  domain      = "*.guesswith.us"
  statuses    = ["ISSUED"]
  most_recent = true
}


data "aws_acm_certificate" "sonar_cert" {
  domain      = "*.unclaimedstreets.com"
  statuses    = ["ISSUED"]
  most_recent = true
}

module "alb_sg" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 4.0"

  name        = "${local.name}-service"
  description = "Service security group"
  vpc_id      = module.vpc.vpc_id

  ingress_rules       = ["http-80-tcp", "https-443-tcp"]
  ingress_cidr_blocks = ["0.0.0.0/0"]

  egress_rules       = ["all-all"]
  egress_cidr_blocks = module.vpc.private_subnets_cidr_blocks

  tags = local.tags
}

module "sonar_alb_sg" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 4.0"

  name        = "sonar-service"
  description = "Service security group"
  vpc_id      = module.vpc.vpc_id

  ingress_rules       = ["http-80-tcp", "https-443-tcp"]
  ingress_cidr_blocks = ["0.0.0.0/0"]

  egress_rules       = ["all-all"]
  egress_cidr_blocks = module.vpc.private_subnets_cidr_blocks

  tags = local.tags
}

resource "aws_route53_record" "api_guesswith_us_record" {
  zone_id = "Z02223351NOY9TTILWBS2"
  name    = "api.guesswith.us"
  type    = "A"

  alias {
    name                   = module.alb.lb_dns_name
    zone_id                = module.alb.lb_zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "sonar_route_53_record" {
  zone_id = "Z03197012NLXJ6B0OCCFX"
  name    = "api.eeee.rsvp"
  type    = "A"

  alias {
    name                   = module.sonar_alb.lb_dns_name
    zone_id                = module.sonar_alb.lb_zone_id
    evaluate_target_health = true
  }
}

module "alb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 8.0"

  name = local.name

  load_balancer_type = "application"

  vpc_id          = module.vpc.vpc_id
  subnets         = module.vpc.public_subnets
  security_groups = [module.alb_sg.security_group_id]

  http_tcp_listeners = [
    {
      port               = 80
      protocol           = "HTTP"
      target_group_index = 0
    },
  ]

  https_listeners = [
    {
      port               = 443
      protocol           = "HTTPS",
      ssl_policy         = "ELBSecurityPolicy-2016-08"
      certificate_arn    = data.aws_acm_certificate.guesswith_us_cert.arn
      target_group_index = 0
    }
  ]

  target_groups = [
    {
      name             = "${local.name}-${local.core_container_name}"
      backend_protocol = "HTTP"
      backend_port     = local.core_container_port
      target_type      = "ip"
    },
  ]

  tags = local.tags
}

module "sonar_alb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 8.0"

  name = "sonar"

  load_balancer_type = "application"

  vpc_id          = module.vpc.vpc_id
  subnets         = module.vpc.public_subnets
  security_groups = [module.sonar_alb_sg.security_group_id]

  http_tcp_listeners = [
    {
      port               = 80
      protocol           = "HTTP"
      target_group_index = 0
    },
  ]

  https_listeners = [
    {
      port               = 443
      protocol           = "HTTPS",
      ssl_policy         = "ELBSecurityPolicy-2016-08"
      certificate_arn    = data.aws_acm_certificate.sonar_cert.arn
      target_group_index = 0
    }
  ]

  target_groups = [
    {
      name             = "sonar-${local.core_container_name}"
      backend_protocol = "HTTP"
      backend_port     = local.core_container_port
      target_type      = "ip"
    },
  ]

  tags = local.tags
}


module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 4.0"

  name = local.name
  cidr = local.vpc_cidr

  azs             = local.azs
  private_subnets = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 4, k)]
  public_subnets  = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 8, k + 48)]

  enable_nat_gateway = true
  single_nat_gateway = true

  tags = local.tags
}

resource "aws_secretsmanager_secret" "db_password" {
  name = "DB_PASSWORD"
}

variable "db_password" {
  description = "Password for da db"
  type        = string
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = var.db_password
}

resource "aws_secretsmanager_secret" "sendgrid_api_key" {
  name = "SENDGRID_API_KEY"
}

variable "sendgrid_api_key" {
  description = "key for the sendgrid"
  type        = string
}

resource "aws_secretsmanager_secret_version" "sendgrid_api_key" {
  secret_id     = aws_secretsmanager_secret.sendgrid_api_key.id
  secret_string = var.sendgrid_api_key
}

resource "aws_subnet" "main" {
  vpc_id            = module.vpc.vpc_id
  cidr_block        = "10.0.92.0/28"
  availability_zone = "us-east-1a"
}

resource "aws_subnet" "secondary" {
  vpc_id            = module.vpc.vpc_id
  cidr_block        = "10.0.184.0/28"
  availability_zone = "us-east-1b"
}

resource "aws_db_subnet_group" "main" {
  name       = "main"
  subnet_ids = [aws_subnet.main.id, aws_subnet.secondary.id]

  tags = {
    Name = "main"
  }
}

resource "aws_elasticache_subnet_group" "redis_subnet_group" {
  name       = "redis-subnet-group"
  subnet_ids = [aws_subnet.main.id, aws_subnet.secondary.id]
}

resource "aws_security_group" "redis_sg" {
  name        = "redis_sg"
  description = "Allow inbound traffic for Redis"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "Redis"
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "poltergeist-redis"
  engine               = "redis"
  node_type            = "cache.t3.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis6.x"
  engine_version       = "6.x"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.redis_subnet_group.name
  security_group_ids   = [aws_security_group.redis_sg.id]
}

resource "aws_security_group" "db_sg" {
  name        = "db_sg"
  description = "Allow inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "PostgreSQL"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "poltergeist-db" {
  identifier             = "poltergeist"
  engine                 = "postgres"
  engine_version         = "15.7"
  instance_class         = "db.t3.micro"
  allocated_storage      = 20
  db_name                = "poltergeist"
  username               = "db_user"
  password               = var.db_password
  skip_final_snapshot    = true
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.db_sg.id]

  parameter_group_name    = "default.postgres15"
  backup_retention_period = 1
}

resource "aws_key_pair" "lappentoppen" {
  key_name   = "lappentoppen"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHR4WRxoB3jd6Q/qFXBIgKqkPwo9gXyzUHctXpZgeMx0"
}

resource "aws_security_group" "allow_ssh" {
  name        = "allow_ssh"
  description = "Allow SSH inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "ssh_box" {
  ami                    = "ami-05c13eab67c5d8861" // free cheap linux box
  instance_type          = "t2.nano"
  key_name               = aws_key_pair.lappentoppen.key_name
  subnet_id              = module.vpc.public_subnets[0]
  vpc_security_group_ids = [aws_security_group.allow_ssh.id, aws_security_group.db_sg.id, aws_security_group.redis_sg.id]
  associate_public_ip_address = true

  tags = {
    Name = "SshBox"
  }
}

