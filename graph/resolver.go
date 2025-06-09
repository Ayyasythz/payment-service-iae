package graph

import "payment-service-iae/midtrans"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
type Resolver struct {
	midtransClient *midtrans.Client
}

func NewResolver(midtransClient *midtrans.Client) *Resolver {
	return &Resolver{midtransClient: midtransClient}
}
