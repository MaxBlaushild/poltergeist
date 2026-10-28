terraform {
  required_version = ">= 1.5.0, < 2.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.25.0"
    }
  }
}

provider "aws" {
  region              = "us-east-1"
  allowed_account_ids = ["872408892710"]

  default_tags {
    tags = {
      Application = "travel-angels"
      Component   = "web"
      ManagedBy   = "terraform"
    }
  }
}
