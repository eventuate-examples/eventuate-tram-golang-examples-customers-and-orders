package main

import (
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/gorm"
)

func publishEvent(dataBase *gorm.DB, domainEvent interface{}, aggregateType string, aggregateId int64, eventType string) {
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