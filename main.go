package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
)

// SignupRequest JSON
type SignupRequest struct {
    Firstname  string `json:"firstname"`
    Lastname   string `json:"lastname"`
    NationalID string `json:"national_id"`
    Phone      string `json:"phone"`
    Email      string `json:"email"`
    Password   string `json:"password"`
    Usertype   string `json:"usertype"` // "patient" or "doctor"
}

// SignupResponse JSON
type SignupResponse struct {
    Message string `json:"message"`
}

// SigninRequest JSON
type SigninRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// SigninResponse JSON
type SigninResponse struct {
    Email   string `json:"email"`
    Message string `json:"message"`
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req SignupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // validate usertype
    if req.Usertype != "patient" && req.Usertype != "doctor" {
        http.Error(w, "Invalid user type", http.StatusBadRequest)
        return
    }

    // check if email already exists
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM userprofile WHERE email=$1)", req.Email).Scan(&exists)
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    if exists {
        http.Error(w, "Email already registered", http.StatusConflict)
        return
    }

    // insert user
    _, err = db.Exec(
        `INSERT INTO userprofile (firstname, lastname, national_id, phone, email, password, usertype)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        req.Firstname, req.Lastname, req.NationalID, req.Phone, req.Email, req.Password, req.Usertype,
    )
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    resp := SignupResponse{Message: "Signup successful"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req SigninRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // check email and password
    var password string
    err := db.QueryRow("SELECT password FROM userprofile WHERE email=$1", req.Email).Scan(&password)
    if err == sql.ErrNoRows {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    } else if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if password != req.Password {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    resp := SigninResponse{
        Email:   req.Email,
        Message: "Login successful",
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func corsMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

                // Preflight request
                if r.Method == http.MethodOptions {
                        w.WriteHeader(http.StatusNoContent)
                        return
                }

                next.ServeHTTP(w, r)
        })
}




func main() {
    initDB()

    mux := http.NewServeMux()
    mux.HandleFunc("/signup", signupHandler)
    mux.HandleFunc("/signin", signinHandler)
    mux.HandleFunc("/doctorappointment", createDoctorAppointment)

    log.Println("Server running on :80")
    log.Fatal(http.ListenAndServe(":80", corsMiddleware(mux)))
}
