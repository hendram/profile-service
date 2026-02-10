package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
        "patienttracker/db"
)

type AppointmentRequest struct {
	PatientName       string `json:"patient_name"`
	BirthPlace        string `json:"birth_place"`
	BirthDate         string `json:"birth_date"`         // YYYY-MM-DD
	Phone             string `json:"phone"`
	Email             string `json:"email"`
	Doctor            string `json:"doctor"`
	Clinic            string `json:"clinic"`              // REQUIRED
	CountyOfResidence string `json:"county_of_residence"` // OPTIONAL
	AppointmentDate   string `json:"appointment_date"`   // YYYY-MM-DD
	AppointmentTime   string `json:"appointment_time"`   // HH:MM
}

func createDoctorAppointment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req AppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Required field validation
	if req.Clinic == "" {
		http.Error(w, "clinic is required", http.StatusBadRequest)
		return
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		http.Error(w, "Invalid birth_date format", http.StatusBadRequest)
		return
	}

	appointmentDate, err := time.Parse("2006-01-02", req.AppointmentDate)
	if err != nil {
		http.Error(w, "Invalid appointment_date format", http.StatusBadRequest)
		return
	}

	appointmentTime, err := time.Parse("15:04", req.AppointmentTime)
	if err != nil {
		http.Error(w, "Invalid appointment_time format", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO appointments
		(
			patient_name,
			birth_place,
			birth_date,
			phone,
			email,
			doctor,
			clinic,
			county_of_residence,
			appointment_date,
			appointment_time
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var appointmentID int64

	err = db.DB.QueryRow(
		query,
		req.PatientName,
		req.BirthPlace,
		birthDate,
		req.Phone,
		req.Email,
		req.Doctor,
		req.Clinic,
		req.CountyOfResidence,
		appointmentDate,
		appointmentTime,
	).Scan(&appointmentID)

	if err != nil {
		log.Println("DB ERROR:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "appointment created",
		"appointment_id": appointmentID,
	})
}

