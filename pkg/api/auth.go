package api

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func generateToken(username string) (string, error) {
    jwtKey := datastore.GetJWTKey() // Use the dynamic JWT key from your configuration
    expirationTime := time.Now().Add(60 * time.Minute) // Token expiration
    claims := &jwt.StandardClaims{
        Subject:   username,
        ExpiresAt: expirationTime.Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)

    return tokenString, err
}

func tokenAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization token required"})
            c.Abort()
            return
        }

        tokenString = strings.TrimPrefix(tokenString, "Bearer ")

        token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
            return datastore.GetJWTKey(), nil
        })

        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
            c.Abort()
            return
        }

        if !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
            c.Abort()
            return
        }

        c.Next()
    }
}


func checkCredentials(providedUsername, providedPassword string) bool {
    config := datastore.GetConfig() 
    return providedUsername == config.Api.Username && providedPassword == config.Api.Password
}

func loginHandler(c *gin.Context) {
    username, password, hasAuth := c.Request.BasicAuth()

    if !hasAuth || !checkCredentials(username, password) {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "authentication failed"})
        return
    }

    tokenString, err := generateToken(username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "could not generate token"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
