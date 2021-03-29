package main

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Money int64 `json:"money"`
}

type OrderDetails struct {
	CustomerId int64 `json:"customerId"`
	OrderTotal int64 `json:"orderTotal"`
}
