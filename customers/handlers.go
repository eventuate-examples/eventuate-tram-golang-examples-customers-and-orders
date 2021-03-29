package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"gorm.io/gorm"
)

func handleCreateCustomerRequest(dataBase *gorm.DB, w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	var createCustomerRequest CreateCustomerRequest
	json.Unmarshal(body, &createCustomerRequest)

	customer := Customer{Name: createCustomerRequest.Name, Money: createCustomerRequest.Money}
	dataBase.Create(&customer)

	publishEvent(
		dataBase,
		&CustomerCreatedEvent{customer.Name, customer.Money},
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
		 customer.Id,
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreatedEvent")
}

func handleOrderCreatedEvent(dataBase *gorm.DB, w http.ResponseWriter, r *http.Request) {
	aggregateId := mux.Vars(r)["aggregateId"]

	orderId, err := strconv.ParseInt(aggregateId, 10, 64)

	if (err != nil) {
		panic(err)
	}

	body, _ := ioutil.ReadAll(r.Body)

	var orderCreatedEvent OrderCreatedEvent

	json.Unmarshal(body, &orderCreatedEvent)

	reserveCredit(dataBase, orderCreatedEvent.OrderDetails.CustomerId, orderId, orderCreatedEvent.OrderDetails.OrderTotal)
}

func handleOrderCanceledEvent(dataBase *gorm.DB, w http.ResponseWriter, r *http.Request) {
	aggregateId := mux.Vars(r)["aggregateId"]

	body, _ := ioutil.ReadAll(r.Body)

	var orderCanceledEvent OrderCanceledEvent

	json.Unmarshal(body, &orderCanceledEvent)

	dataBase.Where("customerId = ? and orderId = ?", orderCanceledEvent.OrderDetails.CustomerId, aggregateId).Delete(CreditReservation{})
}

func reserveCredit(dataBase *gorm.DB, customerId int64, orderId int64, orderTotal int64) {
	customer := Customer{}

	err := dataBase.First(&customer, customerId).Error

	if (err != nil) {
		if (errors.Is(err, gorm.ErrRecordNotFound)) {
			publishEvent(
				dataBase,
				&CustomerValidationFailedEvent{OrderEvent{orderId}},
				"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
				customerId,
				"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerValidationFailedEvent")

			return
		} else {
			panic(err)
		}
	}

	var sum int64 = 0

	for _, reservation := range customer.CreditReservations {
		sum = sum + reservation.Money
	}

	if (customer.Money - sum < orderTotal) {
		publishEvent(
			dataBase,
			&CustomerCreditReservationFailedEvent{OrderEvent{orderId}},
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
			customerId,
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservationFailedEvent")

		return
	}

	customer.CreditReservations = append(customer.CreditReservations, CreditReservation{orderId, customerId, orderTotal})

	dataBase.Save(&customer)

	publishEvent(
		dataBase,
		&CustomerCreditReservedEvent{OrderEvent{orderId}},
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
		customerId,
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservedEvent")
}