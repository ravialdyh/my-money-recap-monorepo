terraform {
  required_version = ">= 1.12.0"

  required_providers {
    keycloak = {
      source  = "keycloak/keycloak"
      version = "5.9.0"
    }
  }
}

provider "keycloak" {}

variable "ssl_required" {
  description = "Keycloak's sslRequired for the mmr realm: none on the laptop, where every caller speaks plain HTTP, external on the machine. The provider defaults it to external, so this root never lets it default (ARCHITECTURE A13)."
  type        = string
}

variable "bff_client_secret" {
  description = "The bff client's secret. Explicit rather than generated, because a10-smoke.sh and the members wrong-audience test authenticate as that client and must know it in advance (A10)."
  type        = string
  sensitive   = true
}

variable "test_user_password" {
  description = "The realm test Member's password, the seam the Connect-boundary suite and the iOS UI test sign in with."
  type        = string
  sensitive   = true
}
