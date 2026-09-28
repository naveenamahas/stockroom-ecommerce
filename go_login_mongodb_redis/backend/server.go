package main

import "net/http"

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/signup", signupHandler)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/profile", profileHandler)
	mux.HandleFunc("/api/logout", logoutHandler)
	mux.HandleFunc("/api/products", productsHandler)
	mux.HandleFunc("/api/products/", productsHandler)
	mux.HandleFunc("/api/shop/products", shopProductsHandler)
	mux.HandleFunc("/api/orders", ordersHandler)
	return cors(mux)
}
