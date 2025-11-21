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

