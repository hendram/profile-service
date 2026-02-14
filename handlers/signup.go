package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"onlineshop/db"
	"onlineshop/helpers"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserType string `json:"usertype"` // single string
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fail := func(stage string, err error) {
		if err != nil {
			log.Printf("[SIGNUP ERROR] stage=%s err=%v\n", stage, err)
		} else {
			log.Printf("[SIGNUP ERROR] stage=%s\n", stage)
		}
		http.Error(w, "Signup failed: "+stage, http.StatusBadRequest)
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[LOG] Invalid JSON")
		fail("Signup error", err)
		return
	}

	if req.Email == "" || req.Password == "" || req.UserType == "" {
		fail("Signup error", nil)
		return
	}

	log.Printf("[LOG] Signup request: email=%s, usertype=%s\n", req.Email, req.UserType)

	if req.UserType == "admin" || req.UserType == "superadmin" {
		fail("Signup error", nil)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		fail("Signup error", err)
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var userID int
	err = tx.QueryRow(`SELECT id FROM signup WHERE email=$1`, req.Email).Scan(&userID)
	if err != nil && err != sql.ErrNoRows {
		fail("Signup error", err)
		return
	}

	isNewUser := err == sql.ErrNoRows
	if isNewUser {
		err = tx.QueryRow(`
			INSERT INTO signup (email, password_hash)
			VALUES ($1,$2)
			RETURNING id
		`, req.Email, helpers.HashPassword(req.Password)).Scan(&userID)
		if err != nil {
			fail("Signup error", err)
			return
		}
		log.Printf("[LOG] New user inserted: userID=%d\n", userID)
	} else {
		log.Printf("[LOG] Existing user found: userID=%d\n", userID)
	}

	// Check if user already has this role
	var roleID int
	err = tx.QueryRow(`SELECT id FROM roles WHERE name=$1`, req.UserType).Scan(&roleID)
	if err != nil {
		fail("Signup error", err)
		return
	}

	var exists bool
	if !isNewUser {
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM user_roles WHERE user_id=$1 AND role_id=$2
			)
		`, userID, roleID).Scan(&exists)
		if err != nil {
			fail("Signup error", err)
			return
		}
		if exists {
			fail("Signup error", nil)
			return
		}
	}

	// Assign role
	_, err = tx.Exec(`
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1,$2)
		ON CONFLICT DO NOTHING
	`, userID, roleID)
	if err != nil {
		fail("Signup error", err)
		return
	}
	log.Printf("[LOG] Role %s assigned to user %d\n", req.UserType, userID)

	// Verification code
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
		fail("Signup error", err)
		return
	}

	if err = tx.Commit(); err != nil {
		fail("Signup error", err)
		return
	}

	if err := helpers.SendVerificationEmail(req.Email, code); err != nil {
		fail("Signup error", err)
		return
	}

	log.Printf("[LOG] Verification email sent to %s\n", req.Email)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Signup successful, verification email sent",
	})
}
