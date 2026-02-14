package handlers

import (
	"encoding/json"
	"net/http"

	"onlineshop/db"
	"onlineshop/middleware"
)

type CreateProfileRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BirthPlace string `json:"birth_place"`
	BirthDate  string `json:"birth_date"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	NationalID string `json:"national_id"`
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {

	emailVal := r.Context().Value(middleware.ContextEmail)
	if emailVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	email := emailVal.(string)

	// resolve user id
	var userID int
	err := db.DB.QueryRow(
		`SELECT id FROM signup WHERE email=$1`,
		email,
	).Scan(&userID)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {

	// =======================
	// CHECK PROFILE EXISTS
	// =======================
	case http.MethodGet:

		var exists bool
		err := db.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM profiles WHERE user_id=$1
			)
		`, userID).Scan(&exists)

		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}

		if !exists {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		return

	// =======================
	// CREATE PROFILE
	// =======================
	case http.MethodPost:

		var req CreateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		var exists bool
		_ = db.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM profiles WHERE user_id=$1
			)
		`, userID).Scan(&exists)

		if exists {
			http.Error(w, "Profile already exists", http.StatusConflict)
			return
		}

		_, err = db.DB.Exec(`
			INSERT INTO profiles
			(user_id,first_name,last_name,birth_place,birth_date,address,phone,national_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`,
			userID,
			req.FirstName,
			req.LastName,
			req.BirthPlace,
			req.BirthDate,
			req.Address,
			req.Phone,
			req.NationalID,
		)

		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
