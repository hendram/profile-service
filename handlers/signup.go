package handlers

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "strings"

    "patienttracker/db"
    "patienttracker/helpers"
)

type SignupRequest struct {
    Email    string   `json:"email"`
    Password string   `json:"password"`
    Roles    []string `json:"roles"` // optional
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    fail := func(stage string, err error) {
        log.Printf("[SIGNUP ERROR] stage=%s err=%v\n", stage, err)
        http.Error(w, "Signup failed", http.StatusBadRequest)
    }

    var req SignupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        fail("decode_json", err)
        return
    }

    if req.Email == "" || req.Password == "" {
        fail("missing_fields", nil)
        return
    }

    // timing attack mitigation
    _ = helpers.HashPassword(req.Password)

    if len(req.Roles) == 0 {
        req.Roles = []string{"customer"}
    }

    tx, err := db.DB.Begin()
    if err != nil {
        fail("begin_tx", err)
        return
    }

    defer func() {
        if err != nil {
            _ = tx.Rollback()
        }
    }()

    // insert user
    var userID int
    err = tx.QueryRow(`
        INSERT INTO signup (email, password_hash)
        VALUES ($1,$2)
        RETURNING id
    `,
        req.Email,
        helpers.HashPassword(req.Password),
    ).Scan(&userID)

    if err != nil {
        if strings.Contains(err.Error(), "duplicate") {
            fail("duplicate_email", err)
        } else {
            fail("insert_user", err)
        }
        return
    }

    // assign roles
    seen := map[string]bool{}

    for _, roleName := range req.Roles {

        if seen[roleName] {
            continue
        }
        seen[roleName] = true

        // prevent privilege escalation
        if roleName == "admin" || roleName == "superadmin" {
            fail("forbidden_role_attempt", nil)
            return
        }

        var roleID int
        err = tx.QueryRow(
            "SELECT id FROM roles WHERE name=$1",
            roleName,
        ).Scan(&roleID)

        if err == sql.ErrNoRows {
            fail("invalid_role", nil)
            return
        }
        if err != nil {
            fail("role_lookup", err)
            return
        }

        _, err = tx.Exec(`
            INSERT INTO user_roles (user_id, role_id)
            VALUES ($1,$2)
        `, userID, roleID)

        if err != nil {
            fail("insert_user_role", err)
            return
        }
    }

    // verification
    code := helpers.GenerateVerificationCode(req.Email)

    _, err = tx.Exec(`
        INSERT INTO email_verifications (signup_id, verification_code, expires_at)
        VALUES ($1,$2,NOW()+INTERVAL '10 minutes')
        ON CONFLICT (signup_id)
        DO UPDATE SET
            verification_code=EXCLUDED.verification_code,
            expires_at=EXCLUDED.expires_at,
            created_at=NOW()
    `, userID, code)

    if err != nil {
        fail("verification_insert", err)
        return
    }

    err = tx.Commit()
    if err != nil {
        fail("commit", err)
        return
    }

    // send email AFTER commit
    if err := helpers.SendVerificationEmail(req.Email, code); err != nil {
        log.Println("Email send fail:", err)
        http.Error(w, "Signup failed", http.StatusBadRequest)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "message": "Signup successful",
    })
}
