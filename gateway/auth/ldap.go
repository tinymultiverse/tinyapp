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

package auth

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/tinymultiverse/tinyapp/gateway/internal"
	"go.uber.org/zap"
)

type LDAPAuthenticator struct {
	config internal.EnvVars
}

func NewLDAPAuthenticator(config internal.EnvVars) *LDAPAuthenticator {
	return &LDAPAuthenticator{
		config: config,
	}
}

// Authenticate performs only LDAP authentication without authorization
func (la *LDAPAuthenticator) Authenticate(req *http.Request) (string, error) {
	if !la.config.LdapEnabled {
		return "", nil
	}

	username, password, err := la.extractCredentials(req)
	if err != nil {
		return "", err
	}

	return la.authenticateLDAP(username, password)
}

// AuthorizeUser checks if the authenticated user is in the allowed users list
func (la *LDAPAuthenticator) AuthorizeUser(username string) error {
	return la.authorizeUser(username)
}

func (la *LDAPAuthenticator) extractCredentials(req *http.Request) (string, string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", fmt.Errorf("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", fmt.Errorf("unsupported authorization type")
	}

	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("invalid base64 encoding: %w", err)
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return "", "", fmt.Errorf("invalid credentials format")
	}

	return credentials[0], credentials[1], nil
}

// authenticateLDAP performs LDAP authentication
func (la *LDAPAuthenticator) authenticateLDAP(username, password string) (string, error) {
	conn, err := la.connectLDAP()
	if err != nil {
		return "", fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer conn.Close()

	if la.config.LdapBindDN != "" {
		err = conn.Bind(la.config.LdapBindDN, la.config.LdapBindPassword)
		if err != nil {
			zap.S().Errorw("failed to bind with service account", "error", err)
			return "", fmt.Errorf("LDAP service account bind failed")
		}
	}

	userDN, err := la.searchUser(conn, username)
	if err != nil {
		return "", fmt.Errorf("user search failed: %w", err)
	}

	// Authenticate user by binding with their credentials
	err = conn.Bind(userDN, password)
	if err != nil {
		zap.S().Debugw("user authentication failed", "username", username, "error", err)
		return "", fmt.Errorf("authentication failed")
	}

	zap.S().Infow("user authenticated successfully", "username", username)
	return username, nil
}

// connectLDAP establishes connection to LDAP server
func (la *LDAPAuthenticator) connectLDAP() (*ldap.Conn, error) {
	var scheme string
	if la.config.LdapTLS {
		scheme = "ldaps"
	} else {
		scheme = "ldap"
	}

	ldapURL := fmt.Sprintf("%s://%s:%d", scheme, la.config.LdapServer, la.config.LdapPort)
	fmt.Println("Connecting to LDAP server at", ldapURL)

	var conn *ldap.Conn
	var err error

	if la.config.LdapTLS {
		tlsConfig := &tls.Config{
			ServerName: la.config.LdapServer,
		}
		conn, err = ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(tlsConfig))
	} else {
		conn, err = ldap.DialURL(ldapURL)
	}

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// searchUser searches for user in LDAP directory
func (la *LDAPAuthenticator) searchUser(conn *ldap.Conn, username string) (string, error) {
	filter := fmt.Sprintf("(&(%s=%s)%s)",
		la.config.LdapUserAttribute,
		ldap.EscapeFilter(username),
		la.config.LdapUserFilter)

	searchRequest := ldap.NewSearchRequest(
		la.config.LdapBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		la.config.LdapSearchSizeLimit,
		la.config.LdapSearchTimeLimit,
		false,
		filter,
		la.config.LdapReturnAttributes,
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return "", err
	}

	if len(result.Entries) == 0 {
		return "", fmt.Errorf("user not found")
	}

	if len(result.Entries) > 1 {
		return "", fmt.Errorf("multiple users found")
	}

	return result.Entries[0].DN, nil
}

// RequireAuth is a middleware that sends 401 with WWW-Authenticate header
func (la *LDAPAuthenticator) RequireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="LDAP Authentication"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

// authorizeUser checks if the authenticated user is in the allowed users list
func (la *LDAPAuthenticator) authorizeUser(username string) error {
	if !la.config.AuthorizationEnabled {
		return nil
	}

	if len(la.config.AllowedUsers) == 0 {
		return nil
	}

	for _, allowedUser := range la.config.AllowedUsers {
		if strings.TrimSpace(allowedUser) == username {
			zap.S().Debugw("user authorized", "username", username)
			return nil
		}
	}

	zap.S().Warnw("user not in allowed users list", "username", username, "allowedUsers", la.config.AllowedUsers)
	return fmt.Errorf("user not authorized")
}
