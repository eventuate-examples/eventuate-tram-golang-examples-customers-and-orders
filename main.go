package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/gorilla/mux"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Money int64 `json:"money"`
}

type OrderDetails struct {
	CustomerId int64 `json:"customerId"`
	OrderTotal int64 `json:"orderTotal"`
}

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

type CreditReservation struct {
	OrderId int64 `gorm:"primaryKey;autoIncrement:false"`
	CustomerId int64 `gorm:"primaryKey;autoIncrement:false"`
	Money int64
}

type Message struct {
	Id string `gorm:"primaryKey"` 
	Destination string
	Headers string
	Payload string
	Published int8
	CreationTime int64
}

func (Message) TableName() string {
	return "message"
}

type Customer struct {
	Id int64 `gorm:"primaryKey;autoIncrement:true"`
	Name string
	Money int64
	CreditReservations []CreditReservation
}

var dataBase *gorm.DB

func generateId() (string) {
	return uuid.New().String()
}

func publishEvent(domainEvent interface{}, aggregateType string, aggregateId int64, eventType string) {
	id := generateId()	

    payload, err := json.Marshal(domainEvent)
    
	if err != nil {
        log.Fatal(err)
    }

	dataBase.Exec("insert into eventuate.message (id, destination, headers, payload, published, creation_time) values (?, ?, ?, ?, ?, ROUND(UNIX_TIMESTAMP(CURTIME(4)) * 1000))",
		id,
		aggregateType,
		fmt.Sprintf("{\"ID\" : \"%v\", \"PARTITION_ID\" : %v, \"DESTINATION\" : %v, \"event-type\" : %v, \"event-aggregate-type\" : %v, \"event-aggregate-id\" : %v}",
			id, aggregateId, aggregateType, eventType, aggregateType, aggregateId),
		payload,
		0)	
}

func reserveCredit(customerId int64, orderId int64, orderTotal int64) {
	customer := Customer{}

	err := dataBase.First(&customer, customerId).Error

	if (err != nil) {
		if (errors.Is(err, gorm.ErrRecordNotFound)) {
			publishEvent(
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
			&CustomerCreditReservationFailedEvent{OrderEvent{orderId}},
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
			customerId,
			"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservationFailedEvent")

		return	
	}

	customer.CreditReservations = append(customer.CreditReservations, CreditReservation{orderId, customerId, orderTotal})

	dataBase.Save(&customer)

	publishEvent(
		&CustomerCreditReservedEvent{OrderEvent{orderId}},
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
		customerId,
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreditReservedEvent")
}

func handleCreateCustomerRequest(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	var createCustomerRequest CreateCustomerRequest
	json.Unmarshal(body, &createCustomerRequest)

	customer := Customer{Name: createCustomerRequest.Name, Money: createCustomerRequest.Money}
	dataBase.Create(&customer)

	publishEvent(
		&CustomerCreatedEvent{customer.Name, customer.Money},
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.Customer",
		 customer.Id,
		"io.eventuate.examples.tram.ordersandcustomers.customers.domain.events.CustomerCreatedEvent")
}

func handleOrderCreatedEvent(w http.ResponseWriter, r *http.Request) {
	aggregateId := mux.Vars(r)["aggregateId"]

	orderId, err := strconv.ParseInt(aggregateId, 10, 64)

	if (err != nil) {
		panic(err)
	}

	body, _ := ioutil.ReadAll(r.Body)

	var orderCreatedEvent OrderCreatedEvent

	json.Unmarshal(body, &orderCreatedEvent)

	reserveCredit(orderCreatedEvent.OrderDetails.CustomerId, orderId, orderCreatedEvent.OrderDetails.OrderTotal)
}

func handleOrderCanceledEvent(w http.ResponseWriter, r *http.Request) {
	aggregateId := mux.Vars(r)["aggregateId"]

	body, _ := ioutil.ReadAll(r.Body)

	var orderCanceledEvent OrderCanceledEvent

	json.Unmarshal(body, &orderCanceledEvent)

	dataBase.Where("customerId = ? and orderId = ?", orderCanceledEvent.OrderDetails.CustomerId, aggregateId).Delete(CreditReservation{})
}

func route() {
	router := mux.NewRouter()
	router.HandleFunc("/customers", handleCreateCustomerRequest).Methods("POST")
	router.HandleFunc("/events/orderServiceEvents/io.eventuate.examples.tram.ordersandcustomers.orders.domain.Order/{aggregateId}/io.eventuate.examples.tram.ordersandcustomers.orders.domain.events.OrderCreatedEvent/{eventId}", handleOrderCreatedEvent).Methods("POST")
	router.HandleFunc("/events/orderServiceEvents/io.eventuate.examples.tram.ordersandcustomers.orders.domain.Order/{aggregateId}/io.eventuate.examples.tram.ordersandcustomers.orders.domain.events.OrderCanceledEvent/{eventId}", handleOrderCanceledEvent).Methods("POST")
	log.Fatal(http.ListenAndServe(":10000", router))
}

func initDatabase() {
	dsn := "mysqluser:mysqlpw@tcp(127.0.0.1:3306)/eventuate?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if (err != nil) {
		log.Fatal(err)
	}

	dataBase = db

	dataBase.AutoMigrate(&Customer{})
	dataBase.AutoMigrate(&CreditReservation{})
}

func main() {
	initDatabase()
	route()
}
