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

type ProfileResponse struct {
        FirstName  string `json:"first_name"`
        LastName   string `json:"last_name"`
        BirthPlace string `json:"birth_place"`
        BirthDate  string `json:"birth_date"`
        Address    string `json:"address"`
        Phone      string `json:"phone"`
        NationalID string `json:"national_id"`
        City       string `json:"city"`
        Country    string `json:"country"`
        Photo      string `json:"photo,omitempty"`
}

type CreateProfileRequest struct {
        FirstName  string `json:"first_name"`
        LastName   string `json:"last_name"`
        BirthPlace string `json:"birth_place"`
        BirthDate  string `json:"birth_date"`
        Address    string `json:"address"`
        Phone      string `json:"phone"`
        NationalID string `json:"national_id"`
        City       string `json:"city"`
        Country    string `json:"country"`
        Photo      string `json:"photo"`
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {

        log.Println("---- ProfileHandler hit ----")
        log.Println("Method:", r.Method)

        w.Header().Set("Content-Type", "application/json")

        emailVal := r.Context().Value(middleware.ContextEmail)
        log.Println("Context email:", emailVal)

        if emailVal == nil {
                log.Println("No email in context -> 401")
                w.WriteHeader(http.StatusUnauthorized)
                return
        }

        var userID int
        err := db.DB.QueryRow(
                `SELECT id FROM signup WHERE email=$1`,
                emailVal.(string),
        ).Scan(&userID)

        if err != nil {
                log.Println("User lookup failed:", err)
                w.WriteHeader(http.StatusUnauthorized)
                return
        }

        log.Println("Resolved userID:", userID)

        switch r.Method {

        // =====================
        // GET PROFILE
        // =====================
        case http.MethodGet:

                log.Println("GET /profile called")

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
                        log.Println("Profile NOT FOUND for user:", userID)
                        w.WriteHeader(http.StatusNotFound)
                        return
                }
                if err != nil {
                        log.Println("Profile query error:", err)
                        w.WriteHeader(http.StatusInternalServerError)
                        return
                }

                log.Println("Profile FOUND for user:", userID)

                p.BirthDate = birth.Format("2006-01-02")
                if len(photo) > 0 {
                        log.Println("Photo exists, encoding base64")
                        p.Photo = base64.StdEncoding.EncodeToString(photo)
                } else {
                        log.Println("No photo stored")
                }

                w.WriteHeader(http.StatusOK)
                json.NewEncoder(w).Encode(p)

                log.Println("Profile response sent OK")
                return

        // =====================
        // CREATE PROFILE
        // =====================
        case http.MethodPost:

                log.Println("POST /profile called")

                var req CreateProfileRequest
                if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                        log.Println("JSON decode error:", err)
                        w.WriteHeader(http.StatusBadRequest)
                        return
                }

                log.Println("Decoded profile request for:", req.FirstName, req.LastName)

                birthDate, err := time.Parse("2006-01-02", req.BirthDate)
                if err != nil {
                        log.Println("Birthdate parse error:", err)
                        w.WriteHeader(http.StatusBadRequest)
                        return
                }

                var photoBytes []byte
                if req.Photo != "" {
                        log.Println("Decoding photo base64")
                        photoBytes, err = base64.StdEncoding.DecodeString(req.Photo)
                        if err != nil {
                                log.Println("Photo decode failed:", err)
                                w.WriteHeader(http.StatusBadRequest)
                                return
                        }
                } else {
                        log.Println("No photo provided")
                }

                log.Println("Inserting profile for user:", userID)

                _, err = db.DB.Exec(`
                        INSERT INTO profiles
                        (user_id,first_name,last_name,birth_place,birth_date,
                         address,phone,national_id,city,country,photo)
                        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
                        ON CONFLICT (user_id) DO NOTHING
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

                log.Println("Profile inserted (or already exists)")

                w.WriteHeader(http.StatusCreated)
                return

        default:
                log.Println("Invalid method:", r.Method)
                w.WriteHeader(http.StatusMethodNotAllowed)
        }
}
