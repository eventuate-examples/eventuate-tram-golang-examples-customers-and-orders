package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

	logf := func(message string) {
		log.Println(fmt.Sprintf("handleOrderCreatedEvent (orderId = %v); %v", aggregateId, message))
	}

	logf("started")

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

	logf("finished")
}

func handleOrderCanceledEvent(dataBase *gorm.DB, w http.ResponseWriter, r *http.Request) {
	aggregateId := mux.Vars(r)["aggregateId"]

	logf := func(message string) {
		log.Println(fmt.Sprintf("handleOrderCanceledEvent (orderId = %v); %v", aggregateId, message))
	}

	logf("started")

	body, _ := ioutil.ReadAll(r.Body)

	var orderCanceledEvent OrderCanceledEvent

	json.Unmarshal(body, &orderCanceledEvent)

	dataBase.Where("customerId = ? and orderId = ?", orderCanceledEvent.OrderDetails.CustomerId, aggregateId).Delete(CreditReservation{})

	logf("finished")
}

func reserveCredit(dataBase *gorm.DB, customerId int64, orderId int64, orderTotal int64) error {
	logf := func(message string) {
		log.Println(fmt.Sprintf("reserveCredit (customerId = %v, orderId = %v, orderTotal = %v); %v",
			customerId, orderId, orderTotal, message))
	}

	customer := Customer{}

	err := dataBase.First(&customer, customerId).Error

	if (err != nil) {
		if (errors.Is(err, gorm.ErrRecordNotFound)) {
			logf("publishing CustomerValidationFailedEvent")

			err = publishEvent(
				dataBase,
				&CustomerValidationFailedEvent{OrderEvent{orderId}},
				"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
				customerId,
				"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerValidationFailedEvent")

			if err != nil {
				return err
			}

			logf("published CustomerValidationFailedEvent")
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
		logf("publishing CustomerCreditReservationFailedEvent")

		err := publishEvent(
			dataBase,
			&CustomerCreditReservationFailedEvent{OrderEvent{orderId}},
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
			customerId,
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservationFailedEvent")

		if err != nil {
			return err
		}

		logf("published CustomerCreditReservationFailedEvent")

		return nil
	}

	customer.CreditReservations = append(customer.CreditReservations, CreditReservation{orderId, customerId, orderTotal})

	err = dataBase.Save(&customer).Error

	if err != nil {
		return err
	}

	logf("publishing CustomerCreditReservedEvent")

	err = publishEvent(
		dataBase,
		&CustomerCreditReservedEvent{OrderEvent{orderId}},
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
		customerId,
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservedEvent")

	if err != nil {
		return err
	}

	logf("published CustomerCreditReservedEvent")

	return nil
}