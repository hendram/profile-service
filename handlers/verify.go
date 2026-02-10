package handlers

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "patienttracker/db"
)

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Code string `json:"code"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

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

    if time.Now().Unix()-ts > 300 {
        http.Error(w, "Code expired", http.StatusUnauthorized)
        return
    }

    secret := []byte(os.Getenv("VERIFY_SECRET"))
    payload := fmt.Sprintf("%s|%s", email, tsStr)

    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(payload))
    expectedSig := fmt.Sprintf("%x", mac.Sum(nil))

    if !hmac.Equal([]byte(sigHex), []byte(expectedSig)) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // ✅ FIXED: use db.DB
    _, err = db.DB.Exec(`
        UPDATE signup SET is_verified = TRUE WHERE email = $1
    `, email)

    if err != nil {
        http.Error(w, "User not found or server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
