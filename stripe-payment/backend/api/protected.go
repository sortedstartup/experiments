package api

import (
	"net/http"
	"stripe-payment/db"

	"github.com/gin-gonic/gin"
)

func CheckUserAccess(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	user, err := db.GetUserByEmail(email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"access": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access": true})
}
