package handlers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"profile-service/db"
	"profile-service/middleware"
)

func CreateProfileHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("---- CreateProfileHandler hit ----")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	uid := r.Context().Value(middleware.ContextUserID)

	if uid == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var userID int

	switch v := uid.(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	default:
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req CreateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var photoBytes []byte

	if req.Photo != "" {
		photoBytes, err = base64.StdEncoding.DecodeString(req.Photo)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	_, err = db.DB.Exec(`
		INSERT INTO profiles
		(user_id,first_name,last_name,birth_place,birth_date,
		 address,phone,national_id,city,country,photo)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		userID,
		req.FirstName,
		req.LastName,
		req.BirthPlace,
		birthDate,
		req.Address,
		req.Phone,
		req.NationalID,
		req.City,
		req.Country,
		photoBytes,
	)

	if err != nil {
		log.Println("Insert failed:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "profile created",
	})
}

