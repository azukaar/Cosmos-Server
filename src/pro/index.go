package pro

import "github.com/gorilla/mux"

func IsPro() bool {
	return false
}

func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/groups", GroupsRoute)
	router.HandleFunc("/api/groups/{id}", GroupsIdRoute)
}
