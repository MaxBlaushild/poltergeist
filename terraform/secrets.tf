resource "aws_secretsmanager_secret" "auth_private_key" {
  name = "AUTH_PRIVATE_KEY"
}

resource "aws_secretsmanager_secret_version" "auth_private_key" {
  secret_id     = aws_secretsmanager_secret.auth_private_key.id
  secret_string = var.auth_private_key
}

resource "aws_secretsmanager_secret" "open_ai_key" {
  name = "OPEN_AI_KEY"
}

resource "aws_secretsmanager_secret_version" "open_ai_key" {
  secret_id     = aws_secretsmanager_secret.open_ai_key.id
  secret_string = var.open_ai_key
}

resource "aws_secretsmanager_secret" "slack_scorekeeper_token" {
  name = "SLACK_SCOREKEEPER_TOKEN"
}

resource "aws_secretsmanager_secret_version" "slack_scorekeeper_token" {
  secret_id     = aws_secretsmanager_secret.slack_scorekeeper_token.id
  secret_string = var.slack_scorekeeper_token
}

resource "aws_secretsmanager_secret" "slack_scorekeeper_webhook_url" {
  name = "SLACK_SCOREKEEPER_WEBHOOK_URL"
}

resource "aws_secretsmanager_secret_version" "slack_scorekeeper_webhook_url" {
  secret_id     = aws_secretsmanager_secret.slack_scorekeeper_webhook_url.id
  secret_string = var.slack_scorekeeper_webhook_url
}

resource "aws_secretsmanager_secret" "twilio_phone_number" {
  name = "TWILIO_PHONE_NUMBER"
}

resource "aws_secretsmanager_secret_version" "twilio_phone_number" {
  secret_id     = aws_secretsmanager_secret.twilio_phone_number.id
  secret_string = var.twilio_phone_number
}

resource "aws_secretsmanager_secret" "google_maps_api_key" {
  name = "GOOGLE_MAPS_API_KEY"
}

resource "aws_secretsmanager_secret_version" "google_maps_api_key" {
  secret_id     = aws_secretsmanager_secret.google_maps_api_key.id
  secret_string = var.google_maps_api_key
}

resource "aws_secretsmanager_secret" "twilio_auth_token" {
  name = "TWILIO_AUTH_TOKEN"
}

resource "aws_secretsmanager_secret_version" "twilio_auth_token" {
  secret_id     = aws_secretsmanager_secret.twilio_auth_token.id
  secret_string = var.twilio_auth_token
}

resource "aws_secretsmanager_secret" "twilio_account_sid" {
  name = "TWILIO_ACCOUNT_SID"
}

resource "aws_secretsmanager_secret_version" "twilio_account_sid" {
  secret_id     = aws_secretsmanager_secret.twilio_account_sid.id
  secret_string = var.twilio_account_sid
}

resource "aws_secretsmanager_secret" "stripe_secret_key" {
  name = "STRIPE_SECRET_KEY"
}

resource "aws_secretsmanager_secret_version" "stripe_secret_key" {
  secret_id     = aws_secretsmanager_secret.stripe_secret_key.id
  secret_string = var.stripe_secret_key
}

resource "aws_secretsmanager_secret" "use_api_key" {
  name = "USE_API_KEY"
}

resource "aws_secretsmanager_secret_version" "use_api_key" {
  secret_id     = aws_secretsmanager_secret.use_api_key.id
  secret_string = var.use_api_key
}

resource "aws_secretsmanager_secret" "mapbox_api_key" {
  name = "MAPBOX_API_KEY"
}

resource "aws_secretsmanager_secret_version" "mapbox_api_key" {
  secret_id     = aws_secretsmanager_secret.mapbox_api_key.id
  secret_string = var.mapbox_api_key
}

resource "aws_secretsmanager_secret" "imagine_api_key" {
  name = "IMAGINE_API_KEY"
}

resource "aws_secretsmanager_secret_version" "imagine_api_key" {
  secret_id     = aws_secretsmanager_secret.imagine_api_key.id
  secret_string = var.imagine_api_key
}

resource "aws_secretsmanager_secret" "db_password" {
  name = "DB_PASSWORD"
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = var.db_password
}

resource "aws_secretsmanager_secret" "sendgrid_api_key" {
  name = "SENDGRID_API_KEY"
}

resource "aws_secretsmanager_secret_version" "sendgrid_api_key" {
  secret_id     = aws_secretsmanager_secret.sendgrid_api_key.id
  secret_string = var.sendgrid_api_key
}

resource "aws_secretsmanager_secret" "google_drive_client_id" {
  name = "GOOGLE_DRIVE_CLIENT_ID"
}

resource "aws_secretsmanager_secret_version" "google_drive_client_id" {
  secret_id     = aws_secretsmanager_secret.google_drive_client_id.id
  secret_string = var.google_drive_client_id
}

resource "aws_secretsmanager_secret" "google_drive_client_secret" {
  name = "GOOGLE_DRIVE_CLIENT_SECRET"
}

resource "aws_secretsmanager_secret_version" "google_drive_client_secret" {
  secret_id     = aws_secretsmanager_secret.google_drive_client_secret.id
  secret_string = var.google_drive_client_secret
}

resource "aws_secretsmanager_secret" "travel_angels_stripe_secret_key" {
  name = "TRAVEL_ANGELS_STRIPE_SECRET_KEY"
}

resource "aws_secretsmanager_secret_version" "travel_angels_stripe_secret_key" {
  secret_id     = aws_secretsmanager_secret.travel_angels_stripe_secret_key.id
  secret_string = var.travel_angels_stripe_secret_key
}

resource "aws_secretsmanager_secret" "hue_bridge_hostname" {
  name = "HUE_BRIDGE_HOSTNAME"
}

resource "aws_secretsmanager_secret_version" "hue_bridge_hostname" {
  secret_id     = aws_secretsmanager_secret.hue_bridge_hostname.id
  secret_string = var.hue_bridge_hostname
}

resource "aws_secretsmanager_secret" "hue_bridge_username" {
  name = "HUE_BRIDGE_USERNAME"
}

resource "aws_secretsmanager_secret_version" "hue_bridge_username" {
  secret_id     = aws_secretsmanager_secret.hue_bridge_username.id
  secret_string = var.hue_bridge_username
}

resource "aws_secretsmanager_secret" "hue_client_id" {
  name = "HUE_CLIENT_ID"
}

resource "aws_secretsmanager_secret_version" "hue_client_id" {
  secret_id     = aws_secretsmanager_secret.hue_client_id.id
  secret_string = var.hue_client_id
}

resource "aws_secretsmanager_secret" "hue_client_secret" {
  name = "HUE_CLIENT_SECRET"
}

resource "aws_secretsmanager_secret_version" "hue_client_secret" {
  secret_id     = aws_secretsmanager_secret.hue_client_secret.id
  secret_string = var.hue_client_secret
}

resource "aws_secretsmanager_secret" "hue_application_key" {
  name = "HUE_APPLICATION_KEY"
}

resource "aws_secretsmanager_secret_version" "hue_application_key" {
  secret_id     = aws_secretsmanager_secret.hue_application_key.id
  secret_string = var.hue_application_key
}

resource "aws_secretsmanager_secret" "dropbox_client_id" {
  name = "DROPBOX_CLIENT_ID"
}

resource "aws_secretsmanager_secret_version" "dropbox_client_id" {
  secret_id     = aws_secretsmanager_secret.dropbox_client_id.id
  secret_string = var.dropbox_client_id
}

resource "aws_secretsmanager_secret" "dropbox_client_secret" {
  name = "DROPBOX_CLIENT_SECRET"
}

resource "aws_secretsmanager_secret_version" "dropbox_client_secret" {
  secret_id     = aws_secretsmanager_secret.dropbox_client_secret.id
  secret_string = var.dropbox_client_secret
}