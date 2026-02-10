package helpers

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "net/smtp"
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)


func HashPassword(password string) string {
    hash, err := bcrypt.GenerateFromPassword(
        []byte(password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        panic(err)
    }
    return string(hash)
}

func CheckPasswordHash(password, hash string) bool {
    return bcrypt.CompareHashAndPassword(
        []byte(hash),
        []byte(password),
    ) == nil
}

func GenerateJWT(email, usertype string) (string, error) {
    secret := []byte(os.Getenv("JWT_SECRET"))

    claims := jwt.MapClaims{
        "email": email,
        "role":  usertype,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
        "iat":   time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

func GenerateVerificationCode(email string) string {
    secret := []byte(os.Getenv("VERIFY_SECRET"))

    ts := time.Now().Unix()
    payload := fmt.Sprintf("%s|%d", email, ts)

    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(payload))
    sig := fmt.Sprintf("%x", mac.Sum(nil))

    raw := fmt.Sprintf("%s|%d|%s", email, ts, sig)
    return base64.StdEncoding.EncodeToString([]byte(raw))
}

func SendVerificationEmail(toEmail, code string) error {
    host := "smtp-relay.brevo.com"
    port := "587"

    auth := smtp.PlainAuth(
        "",
        "apikey",
        os.Getenv("BREVO_SMTP_KEY"),
        host,
    )

    msg := []byte(
        "From: greeny.bignose@gmail.com\r\n" +
            "To: " + toEmail + "\r\n" +
            "Subject: Verify your account\r\n" +
            "MIME-Version: 1.0\r\n" +
            "Content-Type: text/plain; charset=utf-8\r\n\r\n" +
            "Your verification code:\n\n" + code +
            "\n\nThis code expires in 5 minutes.\r\n",
    )

    return smtp.SendMail(
        host+":"+port,
        auth,
        "greeny.bignose@gmail.com",
        []string{toEmail},
        msg,
    )
}
