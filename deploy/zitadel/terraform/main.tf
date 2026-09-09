terraform {
  required_version = ">= 1.10.0"

  backend "local" {
    path = "/state/terraform.tfstate"
  }

  required_providers {
    zitadel = {
      source  = "zitadel/zitadel"
      version = "3.6.0"
    }
  }
}

provider "zitadel" {
  domain           = "auth.localhost"
  port             = "8081"
  insecure         = true
  jwt_profile_file = "/bootstrap/terraform-admin.json"
}

variable "super_admin_password" {
  type      = string
  sensitive = true
}

variable "regular_user_password" {
  type      = string
  sensitive = true
}

resource "zitadel_org" "local" {
  name = "local-org"
}

resource "zitadel_project" "local" {
  name                   = "local-prj"
  org_id                 = zitadel_org.local.id
  project_role_assertion = true
  project_role_check     = true
  has_project_check      = true
}

resource "zitadel_application_oidc" "local" {
  project_id                  = zitadel_project.local.id
  org_id                      = zitadel_org.local.id
  name                        = "local-app"
  redirect_uris               = ["http://localhost:8080/auth/callback"]
  response_types              = ["OIDC_RESPONSE_TYPE_CODE"]
  grant_types                 = ["OIDC_GRANT_TYPE_AUTHORIZATION_CODE"]
  post_logout_redirect_uris   = ["http://localhost:8080/"]
  app_type                    = "OIDC_APP_TYPE_USER_AGENT"
  auth_method_type            = "OIDC_AUTH_METHOD_TYPE_NONE"
  version                     = "OIDC_VERSION_1_0"
  dev_mode                    = true
  access_token_type           = "OIDC_TOKEN_TYPE_JWT"
  access_token_role_assertion = true
  id_token_role_assertion     = true
  additional_origins          = ["http://localhost:8080"]
}

resource "zitadel_project_role" "super_admin" {
  org_id       = zitadel_org.local.id
  project_id   = zitadel_project.local.id
  role_key     = "super-admin"
  display_name = "Super administrator"
  group        = "local-access"
}

resource "zitadel_project_role" "regular_user" {
  org_id       = zitadel_org.local.id
  project_id   = zitadel_project.local.id
  role_key     = "regular-user"
  display_name = "Regular user"
  group        = "local-access"
}

resource "zitadel_human_user" "super_admin" {
  org_id                       = zitadel_org.local.id
  user_name                    = "super-admin@example.test"
  first_name                   = "Super"
  last_name                    = "Admin"
  display_name                 = "Local Super Admin"
  email                        = "super-admin@example.test"
  is_email_verified            = true
  initial_password             = var.super_admin_password
  initial_skip_password_change = true
}

resource "zitadel_human_user" "regular_user" {
  org_id                       = zitadel_org.local.id
  user_name                    = "regular-user@example.test"
  first_name                   = "Regular"
  last_name                    = "User"
  display_name                 = "Local Regular User"
  email                        = "regular-user@example.test"
  is_email_verified            = true
  initial_password             = var.regular_user_password
  initial_skip_password_change = true
}

resource "zitadel_org_member" "super_admin" {
  org_id  = zitadel_org.local.id
  user_id = zitadel_human_user.super_admin.id
  roles   = ["ORG_OWNER"]
}

resource "zitadel_user_grant" "super_admin" {
  org_id     = zitadel_org.local.id
  project_id = zitadel_project.local.id
  user_id    = zitadel_human_user.super_admin.id
  role_keys  = [zitadel_project_role.super_admin.role_key]
}

resource "zitadel_user_grant" "regular_user" {
  org_id     = zitadel_org.local.id
  project_id = zitadel_project.local.id
  user_id    = zitadel_human_user.regular_user.id
  role_keys  = [zitadel_project_role.regular_user.role_key]
}

output "app_env" {
  sensitive = true
  value     = <<-EOT
    ZITADEL_ISSUER=http://auth.localhost:8081
    ZITADEL_CLIENT_ID=${zitadel_application_oidc.local.client_id}
    APPLICATION_ORIGIN=http://localhost:8080
  EOT
}