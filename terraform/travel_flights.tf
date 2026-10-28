variable "flightaware_api_key_secret_arn" {
  description = "Optional ARN of an existing Secrets Manager secret containing a plain FlightAware AeroAPI key for Travel Angels flight lookup."
  type        = string
  default     = ""
  nullable    = false

  validation {
    condition     = var.flightaware_api_key_secret_arn == "" || can(regex("^arn:(aws|aws-us-gov|aws-cn):secretsmanager:[a-z0-9-]+:[0-9]{12}:secret:[A-Za-z0-9/_+=.@-]+$", var.flightaware_api_key_secret_arn))
    error_message = "Provide a full Secrets Manager secret ARN, or leave it empty to keep flight lookup disabled."
  }
}

locals {
  travel_angels_flight_secret_arns = var.flightaware_api_key_secret_arn == "" ? [] : [var.flightaware_api_key_secret_arn]
  travel_angels_flight_secrets = var.flightaware_api_key_secret_arn == "" ? [] : [{
    name      = "FLIGHTAWARE_API_KEY"
    valueFrom = var.flightaware_api_key_secret_arn
  }]
}
