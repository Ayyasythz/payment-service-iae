package graphql

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"payment-service-iae/middleware"
	"payment-service-iae/models"
	"payment-service-iae/services"
	_ "strconv"
	"time"
)

type Resolver struct {
	paymentService *services.PaymentService
}

func NewResolver(paymentService *services.PaymentService) *Resolver {
	return &Resolver{
		paymentService: paymentService,
	}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Payment(ctx context.Context, id string) (*models.Payment, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	paymentID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid payment ID")
	}

	payment, err := r.paymentService.GetPayment(paymentID)
	if err != nil {
		return nil, err
	}

	// Check if user owns this payment or has admin permissions
	if payment.UserID != user.UserID && !hasPermission(user.Permissions, "read_all_payments") {
		return nil, errors.New("access denied")
	}

	return payment, nil
}

func (r *queryResolver) Payments(ctx context.Context, limit *int, offset *int, filters *PaymentFilters) (*PaymentConnection, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	limitVal := 20
	offsetVal := 0

	if limit != nil {
		limitVal = *limit
		if limitVal > 100 {
			limitVal = 100 // Maximum limit
		}
	}
	if offset != nil {
		offsetVal = *offset
	}

	// Convert GraphQL filters to service filters
	serviceFilters := convertFilters(filters)

	var payments []*models.Payment
	var total int64
	var err error

	// Check if user has permission to read all payments (admin)
	if hasPermission(user.Permissions, "read_all_payments") {
		payments, total, err = r.paymentService.GetAllPayments(limitVal, offsetVal, serviceFilters)
	} else {
		payments, total, err = r.paymentService.GetPaymentsByUserID(user.UserID, limitVal, offsetVal, serviceFilters)
	}

	if err != nil {
		return nil, err
	}

	hasNextPage := int64(offsetVal+limitVal) < total
	hasPreviousPage := offsetVal > 0

	return &PaymentConnection{
		Edges:           payments,
		TotalCount:      int(total),
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPreviousPage,
	}, nil
}

func (r *queryResolver) PaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	payment, err := r.paymentService.GetPaymentByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	// Check if user owns this payment or has admin permissions
	if payment.UserID != user.UserID && !hasPermission(user.Permissions, "read_all_payments") {
		return nil, errors.New("access denied")
	}

	return payment, nil
}

func (r *queryResolver) PaymentStats(ctx context.Context) (*PaymentStats, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	// This is a simplified implementation
	// In a real system, you'd have dedicated stats queries
	var payments []*models.Payment
	var err error

	if hasPermission(user.Permissions, "read_all_payments") {
		payments, _, err = r.paymentService.GetAllPayments(0, 0, nil)
	} else {
		payments, _, err = r.paymentService.GetPaymentsByUserID(user.UserID, 0, 0, nil)
	}

	if err != nil {
		return nil, err
	}

	stats := &PaymentStats{}
	for _, payment := range payments {
		stats.TotalPayments++
		stats.TotalAmount += payment.Amount

		switch payment.Status {
		case models.PaymentStatusPending:
			stats.PendingCount++
		case models.PaymentStatusPaid:
			stats.PaidCount++
		case models.PaymentStatusCanceled:
			stats.CanceledCount++
		case models.PaymentStatusFailed, models.PaymentStatusExpired:
			stats.FailedCount++
		}
	}

	return stats, nil
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePayment(ctx context.Context, input CreatePaymentInput) (*models.Payment, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	// Check create permission
	if !hasPermission(user.Permissions, "create_payment") {
		return nil, errors.New("insufficient permissions")
	}

	serviceInput := services.CreatePaymentInput{
		UserID:        user.UserID,
		Amount:        input.Amount,
		Currency:      input.Currency,
		PaymentMethod: models.PaymentMethod(input.PaymentMethod),
		Description:   input.Description,
	}

	return r.paymentService.CreatePayment(serviceInput)
}

func (r *mutationResolver) CancelPayment(ctx context.Context, id string) (*models.Payment, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	paymentID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid payment ID")
	}

	// Check update permission
	if !hasPermission(user.Permissions, "update_payment") {
		return nil, errors.New("insufficient permissions")
	}

	if err := r.paymentService.CancelPayment(paymentID, user.UserID); err != nil {
		return nil, err
	}

	return r.paymentService.GetPayment(paymentID)
}

// Helper functions
func convertFilters(filters *PaymentFilters) *services.PaymentSearchFilters {
	if filters == nil {
		return nil
	}

	serviceFilters := &services.PaymentSearchFilters{}

	if filters.Status != nil {
		status := models.PaymentStatus(*filters.Status)
		serviceFilters.Status = &status
	}

	if filters.PaymentMethod != nil {
		method := models.PaymentMethod(*filters.PaymentMethod)
		serviceFilters.PaymentMethod = &method
	}

	if filters.StartDate != nil {
		if date, err := time.Parse(time.RFC3339, *filters.StartDate); err == nil {
			serviceFilters.StartDate = &date
		}
	}

	if filters.EndDate != nil {
		if date, err := time.Parse(time.RFC3339, *filters.EndDate); err == nil {
			serviceFilters.EndDate = &date
		}
	}

	if filters.MinAmount != nil {
		serviceFilters.MinAmount = filters.MinAmount
	}

	if filters.MaxAmount != nil {
		serviceFilters.MaxAmount = filters.MaxAmount
	}

	return serviceFilters
}

func hasPermission(permissions []string, required string) bool {
	for _, perm := range permissions {
		if perm == required {
			return true
		}
	}
	return false
}

// GraphQL input types
type CreatePaymentInput struct {
	Amount        float64              `json:"amount"`
	Currency      string               `json:"currency"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod"`
	Description   string               `json:"description"`
}

type PaymentFilters struct {
	Status        *models.PaymentStatus `json:"status"`
	PaymentMethod *models.PaymentMethod `json:"paymentMethod"`
	StartDate     *string               `json:"startDate"`
	EndDate       *string               `json:"endDate"`
	MinAmount     *float64              `json:"minAmount"`
	MaxAmount     *float64              `json:"maxAmount"`
}

type PaymentConnection struct {
	Edges           []*models.Payment `json:"edges"`
	TotalCount      int               `json:"totalCount"`
	HasNextPage     bool              `json:"hasNextPage"`
	HasPreviousPage bool              `json:"hasPreviousPage"`
}

type PaymentStats struct {
	TotalPayments int     `json:"totalPayments"`
	TotalAmount   float64 `json:"totalAmount"`
	PendingCount  int     `json:"pendingCount"`
	PaidCount     int     `json:"paidCount"`
	CanceledCount int     `json:"canceledCount"`
	FailedCount   int     `json:"failedCount"`
}

// Interface implementations
type QueryResolver interface {
	Payment(ctx context.Context, id string) (*models.Payment, error)
	Payments(ctx context.Context, limit *int, offset *int, filters *PaymentFilters) (*PaymentConnection, error)
	PaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	PaymentStats(ctx context.Context) (*PaymentStats, error)
}

type MutationResolver interface {
	CreatePayment(ctx context.Context, input CreatePaymentInput) (*models.Payment, error)
	CancelPayment(ctx context.Context, id string) (*models.Payment, error)
}
