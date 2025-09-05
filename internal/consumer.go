package internal

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func SubscribeOnTopic(ctx context.Context, reader *kafka.Reader, messages chan<- kafka.Message) {
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("message reading error: %v", err)
			continue
		}

		messages <- msg
	}
}
