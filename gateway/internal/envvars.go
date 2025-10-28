/*
Copyright 2024 BlackRock, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

type EnvVars struct {
	HttpPort               string `env:"HTTP_PORT" envDefault:"8005"`
	PrimaryTargetPort      string `env:"PRIMARY_TARGET_PORT" envDefault:"5000"`
	SecondaryTargetPattern string `env:"SECONDARY_TARGET_PATTERN" envDefault:""`
	SecondaryTargetPort    string `env:"SECONDARY_TARGET_PORT" envDefault:""`
	MetricsEnabled         bool   `env:"METRICS_ENABLED" envDefault:"true"`
	MetricsTlsEnabled      bool   `env:"METRICS_TLS_ENABLED" envDefault:"false"`
	MetricsPort            string `env:"METRICS_PORT"` // Required if METRICS_ENABLED is true
	MetricsPath            string `env:"METRICS_PATH"` // Required if METRICS_ENABLED is true
	URLSubPath             string `env:"URL_SUB_PATH" envDefault:"/"`

	// OIDC Configuration
	OIDCEnabled      bool   `env:"OIDC_ENABLED" envDefault:"false"`
	OIDCIssuerURL    string `env:"OIDC_ISSUER_URL"`    // Required if OIDC_ENABLED is true
	OIDCClientID     string `env:"OIDC_CLIENT_ID"`     // Required if OIDC_ENABLED is true
	OIDCClientSecret string `env:"OIDC_CLIENT_SECRET"` // Required if OIDC_ENABLED is true
	OIDCRedirectURL  string `env:"OIDC_REDIRECT_URL"`  // Required if OIDC_ENABLED is true
	OIDCScopes       string `env:"OIDC_SCOPES" envDefault:"openid,profile,email"`

	// Authorization Configuration
	AuthzEnabled        bool   `env:"AUTHZ_ENABLED" envDefault:"false"`     // Enable role-based authorization
	AuthzRequiredRoles  string `env:"AUTHZ_REQUIRED_ROLES" envDefault:""`   // Comma-separated list of required roles
	AuthzRequiredScopes string `env:"AUTHZ_REQUIRED_SCOPES" envDefault:""`  // Comma-separated list of required OAuth scopes
	AuthzRoleClaim      string `env:"AUTHZ_ROLE_CLAIM" envDefault:"roles"`  // JWT claim containing user roles
	AuthzScopeClaim     string `env:"AUTHZ_SCOPE_CLAIM" envDefault:"scope"` // JWT claim containing OAuth scopes
	AuthzAdminRoles     string `env:"AUTHZ_ADMIN_ROLES" envDefault:"admin"` // Comma-separated list of admin roles (bypass all checks)
}
