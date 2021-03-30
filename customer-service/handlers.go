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

	var customerId int64

	err := dataBase.Transaction(func(tx *gorm.DB) error {
		customer := Customer{Name: createCustomerRequest.Name, CreditLimit: int64(createCustomerRequest.CreditLimit.Amount * 100)}
		err := tx.Create(&customer).Error

		if err != nil {
		  return err
		}

		customerId = customer.Id
	  
		err = publishEvent(
			tx,
			&CustomerCreatedEvent{customer.Name, createCustomerRequest.CreditLimit},
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
			 customer.Id,
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreatedEvent")
	  
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	response := CreateCustomerResponse{customerId}	

	payload, err := json.Marshal(response)

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
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

	err = dataBase.Transaction(func(tx *gorm.DB) error {
		return reserveCredit(tx, orderCreatedEvent.OrderDetails.CustomerId, orderId, int64(orderCreatedEvent.OrderDetails.OrderTotal.Amount * 100))
	})

	if (err != nil) {
		panic(err)
	}
}

func handleOrderCanceledEvent(dataBase *gorm.DB, w http.ResponseWriter, r *http.Request) {
	aggregateId := mux.Vars(r)["aggregateId"]

	body, _ := ioutil.ReadAll(r.Body)

	var orderCanceledEvent OrderCanceledEvent

	json.Unmarshal(body, &orderCanceledEvent)

	dataBase.Where("customerId = ? and orderId = ?", orderCanceledEvent.OrderDetails.CustomerId, aggregateId).Delete(CreditReservation{})
}

func reserveCredit(dataBase *gorm.DB, customerId int64, orderId int64, orderTotal int64) error {
	customer := Customer{}

	err := dataBase.First(&customer, customerId).Error

	if (err != nil) {
		if (errors.Is(err, gorm.ErrRecordNotFound)) {
			err = publishEvent(
				dataBase,
				&CustomerValidationFailedEvent{OrderEvent{orderId}},
				"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
				customerId,
				"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerValidationFailedEvent")

			if err != nil {
				return err
			}

			return nil
		} else {
			return err
		}
	}

	var sum int64 = 0

	for _, reservation := range customer.CreditReservations {
		sum = sum + reservation.Money
	}

	if (customer.CreditLimit - sum < orderTotal) {
		err := publishEvent(
			dataBase,
			&CustomerCreditReservationFailedEvent{OrderEvent{orderId}},
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
			customerId,
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservationFailedEvent")

		if err != nil {
			return err
		}

		return nil
	}

	customer.CreditReservations = append(customer.CreditReservations, CreditReservation{orderId, customerId, orderTotal})

	err = dataBase.Save(&customer).Error

	if err != nil {
		return err
	}

	err = publishEvent(
		dataBase,
		&CustomerCreditReservedEvent{OrderEvent{orderId}},
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
		customerId,
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservedEvent")

	if err != nil {
		return err
	}

	return nil
}