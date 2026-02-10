package handlers

import (
    "encoding/json"
    "net/http"

    "patienttracker/db"
    "patienttracker/helpers"
)

type SignupRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
    var req SignupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    _, err := db.DB.Exec(`
        INSERT INTO signup (email, password_hash)
        VALUES ($1, $2)
    `, req.Email, helpers.HashPassword(req.Password))

    if err != nil {
        http.Error(w, "Email already registered", http.StatusConflict)
        return
    }

    code := helpers.GenerateVerificationCode(req.Email)

    if err := helpers.SendVerificationEmail(req.Email, code); err != nil {
        http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
