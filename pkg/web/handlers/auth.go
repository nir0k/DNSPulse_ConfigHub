package handlers

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"DNSPulse_ConfigHub/pkg/logger"
	"DNSPulse_ConfigHub/pkg/tools"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

var jwtKey = []byte("your_secret_key")

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}


func GenerateToken(username string) (string, error) {
    expirationTime := time.Now().Add(time.Duration(datastore.GetConfig().WebServer.SesionTimeout) * time.Minute)
    claims := &Claims{
        Username: username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)

    return tokenString, err
}


func ValidateToken(tokenString string) (*jwt.Token, error) {

    claims := &Claims{}

    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    return token, err
}

func shouldRefreshToken(token *jwt.Token) bool {
    const refreshInterval = 5 * time.Minute
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return time.Until(time.Unix(claims.ExpiresAt, 0)) < refreshInterval
    }
    return false
}


func getUsernameFromToken(token *jwt.Token) string {
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims.Username
    }
    return ""
}

func setSessionCookie(w http.ResponseWriter, token string) {
	SesionTimeout := datastore.GetConfig().WebServer.SesionTimeout
    http.SetCookie(w, &http.Cookie{
        Name:   "session_token",
        Value:  token,
        Path:   "/",
        MaxAge: SesionTimeout,
    })
}

func CheckCredentials(username, password string) bool {
	conf := datastore.GetConfig().WebServer
    return username == conf.Username && password == conf.Password
}

func AuthMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        isAPIEndpoint := strings.HasPrefix(r.URL.Path, "/api/")
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            cookie, err := r.Cookie("session_token")
            if err != nil {
                if isAPIEndpoint {
                    w.WriteHeader(http.StatusUnauthorized)
                    json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized: No valid token provided"})
                } else {
                    http.Redirect(w, r, "/login", http.StatusFound)
                }
                return
            }
            tokenString = cookie.Value
        }

        token, err := ValidateToken(tokenString)
        if err != nil || !token.Valid {
            if isAPIEndpoint {
                w.WriteHeader(http.StatusUnauthorized)
                json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized: No valid token provided"})
            } else {
                http.Redirect(w, r, "/login", http.StatusFound)
            }
            return
        }

		if shouldRefreshToken(token) {

            if shouldRefreshToken(token) {
                newTokenString, err := GenerateToken(getUsernameFromToken(token))
                if err != nil {
                    logger.Audit.WithFields(logrus.Fields{
                        "old_token": tools.TruncateToken(tokenString),
                        "new_token": tools.TruncateToken(newTokenString),
                    }).Info("JWT token refreshed")
                    http.Redirect(w, r, "/login", http.StatusFound)
                    return
                }

                setSessionCookie(w, newTokenString)
            }
        }

        handler(w, r)
    }
}