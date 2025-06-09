package midtrans

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"log"
)

type Client struct {
	snapClient snap.Client
}

func NewClient(serverKey string, env midtrans.EnvironmentType) *Client {
	c := new(snap.Client)
	c.New(serverKey, env)
	return &Client{snapClient: *c}
}

func (c *Client) CreateTransaction(orderID string, amount int64, customer *midtrans.CustomerDetails) (*snap.Response, error) {
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},
		CustomerDetail: customer,
	}

	resp, err := c.snapClient.CreateTransaction(req)
	if err != nil {
		log.Printf("Midtrans API error: %v", err)
	} else {
		log.Printf("Midtrans transaction created successfully: %s", resp.Token)
	}

	return resp, err
}
