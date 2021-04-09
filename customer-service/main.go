package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
			"/events/orderserviceevents/io.eventuate.examples.tram.ordersandcustomers.orders.domain.Order/{aggregateId}/io.eventuate.examples.tram.ordersandcustomers.orders.domain.events.OrderCreatedEvent/{eventId}",
			wrapHandler(dataBase, handleOrderCreatedEvent)).
		Methods("POST")

	router.
		HandleFunc(
			"/events/orderserviceevents/io.eventuate.examples.tram.ordersandcustomers.orders.domain.Order/{aggregateId}/io.eventuate.examples.tram.ordersandcustomers.orders.domain.events.OrderCanceledEvent/{eventId}",
			wrapHandler(dataBase, handleOrderCanceledEvent)).
		Methods("POST")

	log.Println("Starting")	
	log.Fatal(http.ListenAndServe(":8080", router))
}


func wrapHandler(database *gorm.DB, handler func(*gorm.DB, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(database, w, r)
	}
}

func initDatabase() *gorm.DB {
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	name := os.Getenv("DATABASE_NAME")

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)

	log.Println("database connection string:")
	log.Println(dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if (err != nil) {
		log.Fatal(err)
	}

	db.AutoMigrate(&Customer{})
	db.AutoMigrate(&CreditReservation{})

	return db
}