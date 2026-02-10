package main

import (
    "log"
    "net/http"

    "patienttracker/handlers"
    "patienttracker/middleware"
    "patienttracker/db"
)

func main() {
    // ✅ app-level init
    LoadEnv()
    db.InitDB()

    mux := http.NewServeMux()

    mux.HandleFunc("/signup", handlers.SignupHandler)
    mux.HandleFunc("/verify", handlers.VerifyHandler)
    mux.HandleFunc("/signin", handlers.SigninHandler)

    log.Println("Server running on :80")
    log.Fatal(http.ListenAndServe(":80", middleware.CORS(mux)))
}
