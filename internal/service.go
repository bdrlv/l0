package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func GetOrderByID(db *sql.DB, orderUID string, cache map[string]Order) (Order, error) {
	order, ok := cache[orderUID]
	if ok {
		return order, nil
	} else {
		log.Printf("Заказ с orderUID == %v в кеше не найден. ", orderUID)
	}

	order, err := getOrderByIdFromDB(context.Background(), db, orderUID)
	if err != nil {
		return order, fmt.Errorf("ошибка получения заказа из бд: %w. ", err)
	}
	cache[orderUID] = order
	log.Printf("Заказ с orderUID == %v добавлен в кеш. ", orderUID)

	return order, nil
}

func ProcessMessage(ctx context.Context, msg kafka.Message, db *sql.DB, cache map[string]Order) error {
	var order Order
	err := json.Unmarshal(msg.Value, &order)
	if err != nil {
		return fmt.Errorf("ошибка десеарилизации сообщения: %w. ", err)
	}

	log.Printf("Процессинг сообщения заказа с id == %v. ", order.OrderUID)

	ok, err := order.ValidateMessageData()
	if !ok {
		if err != nil {
			return fmt.Errorf("ошибка валидации заказа %v, %w", order.OrderUID, err)
		}
	}

	cache[order.OrderUID] = order

	err = saveOrder(ctx, db, order)
	if err != nil {
		return fmt.Errorf("ошибка сохранения заказа с id == %v в бд: %v. ", order.OrderUID, err)
	}

	return nil
}

func FillCache(ctx context.Context, db *sql.DB, cache map[string]Order) error {
	orders, err := getAlllOrders(ctx, db)
	if err != nil {
		return fmt.Errorf("ошибка получения всех заказов: %v. ", err)
	}

	for _, order := range orders {
		cache[order.OrderUID] = order
		fmt.Printf("Занесено в кеш из БД: %v", order)
	}

	return nil
}
