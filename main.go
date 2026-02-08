package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
        "time"
    "github.com/golang-jwt/jwt/v5"
    "strings"
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
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    _, err := db.Exec(`
        INSERT INTO signup (email, password_hash)
        VALUES ($1, $2)
    `, req.Email, hashPassword(req.Password))

    if err != nil {
        http.Error(w, "Email already registered", http.StatusConflict)
        return
    }

    code := generateVerificationCode(req.Email)
    // sendEmail(req.Email, code)

    w.WriteHeader(http.StatusOK)
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Code string `json:"code"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    decoded, err := base64.StdEncoding.DecodeString(req.Code)
    if err != nil {
        http.Error(w, "Invalid code", http.StatusUnauthorized)
        return
    }

    parts := strings.Split(string(decoded), "|")
    if len(parts) != 3 {
        http.Error(w, "Invalid code format", http.StatusUnauthorized)
        return
    }

    email := parts[0]
    tsStr := parts[1]
    sigHex := parts[2]

    ts, err := strconv.ParseInt(tsStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid timestamp", http.StatusUnauthorized)
        return
    }

    // ⏱️ expiry check
    if time.Now().Unix()-ts > 300 {
        http.Error(w, "Code expired", http.StatusUnauthorized)
        return
    }

    // 🔐 verify signature
    secret := []byte(os.Getenv("VERIFY_SECRET"))
    payload := fmt.Sprintf("%s|%s", email, tsStr)

    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(payload))
    expectedSig := fmt.Sprintf("%x", mac.Sum(nil))

    if !hmac.Equal([]byte(sigHex), []byte(expectedSig)) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // ✅ mark verified
    _, err = db.Exec(`
        UPDATE signup
        SET is_verified = TRUE
        WHERE email = $1
    `, email)

    if err != nil {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    w.WriteHeader(http.StatusOK)
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

    var password string
    var usertype string

    err := db.QueryRow(
        "SELECT password, usertype FROM userprofile WHERE email=$1",
        req.Email,
    ).Scan(&password, &usertype)

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

    token, err := generateJWT(req.Email, usertype)
    if err != nil {
        http.Error(w, "Token generation failed", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "email":   req.Email,
        "token":   token,
        "message": "Login successful",
    })
}


var jwtSecret = []byte(">>><<<")

func generateJWT(email string, usertype string) (string, error) {
    
claims := jwt.MapClaims{
        "email":    email,
        "exp":      time.Now().Add(24 * time.Hour).Unix(),
        "iat":      time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}


func corsMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set(
            "Access-Control-Allow-Headers",
            "Content-Type, Authorization",
        )
                // Preflight request
                if r.Method == http.MethodOptions {
                        w.WriteHeader(http.StatusNoContent)
                        return
                }

                next.ServeHTTP(w, r)
        })
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        if auth == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(auth, "Bearer ")

        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}



func main() {
    initDB()

    mux := http.NewServeMux()

    // public routes
    mux.HandleFunc("/signup", signupHandler)
    mux.HandleFunc("/signin", signinHandler)

    // protected routes
    mux.Handle(
        "/doctorappointment",
        authMiddleware(http.HandlerFunc(createDoctorAppointment)),
    )

  mux.Handle(
        "/scanpicture",
        authMiddleware(http.HandlerFunc(scanPicture)),
    )


    log.Println("Server running on :80")
    log.Fatal(http.ListenAndServe(":80", corsMiddleware(mux)))
}

