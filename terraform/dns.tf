resource "aws_route53_record" "sonar_route_53_record" {
  zone_id = "Z03197012NLXJ6B0OCCFX"
  name    = "api.eeee.rsvp"
  type    = "A"

  alias {
    name                   = module.sonar_alb.dns_name
    zone_id                = module.sonar_alb.zone_id
    evaluate_target_health = true
  }
}
