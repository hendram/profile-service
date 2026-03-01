package main

import (
    "log"
    "net/http"

    "onlineshop/handlers"
    "onlineshop/middleware"
    "onlineshop/db"
    "onlineshop/env"
)

func main() {
    // ✅ app-level init
    env.LoadEnv()
    db.InitDB()

    mux := http.NewServeMux()


mux.Handle(
	"/profile",
	middleware.JWTAuth(
		middleware.RequireRoles("customer", "seller")(
			http.HandlerFunc(handlers.ProfileHandler),
		),
	),
)

    log.Println("Server running on :80")
    log.Fatal(http.ListenAndServe(":80", middleware.CORS(mux)))
}
