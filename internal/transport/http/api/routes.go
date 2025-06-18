package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

func SetupRoutes(handler *Handler) *mux.Router {
	router := mux.NewRouter()
	
	// API v1
	apiV1 := router.PathPrefix("/api/v1").Subrouter()
	
	// Аутентификация (не требует авторизации)
	auth := apiV1.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", handler.Register).Methods("POST")
	auth.HandleFunc("/login", handler.Login).Methods("POST")
	auth.HandleFunc("/verify", handler.VerifyEmail).Methods("GET")
	
	// Защищенные маршруты (требуют авторизации)
	protected := apiV1.PathPrefix("/auth").Subrouter()
	protected.Use(mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(handler.AuthMiddleware(next.ServeHTTP))
	}))
	protected.HandleFunc("/me", handler.GetProfile).Methods("GET")
	protected.HandleFunc("/logout", handler.Logout).Methods("POST")
	
	// Управление пользователями (требуют авторизации)
	users := apiV1.PathPrefix("/users").Subrouter()
	users.Use(mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(handler.AuthMiddleware(next.ServeHTTP))
	}))
	users.HandleFunc("/{id}", handler.UpdateProfile).Methods("PATCH")
	
	return router
}
