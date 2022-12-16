package handlers

import (
	"net/http"

	"github.com/Edbeer/paymentapi/pkg/utils"
	"github.com/golang-jwt/jwt/v4"
)

// auth middleware
func AuthJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("x-jwt-token")
		token, err := utils.ValidateJWT(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}
		uid, err := GetUUID(r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}

		if claims["id"] != uid.String() {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}

		next(w, r)
	}
}