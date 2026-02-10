package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "patienttracker/helpers"
    "patienttracker/db"
)

type SigninRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
    var req SigninRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    var hash string
    var usertype string
    var verified bool

    err := db.DB.QueryRow(`
        SELECT password_hash, usertype, is_verified
        FROM signup
        WHERE email = $1
    `, req.Email).Scan(&hash, &usertype, &verified)

    if err == sql.ErrNoRows ||
        !helpers.CheckPasswordHash(req.Password, hash) {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if !verified {
        http.Error(w, "Email not verified", http.StatusForbidden)
        return
    }

    token, err := helpers.GenerateJWT(req.Email, usertype)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "email": req.Email,
        "token": token,
    })
}

