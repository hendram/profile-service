package main

import (
    "log"
    "net/http"

    "profile-service/handlers"
    "profile-service/middleware"
    "profile-service/db"
    "profile-service/env"
)

func main() {
    // ✅ app-level init
    env.LoadEnv()
    db.InitDB()

    mux := http.NewServeMux()


mux.Handle("/profile",
	middleware.JWTAuth(
		middleware.RequireRoles("seller")(
			http.HandlerFunc(handlers.ProfileHandler),
		),
	),
)

mux.Handle("/profile/create",
	middleware.JWTAuth(
		middleware.RequireRoles("seller")(
			http.HandlerFunc(handlers.CreateProfileHandler),
		),
	),
)

    log.Println("Server running on :9000")
    log.Fatal(http.ListenAndServe(":9000", middleware.CORS(mux)))
}
