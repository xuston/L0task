package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func main() {

	kafkaURL := "localhost:9092"
	topic := "orders"
	message := `{
		"order_uid": "b563feb7b2b84b6test",
		"track_number": "WBILMTESTTRACK",
		"entry": "WBIL",
		"delivery": {
			"name": "Test Testov",
			"phone": "+9720000000",
			"zip": "2639809",
			"city": "Kiryat Mozkin",
			"address": "Ploshad Mira 15",
			"region": "Kraiot",
			"email": "test@gmail.com"
		},
		"payment": {
			"transaction": "b563feb7b2b84b6test",
			"request_id": "",
			"currency": "USD",
			"provider": "wbpay",
			"amount": 1817,
			"payment_dt": 1637907727,
			"bank": "alpha",
			"delivery_cost": 1500,
			"goods_total": 317,
			"custom_fee": 0
		},
		"items": [
			{
				"chrt_id": 9934930,
				"track_number": "WBILMTESTTRACK",
				"price": 453,
				"rid": "ab4219087a764ae0btest",
				"name": "Mascaras",
				"sale": 30,
				"total_price": 317,
				"nm_id": 2389212,
				"brand": "Vivienne Sabo",
				"status": 202
			}
		],
		"locale": "en",
		"internal_signature": "",
		"customer_id": "test",
		"delivery_service": "meest",
		"shardkey": "9",
		"sm_id": 99,
		"date_created": "2021-11-26T06:22:19Z",
		"oof_shard": "1"
	}`

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaURL},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("order_key"),
			Value: []byte(message),
		},
	)

	if err != nil {
		log.Fatalf("Ошибка при отправке сообщения в Kafka: %v", err)
	} else {
		fmt.Println("Сообщение успешно отправлено в Kafka")
	}

	err = writer.Close()
	if err != nil {
		log.Fatalf("Ошибка при закрытии продусера Kafka: %v", err)
	}
}
