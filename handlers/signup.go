package handlers

import (
    "encoding/json"
    "log"
    "net/http"

    "patienttracker/db"
    "patienttracker/helpers"
)

type SignupRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    UserType string `json:"user_type"` // optional, defaults to patient
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var req SignupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if req.UserType == "" {
        req.UserType = "patient"
    }

    // Create signup row
    var signupID int
    err := db.DB.QueryRow(`
        INSERT INTO signup (email, password_hash, user_type)
        VALUES ($1, $2, $3)
        RETURNING id
    `,
        req.Email,
        helpers.HashPassword(req.Password),
        req.UserType,
    ).Scan(&signupID)

    if err != nil {
        http.Error(w, "Email already registered", http.StatusConflict)
        return
    }

    // Generate verification code
    code := helpers.GenerateVerificationCode(req.Email)

    // Store verification code (one active per signup)
    _, err = db.DB.Exec(`
        INSERT INTO email_verifications (signup_id, verification_code, expires_at)
        VALUES ($1, $2, NOW() + INTERVAL '10 minutes')
        ON CONFLICT (signup_id) DO UPDATE
        SET verification_code = EXCLUDED.verification_code,
            expires_at = EXCLUDED.expires_at,
            created_at = NOW()
    `, signupID, code)

    if err != nil {
        http.Error(w, "Failed to save verification code", http.StatusInternalServerError)
        return
    }

    // Send email
    if err := helpers.SendVerificationEmail(req.Email, code); err != nil {
        log.Println("SendVerificationEmail failed:", err)
        http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Signup successful. Verification code sent.",
    })
}
