package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"stripe-payment/db"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/webhook"
)

func InitStripe() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// checkout session for one-time payment
func CreateCheckoutSession(c *gin.Context) {
	InitStripe()

	// Step 1: Get the price object
	priceID := os.Getenv("STRIPE_ONE_TIME_PRICE_ID")
	priceObj, err := price.Get(priceID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price"})
		return
	}

	// Step 2: Get product details
	productID := priceObj.Product.ID
	productObj, err := product.Get(productID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	// Step 3: Create the session
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(os.Getenv("FRONTEND_URL") + "/success"),
		CancelURL:  stripe.String(os.Getenv("FRONTEND_URL") + "/cancel"),
	}

	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return session ID and product name
	c.JSON(http.StatusOK, gin.H{
		"sessionId":   s.ID,
		"productName": productObj.Name,
	})
}

// checkout session for subscription
func CreateSubscriptionSession(c *gin.Context) {
	InitStripe()

	productID := os.Getenv("STRIPE_PRODUCT_ID")
	productData, err := product.Get(productID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch product"})
		return
	}

	fmt.Println("Product Name:", productData.Name)
	fmt.Println("Product Description:", productData.Description)

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(os.Getenv("STRIPE_SUBSCRIPTION_PRICE_ID")),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(os.Getenv("FRONTEND_URL") + "/success"),
		CancelURL:  stripe.String(os.Getenv("FRONTEND_URL") + "/cancel"),
	}

	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": s.ID,
		"product": gin.H{
			"name":        productData.Name,
			"description": productData.Description,
		},
	})
}

// Webhook handler to listen for Stripe events
func HandleStripeWebhook(c *gin.Context) {
	InitStripe()

	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusServiceUnavailable, "Error reading request body")
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	sigHeader := c.GetHeader("Stripe-Signature")

	log.Println("Webhook hit")
	log.Println("Stripe Signature:", sigHeader)
	log.Println("Payload body:", string(payload))

	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("Signature verification failed: %v", err)
		c.String(http.StatusBadRequest, fmt.Sprintf("Webhook signature verification failed: %v", err))
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Error parsing webhook JSON: %v", err))
			return
		}

		email := session.CustomerDetails.Email
		if email == "" {
			// Fallback if CustomerDetails is nil
			email = session.CustomerEmail
		}

		isSub := session.Mode == stripe.CheckoutSessionModeSubscription

		fmt.Printf("User %s completed %s. Session ID: %s\n", email, session.Mode, session.ID)

		err = db.SaveUserAfterPayment(email, isSub, session.ID)
		if err != nil {
			fmt.Printf("Error saving user to DB: %v\n", err)
			c.String(http.StatusInternalServerError, "error")
			return
		}

		c.String(http.StatusOK, "User saved")

	case "invoice.payment_failed":
		fmt.Println("Payment failed for a subscription.")

	default:
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	c.String(http.StatusOK, "Webhook received")
}
