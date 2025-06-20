package api

import (
	"github.com/gin-gonic/gin"
)

func RegisterStripeRoutes(r *gin.Engine) {
	paymentGroup := r.Group("/api/payment")
	{
		paymentGroup.POST("/checkout-session", CreateCheckoutSession)
		paymentGroup.POST("/subscription-session", CreateSubscriptionSession)
		paymentGroup.POST("/webhook", HandleStripeWebhook)
	}
}
