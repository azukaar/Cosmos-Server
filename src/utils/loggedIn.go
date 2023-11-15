package utils

import (
	"errors"
	"net/http"
	"strconv"
)


func LoggedInOnlyWithRedirect(w http.ResponseWriter, req *http.Request) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	mfa, _ := strconv.Atoi(req.Header.Get("x-cosmos-mfa"))
	isUserLoggedIn := role > 0

	if !isUserLoggedIn || userNickname == "" {
		Error("LoggedInOnlyWithRedirect: User is not logged in", nil)
		http.Redirect(w, req, "/cosmos-ui/login?notlogged=1&redirect="+req.URL.Path, http.StatusFound)
		return errors.New("User not logged in")
	}

	if(mfa == 1) {
		http.Redirect(w, req, "/cosmos-ui/loginmfa?invalid=1&redirect=" + req.URL.Path + "&" + req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA")
	} else if(mfa == 2) {
		http.Redirect(w, req, "/cosmos-ui/newmfa?invalid=1&redirect=" + req.URL.Path + "&" + req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA Setup")
	}

	return nil
}

func AdminOnlyWithRedirect(w http.ResponseWriter, req *http.Request) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	mfa, _ := strconv.Atoi(req.Header.Get("x-cosmos-mfa"))
	isUserLoggedIn := role > 0
	isUserAdmin := role > 1

	if !isUserLoggedIn || userNickname == "" {
		Error("AdminLoggedInOnlyWithRedirect: User is not logged in", nil)
		http.Redirect(w, req, "/cosmos-ui/login?notlogged=1&redirect="+req.URL.Path, http.StatusFound)
		return errors.New("User is not logged")
	}

	if isUserLoggedIn && !isUserAdmin {
		Error("AdminLoggedInOnly: User is not Authorized", nil)
		HTTPError(w, "User not Authorized", http.StatusUnauthorized, "HTTP004")
		return errors.New("User is not Admin")
	}

	if(mfa == 1) {
		http.Redirect(w, req, "/cosmos-ui/loginmfa?invalid=1&redirect=" + req.URL.Path + "&" + req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA")
	} else if(mfa == 2) {
		http.Redirect(w, req, "/cosmos-ui/newmfa?invalid=1&redirect=" + req.URL.Path + "&" + req.URL.RawQuery, http.StatusTemporaryRedirect)
		return errors.New("User requires MFA Setup")
	}

	return nil
}

func LoggedInWeakOnly(w http.ResponseWriter, req *http.Request) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	isUserLoggedIn := role > 0

	if !isUserLoggedIn || userNickname == "" {
		Error("LoggedInOnly: User is not logged in", nil)
		HTTPError(w, "User not logged in", http.StatusUnauthorized, "HTTP004")
		return errors.New("User not logged in")
	}

	return nil
}

func IsLoggedIn(req *http.Request) bool {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	isUserLoggedIn := role > 0

	if !isUserLoggedIn || userNickname == "" {
		return false
	}

	return true
}

func LoggedInOnly(w http.ResponseWriter, req *http.Request) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	mfa, _ := strconv.Atoi(req.Header.Get("x-cosmos-mfa"))
	isUserLoggedIn := role > 0

	if !isUserLoggedIn || userNickname == "" {
		Error("LoggedInOnly: User is not logged in", nil)
		HTTPError(w, "User not logged in", http.StatusUnauthorized, "HTTP004")
		return errors.New("User not logged in")
	}

	if(mfa == 1) {
		HTTPError(w, "User not logged in (MFA)", http.StatusUnauthorized, "HTTP006")
		return errors.New("User requires MFA")
	} else if(mfa == 2) {
		HTTPError(w, "User requires MFA Setup", http.StatusUnauthorized, "HTTP007")
		return errors.New("User requires MFA Setup")
	}

	return nil
}

func AdminOnly(w http.ResponseWriter, req *http.Request) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	mfa, _ := strconv.Atoi(req.Header.Get("x-cosmos-mfa"))
	isUserLoggedIn := role > 0
	isUserAdmin := role > 1

	if !isUserLoggedIn || userNickname == "" {
		Error("AdminOnly: User is not logged in", nil)
		//http.Redirect(w, req, "/login?notlogged=1&redirect=" + req.URL.Path, http.StatusFound)
		HTTPError(w, "User not logged in", http.StatusUnauthorized, "HTTP004")
		return errors.New("User not logged in")
	}

	if isUserLoggedIn && !isUserAdmin {
		Error("AdminOnly: User is not admin", nil)
		HTTPError(w, "User unauthorized", http.StatusUnauthorized, "HTTP005")
		return errors.New("User not Admin")
	}

	if(mfa == 1) {
		HTTPError(w, "User not logged in (MFA)", http.StatusUnauthorized, "HTTP006")
		return errors.New("User requires MFA")
	} else if(mfa == 2) {
		HTTPError(w, "User requires MFA Setup", http.StatusUnauthorized, "HTTP007")
		return errors.New("User requires MFA Setup")
	}

	return nil
}

func IsAdmin(req *http.Request) bool {
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	return role > 1
}

func AdminOrItselfOnly(w http.ResponseWriter, req *http.Request, nickname string) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	mfa, _ := strconv.Atoi(req.Header.Get("x-cosmos-mfa"))
	isUserLoggedIn := role > 0
	isUserAdmin := role > 1

	if !isUserLoggedIn || userNickname == "" {
		Error("AdminOrItselfOnly: User is not logged in", nil)
		HTTPError(w, "User not logged in", http.StatusUnauthorized, "HTTP004")
		return errors.New("User not logged in")
	}

	if nickname != userNickname && !isUserAdmin {
		Error("AdminOrItselfOnly: User is not admin", nil)
		HTTPError(w, "User unauthorized", http.StatusUnauthorized, "HTTP005")
		return errors.New("User not Admin")
	}

	if(mfa == 1) {
		HTTPError(w, "User not logged in (MFA)", http.StatusUnauthorized, "HTTP006")
		return errors.New("User requires MFA")
	} else if(mfa == 2) {
		HTTPError(w, "User requires MFA Setup", http.StatusUnauthorized, "HTTP007")
		return errors.New("User requires MFA Setup")
	}

	return nil
}