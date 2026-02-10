package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "patienttracker/db"
    "patienttracker/helpers"
)

type SigninRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var req SigninRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    var hash string
    var userType string
    var verified bool

    err := db.DB.QueryRow(`
        SELECT password_hash, user_type, is_verified
        FROM signup
        WHERE email = $1
    `, req.Email).Scan(&hash, &userType, &verified)

    if err == sql.ErrNoRows {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if !helpers.CheckPasswordHash(req.Password, hash) {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    if !verified {
        http.Error(w, "Email not verified", http.StatusForbidden)
        return
    }

    token, err := helpers.GenerateJWT(req.Email, userType)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "email": req.Email,
        "token": token,
    })
}
