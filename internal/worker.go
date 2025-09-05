package internal

import (
	"context"
	"database/sql"
	"log"

	"github.com/segmentio/kafka-go"
)

func Worker(ctx context.Context, messages <-chan kafka.Message, db *sql.DB, cache map[string]Order, reader *kafka.Reader) {
	for {
		msg, ok := <-messages
		if !ok {
			continue
		}

		err := ProcessMessage(ctx, msg, db, cache)
		if err != nil {
			log.Printf("Ошибка обработки входящего сообщения: %v.\n", err)
			continue
		}

		err = reader.CommitMessages(ctx, msg)
		if err != nil {
			log.Printf("Коммит сообщения c offset=%v не удался с ошибкой %v\n", msg.Offset, err)
		} else {
			log.Printf("")
		}
	}
}
