output "website_url" {
  description = "Public HTTPS address for the Travel Angels web app"
  value       = "https://${local.domain}"
}

output "cloudfront_url" {
  description = "Existing CloudFront HTTPS address retained as a fallback"
  value       = "https://${aws_cloudfront_distribution.web.domain_name}"
}

output "bucket_name" {
  description = "Private bucket to receive the Flutter build/web files"
  value       = aws_s3_bucket.web.id
}

output "distribution_id" {
  description = "CloudFront distribution to invalidate after publishing app files"
  value       = aws_cloudfront_distribution.web.id
}
