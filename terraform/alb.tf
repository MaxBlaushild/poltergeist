module "sonar_alb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 10.0"

  name = "sonar"

  load_balancer_type = "application"

  vpc_id          = module.vpc.vpc_id
  subnets         = module.vpc.public_subnets
  security_groups = [module.sonar_alb_sg.security_group_id]

  listeners = {
    http = {
      port     = 80
      protocol = "HTTP"
      forward = {
        target_group_key = "core"
      }
    }
    https = {
      port            = 443
      protocol        = "HTTPS"
      ssl_policy      = "ELBSecurityPolicy-2016-08"
      certificate_arn = data.aws_acm_certificate.sonar_cert.arn
      forward = {
        target_group_key = "core"
      }
    }
  }

  target_groups = {
    core = {
      name_prefix      = "s-core"
      backend_protocol = "HTTP"
      backend_port     = local.core_container_port
      target_type      = "ip"
      create_attachment = false
    }
  }

  tags = local.tags
}

