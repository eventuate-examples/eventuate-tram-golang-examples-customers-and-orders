package main

type OrderCreatedEvent struct {
	OrderDetails OrderDetails `json:"orderDetails"`
}

type OrderCanceledEvent struct {
	OrderDetails OrderDetails `json:"orderDetails"`
}

type CustomerCreatedEvent struct {
	Name string `json:"name"`
	CreditLimit int64 `json:"creditLimit"`
}

type OrderEvent struct {
	OrderId int64
}

type CustomerValidationFailedEvent struct {
	OrderEvent
}

type CustomerCreditReservedEvent struct {
	OrderEvent
}

type CustomerCreditReservationFailedEvent struct {
	OrderEvent
}