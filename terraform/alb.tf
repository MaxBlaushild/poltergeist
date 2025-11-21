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

