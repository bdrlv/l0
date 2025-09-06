package main

import (
	"context"
	"encoding/json"
	"l0/internal"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	// хардкод
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:29092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}

	orderExample := internal.Order{
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: struct {
			Name    string `json:"name"`
			Phone   string `json:"phone"`
			Zip     string `json:"zip"`
			City    string `json:"city"`
			Address string `json:"address"`
			Region  string `json:"region"`
			Email   string `json:"email"`
		}{
			Name:    "Test Testov",
			Phone:   "+98720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: struct {
			Transaction  string `json:"transaction"`
			RequestID    string `json:"request_id"`
			Currency     string `json:"currency"`
			Provider     string `json:"provider"`
			Amount       int    `json:"amount"`
			PaymentDt    int64  `json:"payment_dt"`
			Bank         string `json:"bank"`
			DeliveryCost int    `json:"delivery_cost"`
			GoodsTotal   int    `json:"goods_total"`
			CustomFee    int    `json:"custom_fee"`
		}{
			Transaction: "b563feb7b2b84b6test",
			RequestID:   "",
			Currency:    "USD",
			Provider:    "wbpay",
			Amount:      1817,
			// PaymentDt:    1637907727,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []struct {
			ChrtID      int    `json:"chrt_id"`
			TrackNumber string `json:"track_number"`
			Price       int    `json:"price"`
			Rid         string `json:"rid"`
			Name        string `json:"name"`
			Sale        int    `json:"sale"`
			Size        string `json:"size"`
			TotalPrice  int    `json:"total_price"`
			NmID        int    `json:"nm_id"`
			Brand       string `json:"brand"`
			Status      int    `json:"status"`
		}{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       "2021-11-26T06:22:19Z",
		OofShard:          "1",
	}

	ctx := context.Background()

	log.Println("эмулятор источника данных запущен...")

	i := 1
	originalLen := len(orderExample.OrderUID)
	newuid := orderExample.OrderUID
	for {
		newuid = newuid[:originalLen] + strconv.Itoa(i)
		orderExample.OrderUID = newuid
		orderExample.Payment.Transaction = newuid
		value, err := json.Marshal(orderExample)
		if err != nil {
			log.Printf("ошибка сериализации: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		i++

		err = writer.WriteMessages(ctx, kafka.Message{
			Value: value,
		})
		if err != nil {
			log.Printf("ошибка отправки в топик: %v", err)
		} else {
			log.Printf("сообщение отправлено: %s", orderExample.OrderUID)
		}

		time.Sleep(1 * time.Second)
	}
}
