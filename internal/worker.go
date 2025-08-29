package internal

import (
	"context"
	"database/sql"
	"log"

	"github.com/segmentio/kafka-go"
)

func Worker(ctx context.Context, messages <-chan kafka.Message, db *sql.DB, cache map[string]Order) {
	for {
		msg, ok := <-messages
		if !ok {
			return
		}
		err := ProcessMessage(ctx, msg, db, cache)
		if err != nil {
			log.Println("ошибка обработки входящего сообщения: %w. ")
		}
	}
}
