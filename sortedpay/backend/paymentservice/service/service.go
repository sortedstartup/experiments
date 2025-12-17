package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sortedstartup/sortedpay/paymentservice/dao"

	razorpay "github.com/razorpay/razorpay-go"
)

type PaymentService struct {
	dao            dao.DAO
	razorpayClient *razorpay.Client
}

type PaymentAdminService struct {
	dao            dao.DAO
	razorpayClient *razorpay.Client
}

func NewPaymentService(daoFactory dao.DAOFactory) (*PaymentService, error) {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		slog.Error("paymentservice:service:NewPaymentService", "error", err)
		return nil, err
	}
	return &PaymentService{
		dao:            dao,
		razorpayClient: razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET")),
	}, nil

}

func NewPaymentAdminService(daoFactory dao.DAOFactory) (*PaymentAdminService, error) {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		slog.Error("paymentservice:service:NewPaymentAdminService", "error", err)
		return nil, err
	}
	return &PaymentAdminService{
		dao:            dao,
		razorpayClient: razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET")),
	}, nil
}

func (a *PaymentAdminService) CreateProduct(ctx context.Context, userID string, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, intervalPeriod string) (string, error) {
	slog.Info("paymentservice:service:CreateProduct", "userID", userID, "name", name, "isRecurring", isRecurring, "intervalCount", intervalCount, "intervalPeriod", intervalPeriod)

	if isRecurring {
		if intervalCount <= 0 {
			slog.Error("paymentservice:service:CreateProduct", "error", "invalid intervalCount for recurring product")
			return "", fmt.Errorf("invalid intervalCount for recurring product")
		}
		// Optional: whitelist intervalPeriod values
		switch intervalPeriod {
		case "week", "month", "quarter", "year":
		default:
			slog.Error("paymentservice:service:CreateProduct", "error", "invalid intervalPeriod for recurring product")
			return "", fmt.Errorf("invalid intervalPeriod for recurring product: %v", intervalPeriod)
		}
	}

	// For one-time payments, set interval values to 0 and empty string to store as NULL
	if !isRecurring {
		intervalCount = 0
		intervalPeriod = ""
	}

	// Convert interval period for Stripe (day, week, month, year)
	stripeInterval := a.convertIntervalForStripe(intervalPeriod)
	stripeIntervalCount := intervalCount

	// Handle quarterly for Stripe (convert to monthly with count 3)
	if intervalPeriod == "quarter" {
		stripeIntervalCount = intervalCount * 3
	}

	// Convert interval period for Razorpay (daily, weekly, monthly, quarterly, yearly)
	razorpayInterval := a.convertIntervalForRazorpay(intervalPeriod)

	// Create product on Stripe
	stripeProductID, err := a.CreateProductStripe(ctx, name, description, amountInSmallestUnit, currency, isRecurring, stripeIntervalCount, stripeInterval)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", "failed to create Stripe product", "details", err)
		return "", fmt.Errorf("failed to create Stripe product")
	}
	slog.Info("paymentservice:service:CreateProduct", "stripeProductID", stripeProductID)

	// Create product on Razorpay
	razorpayProductID, err := a.CreateProductRazorpay(ctx, name, description, amountInSmallestUnit, currency, isRecurring, intervalCount, razorpayInterval)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", "failed to create Razorpay product", "details", err)
		return "", fmt.Errorf("failed to create Razorpay product")
	}
	slog.Info("paymentservice:service:CreateProduct", "razorpayProductID", razorpayProductID)

	// Save to database with both provider IDs
	productID, err := a.dao.CreateProduct(stripeProductID, razorpayProductID, userID, name, description, amountInSmallestUnit, currency, isRecurring, intervalCount, intervalPeriod)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", "failed to save product to database", "details", err)
		return "", fmt.Errorf("failed to save product to database")
	}

	slog.Info("paymentservice:service:CreateProduct", "productID", productID, "stripeProductID", stripeProductID, "razorpayProductID", razorpayProductID)
	return productID, nil
}

func (s *PaymentService) ListProducts(ctx context.Context, userID string) ([]*dao.Product, error) {
	slog.Info("paymentservice:service:ListProducts", "userID", userID)
	products, err := s.dao.ListProducts(userID)
	if err != nil {
		slog.Error("paymentservice:service:ListProducts", "error", err)
		return nil, fmt.Errorf("failed to process the request")
	}
	slog.Info("paymentservice:service:ListProducts", "products", products)
	return products, nil
}

// convertIntervalForStripe converts internal interval to Stripe format
func (a *PaymentAdminService) convertIntervalForStripe(intervalPeriod string) string {
	switch intervalPeriod {
	case "week":
		return "week"
	case "month":
		return "month"
	case "quarter":
		return "month" // Stripe doesn't have quarter, use month with count 3
	case "year":
		return "year"
	default:
		return "month" // default to month
	}
}

// convertIntervalForRazorpay converts internal interval to Razorpay format
func (a *PaymentAdminService) convertIntervalForRazorpay(intervalPeriod string) string {
	switch intervalPeriod {
	case "week":
		return "weekly"
	case "month":
		return "monthly"
	case "quarter":
		return "quarterly"
	case "year":
		return "yearly"
	default:
		return "monthly" // default to monthly
	}
}

func (a *PaymentAdminService) CheckUserProductAccess(ctx context.Context, userID, productID string) (bool, error) {
	slog.Info("paymentservice:service:CheckUserProductAccess", "userID", userID, "productID", productID)

	hasAccess, err := a.dao.CheckUserProductAccess(userID, productID)
	if err != nil {
		slog.Error("paymentservice:service:CheckUserProductAccess", "error", err)
		return false, err
	}

	slog.Info("paymentservice:service:CheckUserProductAccess", "result", hasAccess, "userID", userID, "productID", productID)
	return hasAccess, nil
}

func (a *PaymentAdminService) GetTransactions(ctx context.Context, userID string, pageNumber int32, pageSize int32) ([]*dao.Transaction, error) {

	transactions, err := a.dao.GetTransactions(userID, pageNumber, pageSize)
	if err != nil {
		slog.Error("paymentservice:service:GetTransactions", "error", err)
		return nil, fmt.Errorf("failed to get transactions")
	}
	return transactions, nil
}
