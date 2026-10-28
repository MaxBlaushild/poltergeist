# Registration and the hosted zone already belong to the domain owner. This
# root manages only the web aliases and the certificate validation records.
data "aws_route53_zone" "web" {
  name         = local.domain
  private_zone = false
}

# The provider is fixed to us-east-1, as required for CloudFront certificates.
resource "aws_acm_certificate" "web" {
  domain_name               = local.domain
  subject_alternative_names = ["www.${local.domain}"]
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "certificate_validation" {
  for_each = {
    for option in aws_acm_certificate.web.domain_validation_options : option.domain_name => {
      name  = option.resource_record_name
      type  = option.resource_record_type
      value = option.resource_record_value
    }
  }

  zone_id         = data.aws_route53_zone.web.zone_id
  name            = each.value.name
  type            = each.value.type
  records         = [each.value.value]
  ttl             = 60
  allow_overwrite = false
}

resource "aws_acm_certificate_validation" "web" {
  certificate_arn         = aws_acm_certificate.web.arn
  validation_record_fqdns = [for record in aws_route53_record.certificate_validation : record.fqdn]
}

resource "aws_route53_record" "web_ipv4" {
  for_each = local.domains

  zone_id         = data.aws_route53_zone.web.zone_id
  name            = each.value
  type            = "A"
  allow_overwrite = false

  alias {
    name                   = aws_cloudfront_distribution.web.domain_name
    zone_id                = aws_cloudfront_distribution.web.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "web_ipv6" {
  for_each = local.domains

  zone_id         = data.aws_route53_zone.web.zone_id
  name            = each.value
  type            = "AAAA"
  allow_overwrite = false

  alias {
    name                   = aws_cloudfront_distribution.web.domain_name
    zone_id                = aws_cloudfront_distribution.web.hosted_zone_id
    evaluate_target_health = false
  }
}
