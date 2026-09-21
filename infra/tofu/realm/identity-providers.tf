variable "google_client_id" {
  description = "The Google OAuth client's id, google_client_id in the secrets file (A13). A plain variable because Keycloak's vault resolves only the client secret."
  type        = string
}

variable "apple_services_id" {
  description = "The Sign in with Apple Services ID, apple_services_id in the secrets file: the client id the Apple extension signs its client-secret JWT for (PRD §6.1)."
  type        = string
}

variable "apple_team_id" {
  description = "The Apple Developer Team ID, apple_team_id in the secrets file: the JWT's issuer."
  type        = string
}

variable "apple_sign_in_key_id" {
  description = "The Key ID of the Sign in with Apple .p8, apple_sign_in_key_id in the secrets file: the JWT's kid."
  type        = string
}

resource "keycloak_oidc_google_identity_provider" "google" {
  realm         = keycloak_realm.mmr.id
  client_id     = var.google_client_id
  client_secret = "$${vault.google_client_secret}"
  trust_email   = true
}

resource "keycloak_oidc_identity_provider" "apple" {
  realm             = keycloak_realm.mmr.id
  alias             = "apple"
  provider_id       = "apple"
  display_name      = "Sign in with Apple"
  client_id         = var.apple_services_id
  client_secret     = "$${vault.apple_sign_in_key}"
  authorization_url = "https://appleid.apple.com/auth/authorize?response_mode=form_post"
  token_url         = "https://appleid.apple.com/auth/token"
  default_scopes    = "openid name email"

  extra_config = {
    teamId = var.apple_team_id
    keyId  = var.apple_sign_in_key_id
  }
}
