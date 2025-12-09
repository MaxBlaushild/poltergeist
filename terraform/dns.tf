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
