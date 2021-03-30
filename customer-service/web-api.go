package main

type CreateCustomerRequest struct {
	Name string `json:"name"`
	CreditLimit Money `json:"creditLimit"`
}

type CreateCustomerResponse struct {
	CustomerId int64 `json:"customerId"`
}

type OrderDetails struct {
	CustomerId int64 `json:"customerId"`
	OrderTotal Money `json:"orderTotal"`
}
