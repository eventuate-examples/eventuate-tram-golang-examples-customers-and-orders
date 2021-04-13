package main

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

func publishEvent(dataBase *gorm.DB, domainEvent interface{}, aggregateType string, aggregateId int64, eventType string) error {
	payload, err := json.Marshal(domainEvent)

	if err != nil {
		return err
	}

	return dataBase.Exec(
		"insert into eventuate.message (id, destination, headers, payload, published, creation_time) values ('', ?, ?, ?, ?, ROUND(UNIX_TIMESTAMP(CURTIME(4)) * 1000))",
		aggregateType,
		fmt.Sprintf(
			"{\"PARTITION_ID\" : \"%v\", \"DESTINATION\" : \"%v\", \"event-type\" : \"%v\", \"event-aggregate-type\" : \"%v\", \"event-aggregate-id\" : \"%v\"}",
			aggregateId, aggregateType, eventType, aggregateType, aggregateId),
		payload,
		0).Error

}