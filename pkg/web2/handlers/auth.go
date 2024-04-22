package handlers2

import (
	"net/http"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to extract the token from the Authorization header first
		tokenString := r.Header.Get("Authorization")
		if tokenString != "" {
			// If present, trim the 'Bearer ' prefix
			const bearerPrefix = "Bearer "
			if strings.HasPrefix(tokenString, bearerPrefix) {
				tokenString = strings.TrimPrefix(tokenString, bearerPrefix)
			} else {
				// If the Authorization header is not in the expected format,
				// treat it as if no token was provided
				tokenString = ""
			}
		}

		// If the token wasn't in the Authorization header, try the cookie
		if tokenString == "" {
			cookie, err := r.Cookie("session_token")
			if err == nil {
				tokenString = cookie.Value
			}
		}

		// Now, tokenString should contain just the token or be empty if no token was found
		if tokenString == "" {
			// No token found, redirect to login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Validate the token with the external service
		isValid, err := validateTokenWithExternalService(tokenString)
		if err != nil || !isValid {
			// Validation failed, redirect to login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Token is valid, proceed with the request
		next(w, r)
	}
}

func validateTokenWithExternalService(tokenString string) (bool, error) {
	// Create a new request to the external validation endpoint
	req, err := http.NewRequest("GET", "http://localhost:8080/token/validate", nil) // Assuming no body is required
	if err != nil {
		return false, err
	}

	// Set the Authorization header with the Bearer token
	req.Header.Add("Authorization", "Bearer "+tokenString)

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check the HTTP status code to determine if the token is valid
	if resp.StatusCode == http.StatusOK {
		// Optionally, you might want to check the response body for more details
		return true, nil
	}

	// If the status code is not 200 OK, assume the token is invalid
	return false, nil
}
