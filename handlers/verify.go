package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "time"

    "patienttracker/db"
    "patienttracker/helpers"
)

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var req struct {
        Code string `json:"code"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if req.Code == "" {
        http.Error(w, "Verification code required", http.StatusBadRequest)
        return
    }

    var signupID int
    var email string
    var userType string
    var expiresAt time.Time

    err := db.DB.QueryRow(`
        SELECT s.id, s.email, s.user_type, v.expires_at
        FROM signup s
        JOIN email_verifications v ON s.id = v.signup_id
        WHERE v.verification_code = $1
    `, req.Code).Scan(&signupID, &email, &userType, &expiresAt)

    if err == sql.ErrNoRows {
        http.Error(w, "Invalid verification code", http.StatusUnauthorized)
        return
    }

    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if time.Now().After(expiresAt) {
        http.Error(w, "Verification code expired", http.StatusUnauthorized)
        return
    }

    tx, err := db.DB.Begin()
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    _, err = tx.Exec(`
        UPDATE signup
        SET is_verified = TRUE
        WHERE id = $1
    `, signupID)
    if err != nil {
        http.Error(w, "Failed to verify user", http.StatusInternalServerError)
        return
    }

    _, err = tx.Exec(`
        DELETE FROM email_verifications
        WHERE signup_id = $1
    `, signupID)
    if err != nil {
        http.Error(w, "Failed to clean verification code", http.StatusInternalServerError)
        return
    }

    if err := tx.Commit(); err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    token, err := helpers.GenerateJWT(email, userType)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "email": email,
        "token": token,
    })
}
