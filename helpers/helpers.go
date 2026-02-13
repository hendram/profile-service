package helpers

import (
    "crypto/tls"
    "crypto/rand"
    "math/big"
    "fmt"
    "net/smtp"
    "net"
    "os"
    "time"
    "log"
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
    err := bcrypt.CompareHashAndPassword(
        []byte(hash),
        []byte(password),
    )

    ok := err == nil

    log.Printf("CheckPasswordHash result=%v err=%v", ok, err)

    return ok
}

func GenerateJWT(userID int, email string, role string) (string, error) {
    secret := []byte(os.Getenv("JWT_SECRET"))

    if len(secret) == 0 {
        return "", fmt.Errorf("missing JWT_SECRET env")
    }

    if role == "" {
        return "", fmt.Errorf("empty role not allowed")
    }

    claims := jwt.MapClaims{
        "sub":   userID,
        "email": email,
        "role":  role,
        "iss":   "onlineshop-api",
        "aud":   "onlineshop-client",
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
        "iat":   time.Now().Unix(),
        "nbf":   time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}


func GenerateVerificationCode(_ string) string {
    n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
    return fmt.Sprintf("%06d", n.Int64())
}


func SendVerificationEmail(toEmail, code string) error {
    host := "smtp-relay.brevo.com"
    port := "587"
    addr := net.JoinHostPort(host, port)

    auth := smtp.PlainAuth("", os.Getenv("USERNAME"), os.Getenv("BREVO_SMTP_KEY"), host)

    msg := []byte(
        "From: greeny.bignose@gmail.com\r\n" +
            "To: " + toEmail + "\r\n" +
            "Subject: Verify your account\r\n" +
            "MIME-Version: 1.0\r\n" +
            "Content-Type: text/plain; charset=utf-8\r\n\r\n" +
            "Your verification code:\n\n" + code +
            "\n\nThis code expires in 5 minutes.\r\n",
    )

    // Connect to server with timeout
    conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
    if err != nil {
        log.Println("SMTP dial failed:", err)
        return err
    }
    defer conn.Close()

    c, err := smtp.NewClient(conn, host)
    if err != nil {
        log.Println("SMTP client creation failed:", err)
        return err
    }
    defer c.Quit()

    // Upgrade to TLS
    tlsconfig := &tls.Config{
        ServerName: host,
    }
    if err = c.StartTLS(tlsconfig); err != nil {
        log.Println("SMTP STARTTLS failed:", err)
        return err
    }

    // Authenticate
    if err = c.Auth(auth); err != nil {
        log.Println("SMTP auth failed:", err)
        return err
    }

    // Send email
    if err = c.Mail("greeny.bignose@gmail.com"); err != nil {
        log.Println("SMTP MAIL FROM failed:", err)
        return err
    }

    if err = c.Rcpt(toEmail); err != nil {
        log.Println("SMTP RCPT TO failed:", err)
        return err
    }

    w, err := c.Data()
    if err != nil {
        log.Println("SMTP DATA failed:", err)
        return err
    }

    _, err = w.Write(msg)
    if err != nil {
        log.Println("SMTP write failed:", err)
        return err
    }

    err = w.Close()
    if err != nil {
        log.Println("SMTP close failed:", err)
        return err
    }

    log.Println("Email sent successfully to", toEmail)
    return nil
}
