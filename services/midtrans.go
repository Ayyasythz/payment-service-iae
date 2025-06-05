package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"payment-service-iae/config"
	"payment-service-iae/models"
	_ "strings"
	"time"
)

type MidtransService struct {
	config *config.Config
	client *http.Client
}

type SnapRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CreditCard         *CreditCard        `json:"credit_card,omitempty"`
	CustomerDetails    *CustomerDetails   `json:"customer_details,omitempty"`
	ItemDetails        []ItemDetail       `json:"item_details,omitempty"`
	Callbacks          *Callbacks         `json:"callbacks,omitempty"`
}

type TransactionDetails struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type CreditCard struct {
	Secure bool `json:"secure"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type ItemDetail struct {
	ID       string  `json:"id"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Name     string  `json:"name"`
}

type Callbacks struct {
	Finish string `json:"finish"`
}

type SnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type TransactionStatusResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	GrossAmount       string `json:"gross_amount"`
}

func NewMidtransService(cfg *config.Config) *MidtransService {
	return &MidtransService{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (m *MidtransService) CreateSnapToken(payment *models.Payment, customerDetails *CustomerDetails) (*SnapResponse, error) {
	snapRequest := SnapRequest{
		TransactionDetails: TransactionDetails{
			OrderID:     payment.OrderID,
			GrossAmount: payment.Amount,
		},
		CreditCard: &CreditCard{
			Secure: true,
		},
		CustomerDetails: customerDetails,
		ItemDetails: []ItemDetail{
			{
				ID:       "item-1",
				Price:    payment.Amount,
				Quantity: 1,
				Name:     payment.Description,
			},
		},
		Callbacks: &Callbacks{
			Finish: "http://localhost:3000/payment/finish",
		},
	}

	jsonData, err := json.Marshal(snapRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", m.config.MidtransBaseURL+"/snap/transactions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+m.encodeServerKey())

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var snapResponse SnapResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapResponse); err != nil {
		return nil, err
	}

	return &snapResponse, nil
}

func (m *MidtransService) GetTransactionStatus(orderID string) (*TransactionStatusResponse, error) {
	req, err := http.NewRequest("GET", m.config.MidtransBaseURL+"/"+orderID+"/status", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+m.encodeServerKey())

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var statusResponse TransactionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
		return nil, err
	}

	return &statusResponse, nil
}

func (m *MidtransService) VerifySignature(orderID, statusCode, grossAmount, serverKey string) string {
	input := orderID + statusCode + grossAmount + serverKey
	hash := sha512.Sum512([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func (m *MidtransService) encodeServerKey() string {
	auth := m.config.MidtransServerKey + ":"
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
