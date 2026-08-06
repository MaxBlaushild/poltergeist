variable "auth_private_key" {
  description = "Auth private key"
  type        = string
}

variable "slack_scorekeeper_token" {
  description = "Slack Scorekeeper Token"
  type        = string
}

variable "slack_scorekeeper_webhook_url" {
  description = "Slack Scorekeeper Webhook Url"
  type        = string
}

variable "twilio_phone_number" {
  description = "Twilio Phone Number"
  type        = string
}

variable "google_maps_api_key" {
  description = "Google Maps API Key"
  type        = string
}

variable "twilio_auth_token" {
  description = "Twilio AuthToken"
  type        = string
}

variable "twilio_account_sid" {
  description = "Twilio Account SID"
  type        = string
}

variable "stripe_secret_key" {
  description = "Stripe Secret Key"
  type        = string
}

variable "use_api_key" {
  description = "Use API Key"
  type        = string
}

variable "mapbox_api_key" {
  description = "Mapbox API Key"
  type        = string
}

variable "imagine_api_key" {
  description = "Imagine API Key"
  type        = string
}

variable "open_ai_key" {
  description = "API key for Open AI"
  type        = string
}

variable "db_password" {
  description = "Password for da db"
  type        = string
}

variable "sendgrid_api_key" {
  description = "key for the sendgrid"
  type        = string
}

variable "google_drive_client_id" {
  description = "Google Drive Client ID"
  type        = string
}

variable "google_drive_client_secret" {
  description = "Google Drive Client Secret"
  type        = string
}

variable "travel_angels_stripe_secret_key" {
  description = "Travel Angels Stripe Secret Key"
  type        = string
}

variable "hue_bridge_hostname" {
  description = "Hue Bridge Hostname"
  type        = string
}

variable "hue_bridge_username" {
  description = "Hue Bridge Username"
  type        = string
}

variable "hue_client_id" {
  description = "Hue Client ID"
  type        = string
}

variable "hue_client_secret" {
  description = "Hue Client Secret"
  type        = string
}

variable "hue_application_key" {
  description = "Hue Application Key"
  type        = string
}

variable "dropbox_client_id" {
  description = "Dropbox Client ID"
  type        = string
}

variable "dropbox_client_secret" {
  description = "Dropbox Client Secret"
  type        = string
}

variable "ethereum_private_key" {
  description = "Ethereum Private Key for transaction signing"
  type        = string
}

variable "ca_private_key" {
  description = "CA Private Key"
  type        = string
}

variable "polymarket_api_key" {
  description = "Polymarket API key"
  type        = string
  sensitive   = true
}

variable "polymarket_api_secret" {
  description = "Polymarket API secret (base64-url encoded)"
  type        = string
  sensitive   = true
}

variable "polymarket_api_passphrase" {
  description = "Polymarket API passphrase"
  type        = string
  sensitive   = true
}

variable "polymarket_address" {
  description = "Polymarket API address"
  type        = string
  sensitive   = true
}

variable "slant_api_key" {
  description = "Slant 3D API key (sl_*** — from slant3dapi.com account settings). Not yet in real use: R-7.3's sample-part gate hasn't been cleared, so this only powers the operator print queue's manual, per-order \"Send to Slant\" action. Placeholder default (Secrets Manager rejects an empty string) until a real key is supplied — Slant will simply reject requests with an auth error until then."
  type        = string
  sensitive   = true
  default     = "sl-not-configured"
}

variable "slant_platform_id" {
  description = "Slant 3D platform ID (UUID), created once via POST /platforms or the slant3dapi.com dashboard."
  type        = string
  default     = ""
}

variable "google_client_id" {
  description = "Google OAuth client ID for reef-site's \"Sign in with Google\" — also the audience checked when the authenticator verifies Google ID tokens. Not secret (embedded in the frontend bundle too), but stored the same way as google_client_secret for consistency."
  type        = string
}

variable "google_client_secret" {
  description = "Google OAuth client secret. Not currently used by the ID-token verification flow (no server-side code exchange happens), kept for potential future use (e.g. server-side refresh tokens)."
  type        = string
  sensitive   = true
}

variable "reef_stripe_secret_key" {
  description = "Stripe secret key for reef-site's own, separate Stripe account (distinct from travel_angels_stripe_secret_key) — used only when PaymentCheckoutSessionParams.Platform == \"reef\"."
  type        = string
  sensitive   = true
}

variable "reef_stripe_webhook_secret" {
  description = "Signing secret for the webhook endpoint registered under reef-site's Stripe account at https://api.unclaimedstreets.com/billing/stripe-webhook."
  type        = string
  sensitive   = true
}
