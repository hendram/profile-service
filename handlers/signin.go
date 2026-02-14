package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"onlineshop/db"
	"onlineshop/helpers"
)

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserType string `json:"usertype"`
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fail := func(stage string, err error) {
		log.Printf("[SIGNIN ERROR] stage=%s err=%v\n", stage, err)
		http.Error(w, "Signin failed", http.StatusUnauthorized)
	}

	var req SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail("decode_json", err)
		return
	}

	if req.Email == "" || req.Password == "" || req.UserType == "" {
		fail("missing_fields", nil)
		return
	}

	var userID int
	var hash string
	var verified bool

	err := db.DB.QueryRow(`
		SELECT id,password_hash,is_verified
		FROM signup
		WHERE email=$1
	`, req.Email).Scan(&userID, &hash, &verified)

	if err == sql.ErrNoRows {
		fail("email_not_found", err)
		return
	}
	if err != nil {
		fail("query_user", err)
		return
	}

	if !helpers.CheckPasswordHash(req.Password, hash) {
		fail("wrong_password", nil)
		return
	}

	if !verified {
		fail("not_verified", nil)
		return
	}

	// verify user actually has this role
	var exists bool
	err = db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id=$1 AND r.name=$2
		)
	`, userID, req.UserType).Scan(&exists)

	if err != nil {
		fail("role_check", err)
		return
	}

	if !exists {
		fail("role_not_assigned", nil)
		return
	}

	token, err := helpers.GenerateJWT(userID, req.Email, req.UserType)
	if err != nil {
		fail("jwt_generation", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"email": req.Email,
		"role":  req.UserType,
		"token": token,
	})
}
