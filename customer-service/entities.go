package main

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
	CreditLimit int64
	CreditReservations []CreditReservation
}