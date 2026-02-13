package handlers

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"

    "patienttracker/db"
    "patienttracker/helpers"
)

type SigninRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
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

    // fetch roles assigned to user
    rows, err := db.DB.Query(`
        SELECT r.name
        FROM user_roles ur
        JOIN roles r ON r.id = ur.role_id
        WHERE ur.user_id=$1
    `, userID)
    if err != nil {
        fail("role_query", err)
        return
    }
    defer rows.Close()

    var roles []string
    for rows.Next() {
        var role string
        rows.Scan(&role)
        roles = append(roles, role)
    }

    // enforce exactly ONE role
    if len(roles) != 1 {
        fail("invalid_role_count", nil)
        return
    }

    role := roles[0]

token, err := helpers.GenerateJWT(userID, req.Email, role)
    if err != nil {
        fail("jwt_generation", err)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "email": req.Email,
        "role":  role,
        "token": token,
    })
}
