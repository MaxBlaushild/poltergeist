variable "travel_angels_web_url" {
  description = "Public HTTPS origin for Travel Angels share links and guest email delivery. Empty omits the runtime environment variable."
  type        = string
  default     = "https://travel-angels.org"

  validation {
    condition     = var.travel_angels_web_url == "" || can(regex("^https://[A-Za-z0-9][A-Za-z0-9.-]*(:[0-9]+)?/?$", var.travel_angels_web_url))
    error_message = "travel_angels_web_url must be empty or an HTTPS origin without a path, query, or fragment."
  }
}
