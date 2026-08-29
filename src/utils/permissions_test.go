package utils

import (
	"context"
	"net/http"
	"testing"
)

func reqWithPerms(perms []Permission) *http.Request {
	r, _ := http.NewRequest("GET", "/", nil)
	ctx := &AuthContext{Nickname: "u", Permissions: perms, IsSudoed: true}
	return r.WithContext(context.WithValue(r.Context(), AuthCtxKey, ctx))
}

func TestCanGrant(t *testing.T) {
	r := reqWithPerms([]Permission{PERM_LOGIN, PERM_CONFIGURATION})
	if !CanGrant(r, []Permission{PERM_LOGIN, PERM_LOGIN_WEAK, PERM_CONFIGURATION}) {
		t.Error("should grant held permissions (login ones are free)")
	}
	if CanGrant(r, []Permission{PERM_ADMIN}) {
		t.Error("must not grant PERM_ADMIN without holding it")
	}
	tok, _ := http.NewRequest("GET", "/", nil)
	tok = tok.WithContext(context.WithValue(tok.Context(), AuthCtxKey, &AuthContext{APIToken: &APITokenContext{Permissions: []Permission{PERM_ADMIN}}}))
	if CanGrant(tok, []Permission{PERM_USERS}) {
		t.Error("api token must not grant beyond its own permissions")
	}
	anon, _ := http.NewRequest("GET", "/", nil)
	if CanGrant(anon, []Permission{PERM_ADMIN_READ}) {
		t.Error("anonymous cannot grant anything")
	}
}

func TestValidateRolesChange(t *testing.T) {
	prev := map[Role]RoleConfig{}
	for k, v := range defaultRoles {
		prev[k] = v
	}
	// config-only user tries to give USER role PERM_ADMIN
	r := reqWithPerms([]Permission{PERM_LOGIN, PERM_CONFIGURATION})
	next := map[Role]RoleConfig{USER: {Name: "User", Permissions: []Permission{PERM_LOGIN, PERM_ADMIN}}}
	if err := ValidateRolesChange(r, prev, next); err == nil {
		t.Error("escalation via roles matrix must be rejected")
	}
	// same edit by a full admin is fine
	admin := reqWithPerms(defaultRoles[ADMIN].Permissions)
	if err := ValidateRolesChange(admin, prev, next); err != nil {
		t.Errorf("admin edit rejected: %v", err)
	}
	// brand new role compared against nothing: all perms are new grants
	next = map[Role]RoleConfig{Role(7): {Name: "Ops", Permissions: []Permission{PERM_RESOURCES}}}
	if err := ValidateRolesChange(r, prev, next); err == nil {
		t.Error("new role with unheld permission must be rejected")
	}
	// admin lockout guard
	next = map[Role]RoleConfig{ADMIN: {Name: "Admin", Permissions: []Permission{PERM_LOGIN, PERM_USERS, PERM_CONFIGURATION}}}
	if err := ValidateRolesChange(admin, prev, next); err == nil {
		t.Error("stripping PERM_ADMIN from Admin role must be rejected")
	}
}
