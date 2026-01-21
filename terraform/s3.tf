# Get current AWS account ID
data "aws_caller_identity" "current" {}

# S3 bucket for verifiable-sn posts
resource "aws_s3_bucket" "verifiable_sn_posts" {
  bucket = "verifiable-sn-posts"

  tags = {
    Name        = "verifiable-sn-posts"
    Environment = "production"
  }
}

# Block public access settings - allow public read
resource "aws_s3_bucket_public_access_block" "verifiable_sn_posts" {
  bucket = aws_s3_bucket.verifiable_sn_posts.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# Bucket policy for public read and public write
resource "aws_s3_bucket_policy" "verifiable_sn_posts" {
  bucket = aws_s3_bucket.verifiable_sn_posts.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "PublicReadGetObject"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.verifiable_sn_posts.arn}/*"
      },
      {
        Sid    = "PublicWrite"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action = [
          "s3:PutObject",
          "s3:PutObjectAcl"
        ]
        Resource = "${aws_s3_bucket.verifiable_sn_posts.arn}/*"
      },
      {
        Sid    = "PublicListBucket"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action = [
          "s3:ListBucket",
          "s3:GetBucketLocation"
        ]
        Resource = aws_s3_bucket.verifiable_sn_posts.arn
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.verifiable_sn_posts]
}

# Enable CORS for web access (including PUT for uploads)
resource "aws_s3_bucket_cors_configuration" "verifiable_sn_posts" {
  bucket = aws_s3_bucket.verifiable_sn_posts.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD", "PUT", "POST"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# Enable versioning (optional but recommended)
resource "aws_s3_bucket_versioning" "verifiable_sn_posts" {
  bucket = aws_s3_bucket.verifiable_sn_posts.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Server-side encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "verifiable_sn_posts" {
  bucket = aws_s3_bucket.verifiable_sn_posts.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Note: Since the bucket is now open for public writes, we don't need IAM policies
# for the ECS task role. Anyone can write to the bucket via the bucket policy.

