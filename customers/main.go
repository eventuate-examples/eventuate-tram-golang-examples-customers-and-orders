package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	route(initDatabase())
}

func route(dataBase *gorm.DB) {
	router := mux.NewRouter()

	router.HandleFunc("/customers",wrapHandler(dataBase, handleCreateCustomerRequest)).Methods("POST")

	router.
		HandleFunc(
			"/events/orderServiceEvents/io.eventuate.examples.tram.ordersandcustomers.orders.domain.Order/{aggregateId}/io.eventuate.examples.tram.ordersandcustomers.orders.domain.events.OrderCreatedEvent/{eventId}",
			wrapHandler(dataBase, handleOrderCreatedEvent)).
		Methods("POST")

	router.
		HandleFunc(
			"/events/orderServiceEvents/io.eventuate.examples.tram.ordersandcustomers.orders.domain.Order/{aggregateId}/io.eventuate.examples.tram.ordersandcustomers.orders.domain.events.OrderCanceledEvent/{eventId}",
			wrapHandler(dataBase, handleOrderCanceledEvent)).
		Methods("POST")

	log.Fatal(http.ListenAndServe(":10000", router))
}


func wrapHandler(database *gorm.DB, handler func(*gorm.DB, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(database, w, r)
	}
}

func initDatabase() *gorm.DB {
	dsn := "mysqluser:mysqlpw@tcp(127.0.0.1:3306)/eventuate?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if (err != nil) {
		log.Fatal(err)
	}

	db.AutoMigrate(&Customer{})
	db.AutoMigrate(&CreditReservation{})

	return db
}