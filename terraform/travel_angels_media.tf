# Covers must follow the trip's current audience, including private drafts.
# Clients fetch them through the API; no public read policy or CORS is needed.
resource "aws_s3_bucket" "travel_angels_trip_media" {
  bucket = "travel-angels-trip-media-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.tags, { Name = "travel-angels-trip-media" })
}

resource "aws_s3_bucket_public_access_block" "travel_angels_trip_media" {
  bucket = aws_s3_bucket.travel_angels_trip_media.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "travel_angels_trip_media" {
  bucket = aws_s3_bucket.travel_angels_trip_media.id
  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "travel_angels_trip_media" {
  bucket = aws_s3_bucket.travel_angels_trip_media.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
