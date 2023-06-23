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

resource "aws_secretsmanager_secret" "open_ai_key" {
  name = "OPEN_AI_KEY"
}

variable "open_ai_key" {
  description = "API key for Open AI"
  type        = string
}

resource "aws_secretsmanager_secret_version" "open_ai_key" {
  secret_id     = aws_secretsmanager_secret.open_ai_key.id
  secret_string = var.open_ai_key
}

resource "aws_secretsmanager_secret" "smart_things_pat" {
  name = "SMART_THINGS_PAT"
}

variable "smart_things_pat" {
  description = "Personal access token for smart things"
  type        = string
}

resource "aws_secretsmanager_secret_version" "smart_things_pat" {
  secret_id     = aws_secretsmanager_secret.smart_things_pat.id
  secret_string = var.smart_things_pat
}

resource "aws_ecr_repository" "core" {
  name                 = "core"  
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

resource "aws_ecr_repository" "ear" {
  name                 = "ear"  
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
    poltergeist_core = {
      cpu    = 1024
      memory = 4096

      # Container definition(s)
      container_definitions = {
        "core" = {
          cpu       = 256
          memory    = 1024
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
          cpu       = 512
          memory    = 1024
          essential = true
          secrets = [{
            name = "OPEN_AI_KEY",
            valueFrom = "${aws_secretsmanager_secret.open_ai_key.arn}"
          }]
          image     = "${aws_ecr_repository.fount_of_erebos.repository_url}:latest"
          port_mappings = [
            {
              name          = "fount-of-erebos"
              containerPort = 8081
              hostPort      = 8081
              protocol      = "tcp"
            }
          ]
        }

        "ear" = {
          cpu       = 256
          memory    = 1024
          essential = true
          image     = "${aws_ecr_repository.ear.repository_url}:latest"
          port_mappings = [
            {
              name          = "ear"
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

  tags = local.tags
}


################################################################################
# Supporting Resources
################################################################################

resource "aws_service_discovery_http_namespace" "this" {
  name        = local.name
  description = "CloudMap namespace for ${local.name}"
  tags        = local.tags
}

data "aws_acm_certificate" "cert" {
  domain = "digigeist.com"
  statuses = ["ISSUED"]
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

resource "aws_route53_record" "record" {
  zone_id = "Z0649695PXJXXKD92YP0"
  name    = "digigeist.com"
  type    = "A"

  alias {
    name                   = module.alb.lb_dns_name
    zone_id                = module.alb.lb_zone_id
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
      port = 443
      protocol = "HTTPS",
      ssl_policy   = "ELBSecurityPolicy-2016-08"
      certificate_arn   = data.aws_acm_certificate.cert.arn
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