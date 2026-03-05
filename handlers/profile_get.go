package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"profile-service/db"
	"profile-service/middleware"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("---- ProfileHandler hit ----")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	uid := r.Context().Value(middleware.ContextUserID)

	if uid == nil {
		log.Println("No user_id in context -> 401")
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
		log.Println("Invalid user_id type")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var (
		p     ProfileResponse
		birth time.Time
		photo []byte
	)

	err := db.DB.QueryRow(`
		SELECT first_name,last_name,birth_place,birth_date,
		       address,phone,national_id,city,country,photo
		FROM profiles
		WHERE user_id=$1
	`, userID).Scan(
		&p.FirstName,
		&p.LastName,
		&p.BirthPlace,
		&birth,
		&p.Address,
		&p.Phone,
		&p.NationalID,
		&p.City,
		&p.Country,
		&photo,
	)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "profile not found",
		})
		return
	}

	if err != nil {
		log.Println("DB error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p.BirthDate = birth.Format("2006-01-02")

	if len(photo) > 0 {
		p.Photo = base64.StdEncoding.EncodeToString(photo)
	}

	json.NewEncoder(w).Encode(p)
}

