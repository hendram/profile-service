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
        http.Error(w, "Verify failed", http.StatusBadRequest)
        return
    }

    if req.Code == "" {
        http.Error(w, "Verify failed", http.StatusBadRequest)
        return
    }

    var userID int
    var email string
    var expiresAt time.Time

    err := db.DB.QueryRow(`
        SELECT s.id, s.email, v.expires_at
        FROM signup s
        JOIN email_verifications v ON v.signup_id = s.id
        WHERE v.verification_code = $1
    `, req.Code).Scan(&userID, &email, &expiresAt)

    if err == sql.ErrNoRows {
        http.Error(w, "Verify failed", http.StatusUnauthorized)
        return
    }
    if err != nil {
        http.Error(w, "Verify failed", http.StatusInternalServerError)
        return
    }

    if time.Now().After(expiresAt) {
        http.Error(w, "Verify failed", http.StatusUnauthorized)
        return
    }

    tx, err := db.DB.Begin()
    if err != nil {
        http.Error(w, "Verify failed", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    _, err = tx.Exec(`
        UPDATE signup
        SET is_verified = TRUE
        WHERE id = $1
    `, userID)
    if err != nil {
        http.Error(w, "Verify failed", http.StatusInternalServerError)
        return
    }

    _, err = tx.Exec(`
        DELETE FROM email_verifications
        WHERE signup_id = $1
    `, userID)
    if err != nil {
        http.Error(w, "Verify failed", http.StatusInternalServerError)
        return
    }

    if err := tx.Commit(); err != nil {
        http.Error(w, "Verify failed", http.StatusInternalServerError)
        return
    }

    // fetch ONE role to activate session (frontend decides which one later if needed)
    var role string
    err = db.DB.QueryRow(`
        SELECT r.name
        FROM user_roles ur
        JOIN roles r ON r.id = ur.role_id
        WHERE ur.user_id = $1
        ORDER BY r.name
        LIMIT 1
    `, userID).Scan(&role)

    if err != nil {
        http.Error(w, "Verify failed", http.StatusInternalServerError)
        return
    }

    token, err := helpers.GenerateJWT(userID, email, role)
    if err != nil {
        http.Error(w, "Verify failed", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "email": email,
        "role":  role,
        "token": token,
    })
}
