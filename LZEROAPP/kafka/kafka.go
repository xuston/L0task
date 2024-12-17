package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"main.go/models"
)

func Consume(ctx context.Context, topic, broker string, handler func(models.Orders)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "order-group",
	})

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Println("Ошибка чтения сообщения:", err)
			continue
		}

		var order models.Orders
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Println("Ошибка упорядочивания сообщений:", err)
			continue
		}

		handler(order)
	}
}
