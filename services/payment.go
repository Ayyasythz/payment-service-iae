package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-service-iae/models"
	_ "strings"
	"time"
)

type PaymentService struct {
	db         *gorm.DB
	midtrans   *MidtransService
	userClient *UserClient
}

func NewPaymentService(db *gorm.DB, midtrans *MidtransService, userClient *UserClient) *PaymentService {
	return &PaymentService{
		db:         db,
		midtrans:   midtrans,
		userClient: userClient,
	}
}

type CreatePaymentInput struct {
	UserID        uuid.UUID            `json:"user_id"`
	Amount        float64              `json:"amount"`
	Currency      string               `json:"currency"`
	PaymentMethod models.PaymentMethod `json:"payment_method"`
	Description   string               `json:"description"`
}

type PaymentSearchFilters struct {
	Status        *models.PaymentStatus `json:"status"`
	PaymentMethod *models.PaymentMethod `json:"payment_method"`
	StartDate     *time.Time            `json:"start_date"`
	EndDate       *time.Time            `json:"end_date"`
	MinAmount     *float64              `json:"min_amount"`
	MaxAmount     *float64              `json:"max_amount"`
}

func (s *PaymentService) CreatePayment(input CreatePaymentInput) (*models.Payment, error) {
	if input.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if input.Currency == "" {
		input.Currency = "IDR"
	}

	// Get user info from user service
	userInfo, err := s.userClient.GetUserByID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if !userInfo.IsActive {
		return nil, errors.New("user account is not active")
	}

	payment := &models.Payment{
		ID:            uuid.New(),
		UserID:        input.UserID,
		OrderID:       generateOrderID(),
		Amount:        input.Amount,
		Currency:      input.Currency,
		Status:        models.PaymentStatusPending,
		PaymentMethod: input.PaymentMethod,
		Description:   input.Description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Set expiration time (24 hours from now)
	expiredAt := time.Now().Add(24 * time.Hour)
	payment.ExpiredAt = &expiredAt

	if err := s.db.Create(payment).Error; err != nil {
		return nil, err
	}

	// Create Midtrans snap token
	customerDetails := &CustomerDetails{
		Email:     userInfo.Email,
		Phone:     userInfo.Phone,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
	}

	snapResponse, err := s.midtrans.CreateSnapToken(payment, customerDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create midtrans token: %w", err)
	}

	// Update payment with Midtrans token and URL
	payment.MidtransToken = snapResponse.Token
	payment.MidtransURL = snapResponse.RedirectURL

	if err := s.db.Save(payment).Error; err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentService) GetPayment(id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.First(&payment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

func (s *PaymentService) GetPaymentByOrderID(orderID string) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

func (s *PaymentService) GetPaymentsByUserID(userID uuid.UUID, limit, offset int, filters *PaymentSearchFilters) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	query := s.db.Where("user_id = ?", userID)

	// Apply filters
	if filters != nil {
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
		if filters.PaymentMethod != nil {
			query = query.Where("payment_method = ?", *filters.PaymentMethod)
		}
		if filters.StartDate != nil {
			query = query.Where("created_at >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			query = query.Where("created_at <= ?", *filters.EndDate)
		}
		if filters.MinAmount != nil {
			query = query.Where("amount >= ?", *filters.MinAmount)
		}
		if filters.MaxAmount != nil {
			query = query.Where("amount <= ?", *filters.MaxAmount)
		}
	}

	// Count total records
	if err := query.Model(&models.Payment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func (s *PaymentService) GetAllPayments(limit, offset int, filters *PaymentSearchFilters) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	query := s.db.Model(&models.Payment{})

	// Apply filters
	if filters != nil {
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
		if filters.PaymentMethod != nil {
			query = query.Where("payment_method = ?", *filters.PaymentMethod)
		}
		if filters.StartDate != nil {
			query = query.Where("created_at >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			query = query.Where("created_at <= ?", *filters.EndDate)
		}
		if filters.MinAmount != nil {
			query = query.Where("amount >= ?", *filters.MinAmount)
		}
		if filters.MaxAmount != nil {
			query = query.Where("amount <= ?", *filters.MaxAmount)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func (s *PaymentService) UpdatePaymentStatus(orderID string, status models.PaymentStatus, transactionID string) error {
	return s.db.Model(&models.Payment{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":         status,
			"transaction_id": transactionID,
			"updated_at":     time.Now(),
		}).Error
}

func (s *PaymentService) CancelPayment(paymentID uuid.UUID, userID uuid.UUID) error {
	// Verify payment belongs to user and is cancellable
	var payment models.Payment
	if err := s.db.Where("id = ? AND user_id = ?", paymentID, userID).First(&payment).Error; err != nil {
		return errors.New("payment not found")
	}

	if payment.Status != models.PaymentStatusPending {
		return errors.New("payment cannot be cancelled")
	}

	return s.db.Model(&payment).Updates(map[string]interface{}{
		"status":     models.PaymentStatusCanceled,
		"updated_at": time.Now(),
	}).Error
}

func (s *PaymentService) HandleNotification(notification map[string]interface{}) error {
	orderID, ok := notification["order_id"].(string)
	if !ok {
		return errors.New("invalid order_id in notification")
	}

	payment, err := s.GetPaymentByOrderID(orderID)
	if err != nil {
		return err
	}

	// Convert notification to JSON for storage
	rawNotification, _ := json.Marshal(notification)

	// Save notification
	notificationRecord := &models.PaymentNotification{
		PaymentID:         payment.ID,
		TransactionStatus: getStringValue(notification, "transaction_status"),
		FraudStatus:       getStringValue(notification, "fraud_status"),
		PaymentType:       getStringValue(notification, "payment_type"),
		RawNotification:   string(rawNotification),
		CreatedAt:         time.Now(),
	}

	if err := s.db.Create(notificationRecord).Error; err != nil {
		return err
	}

	// Update payment status based on notification
	transactionStatus := getStringValue(notification, "transaction_status")
	var newStatus models.PaymentStatus

	switch transactionStatus {
	case "capture", "settlement":
		newStatus = models.PaymentStatusPaid
	case "cancel":
		newStatus = models.PaymentStatusCanceled
	case "expire":
		newStatus = models.PaymentStatusExpired
	case "failure":
		newStatus = models.PaymentStatusFailed
	default:
		newStatus = models.PaymentStatusPending
	}

	transactionID := getStringValue(notification, "transaction_id")
	return s.UpdatePaymentStatus(orderID, newStatus, transactionID)
}

func generateOrderID() string {
	return fmt.Sprintf("PAY-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
}

func getStringValue(m map[string]interface{}, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}
