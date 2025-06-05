package models

import (
	"github.com/google/uuid"
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusPaid     PaymentStatus = "PAID"
	PaymentStatusCanceled PaymentStatus = "CANCELED"
	PaymentStatusExpired  PaymentStatus = "EXPIRED"
	PaymentStatusFailed   PaymentStatus = "FAILED"
)

type PaymentMethod string

const (
	PaymentMethodCreditCard     PaymentMethod = "CREDIT_CARD"
	PaymentMethodBankTransfer   PaymentMethod = "BANK_TRANSFER"
	PaymentMethodEWallet        PaymentMethod = "E_WALLET"
	PaymentMethodVirtualAccount PaymentMethod = "VIRTUAL_ACCOUNT"
)

type Payment struct {
	ID            uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID     `json:"user_id" gorm:"type:uuid;not null;index"`
	OrderID       string        `json:"order_id" gorm:"unique;not null"`
	Amount        float64       `json:"amount" gorm:"not null"`
	Currency      string        `json:"currency" gorm:"default:'IDR'"`
	Status        PaymentStatus `json:"status" gorm:"default:'PENDING'"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Description   string        `json:"description"`
	MidtransToken string        `json:"midtrans_token"`
	MidtransURL   string        `json:"midtrans_url"`
	TransactionID string        `json:"transaction_id"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	ExpiredAt     *time.Time    `json:"expired_at"`
}

type PaymentNotification struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PaymentID         uuid.UUID `json:"payment_id" gorm:"type:uuid;not null"`
	TransactionStatus string    `json:"transaction_status"`
	FraudStatus       string    `json:"fraud_status"`
	PaymentType       string    `json:"payment_type"`
	RawNotification   string    `json:"raw_notification" gorm:"type:text"`
	CreatedAt         time.Time `json:"created_at"`
	Payment           Payment   `json:"payment" gorm:"foreignKey:PaymentID"`
}
