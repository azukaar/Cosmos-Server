package utils

import (
	"errors"
	"fmt"
	"net/http"
)

func GetAuthContext(req *http.Request) *AuthContext {
	ctx, _ := req.Context().Value(AuthCtxKey).(*AuthContext)
	if ctx == nil {
		return &AuthContext{}
	}
	return ctx
}

func LoggedInOnlyWithRedirect(w http.ResponseWriter, req *http.Request) error {
	ctx := GetAuthContext(req)

	if ctx.Nickname == "" {
		Error("LoggedInOnlyWithRedirect: User is not logged in", nil)
		http.Redirect(w, req, "/cosmos-ui/login?notlogged=1&redirect="+req.URL.Path, http.StatusFound)
		return errors.New("User not logged in")
	}

	if ctx.MFAState == 1 {
		http.Redirect(w, req, "/cosmos-ui/loginmfa?invalid=1&redirect="+req.URL.Path+"&"+req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA")
	} else if ctx.MFAState == 2 {
		http.Redirect(w, req, "/cosmos-ui/newmfa?invalid=1&redirect="+req.URL.Path+"&"+req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA Setup")
	}

	return nil
}

func AdminOnlyWithRedirect(w http.ResponseWriter, req *http.Request) error {
	ctx := GetAuthContext(req)

	if ctx.Nickname == "" {
		Error("AdminLoggedInOnlyWithRedirect: User is not logged in", nil)
		http.Redirect(w, req, "/cosmos-ui/login?notlogged=1&redirect="+req.URL.Path, http.StatusFound)
		return errors.New("User is not logged")
	}

	if !HasPermission(req, PERM_ADMIN_READ) {
		Error("AdminLoggedInOnly: User is not Authorized", nil)
		HTTPError(w, "User not Authorized", http.StatusUnauthorized, "HTTP004")
		return errors.New("User is not Admin")
	}

	if ctx.MFAState == 1 {
		http.Redirect(w, req, "/cosmos-ui/loginmfa?invalid=1&redirect="+req.URL.Path+"&"+req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA")
	} else if ctx.MFAState == 2 {
		http.Redirect(w, req, "/cosmos-ui/newmfa?invalid=1&redirect="+req.URL.Path+"&"+req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA Setup")
	}

	return nil
}

func IsLoggedIn(req *http.Request) bool {
	return GetAuthContext(req).Nickname != ""
}

func containsPermission(perms []Permission, p Permission) bool {
	for _, perm := range perms {
		if perm == p {
			return true
		}
	}
	return false
}

func RoleHasSudoPermissions(role Role) bool {
	perms := GetRolePermissions(role)
	if perms == nil {
		return false
	}
	return PermissionsHaveSudo(perms)
}

func PermissionsHaveSudo(perms []Permission) bool {
	for _, p := range perms {
		if containsPermission(SudoPermissions, p) {
			return true
		}
	}
	return false
}

func apiTokenHasPermission(token *APITokenContext, permission Permission) bool {
	return containsPermission(token.Permissions, permission)
}

func CheckPermissions(w http.ResponseWriter, req *http.Request, permission Permission) error {
	ctx := GetAuthContext(req)

	// API token path
	if ctx.APIToken != nil {
		if permission == PERM_LOGIN {
			return nil
		}
		// Check token has the specific permission
		if !apiTokenHasPermission(ctx.APIToken, permission) {
			Error("CheckPermissions: API token lacks permission: "+PermissionLabel(permission), nil)
			HTTPError(w, fmt.Sprintf("API token missing permission: %s", PermissionLabel(permission)), http.StatusForbidden, "AT003")
			return errors.New("API token lacks permission")
		}
		return nil
	}

	// JWT/session path

	// Step 1: Must be logged in
	if ctx.Nickname == "" {
		Error("CheckPermissions: User is not logged in", nil)
		HTTPError(w, "User not logged in", http.StatusUnauthorized, "HTTP004")
		return errors.New("User not logged in")
	}

	// Step 2: Check permission is in user's permission set
	if !containsPermission(ctx.Permissions, permission) {
		Error("CheckPermissions: User lacks permission: "+PermissionLabel(permission), nil)
		HTTPError(w, fmt.Sprintf("Missing permission: %s", PermissionLabel(permission)),
			http.StatusUnauthorized, "HTTP005")
		return errors.New("User lacks permission")
	}

	// Step 3: If permission requires sudo, check that user is sudoed
	if containsPermission(SudoPermissions, permission) && !ctx.IsSudoed {
		Error("CheckPermissions: Sudo required for: "+PermissionLabel(permission), nil)
		HTTPError(w, fmt.Sprintf("Sudo required for: %s", PermissionLabel(permission)),
			http.StatusUnauthorized, "HTTP008")
		return errors.New("Sudo required")
	}

	// Step 4: MFA enforcement (skip for PERM_LOGIN_WEAK — used by 2FA flow)
	if permission != PERM_LOGIN_WEAK {
		if ctx.MFAState == 1 {
			HTTPError(w, "User not logged in (MFA)", http.StatusUnauthorized, "HTTP006")
			return errors.New("User requires MFA")
		} else if ctx.MFAState == 2 {
			HTTPError(w, "User requires MFA Setup", http.StatusUnauthorized, "HTTP007")
			return errors.New("User requires MFA Setup")
		}
	}

	return nil
}

// CheckPermissionsOrSelf allows access if the user is acting on themselves (nickname matches),
// even without the specified permission. Otherwise falls back to standard permission check.
func CheckPermissionsOrSelf(w http.ResponseWriter, req *http.Request, nickname string, permission Permission) error {
	ctx := GetAuthContext(req)

	// API token path — no self-access concept, check permission directly
	if ctx.APIToken != nil {
		return CheckPermissions(w, req, permission)
	}

	// Must be logged in
	if ctx.Nickname == "" {
		Error("CheckPermissionsOrSelf: User is not logged in", nil)
		HTTPError(w, "User not logged in", http.StatusUnauthorized, "HTTP004")
		return errors.New("User not logged in")
	}

	// Self-access: if nickname matches the logged-in user, skip permission/sudo checks
	if nickname != "" && nickname == ctx.Nickname {
		// Still enforce MFA
		if permission != PERM_LOGIN_WEAK {
			if ctx.MFAState == 1 {
				HTTPError(w, "User not logged in (MFA)", http.StatusUnauthorized, "HTTP006")
				return errors.New("User requires MFA")
			} else if ctx.MFAState == 2 {
				HTTPError(w, "User requires MFA Setup", http.StatusUnauthorized, "HTTP007")
				return errors.New("User requires MFA Setup")
			}
		}
		return nil
	}

	// Not self-access — require the permission
	return CheckPermissions(w, req, permission)
}

func HasPermission(req *http.Request, permission Permission) bool {
	ctx := GetAuthContext(req)

	if ctx.APIToken != nil {
		if permission == PERM_LOGIN || permission == PERM_LOGIN_WEAK {
			return true
		}
		return apiTokenHasPermission(ctx.APIToken, permission)
	}

	if ctx.Nickname == "" {
		return false
	}

	if !containsPermission(ctx.Permissions, permission) {
		return false
	}

	if containsPermission(SudoPermissions, permission) {
		return ctx.IsSudoed
	}

	return true
}
