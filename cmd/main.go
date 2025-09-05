package main

import (
	"context"
	"l0/internal"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.Println("l0 service start")
	err := godotenv.Load()
	if err != nil {
		log.Printf("ошибка загрузки секретов из .env: %v", err)
	}

	connstring := os.Getenv("PG_CONNSTRING") + "?sslmode=" + os.Getenv("PG_SSLMODE")
	db, err := internal.NewDB(connstring)
	if err != nil {
		log.Println(connstring)
		log.Fatal("ошибка подключения к бд: %v. ", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{os.Getenv("KAFKA_CONN")},
		Topic:          os.Getenv("KAFKA_TOPICNAME"),
		GroupID:        os.Getenv("KAFKA_GROUUPID"),
		MinBytes:       10,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		CommitInterval: 0,
	})
	defer reader.Close()

	ctx := context.Background()

	messages := make(chan kafka.Message, 50)
	defer close(messages)
	cache := make(map[string]internal.Order)

	err = internal.FillCache(ctx, db, cache)
	if err != nil {
		log.Println("ошибка заполнения кеша при старте: %w", err)
	}

	go internal.SubscribeOnTopic(ctx, reader, messages)
	go internal.Worker(ctx, messages, db, cache, reader)

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})
	router.GET("/order/:ouid", func(c *gin.Context) {
		orderUID := c.Param("ouid")

		order, err := internal.GetOrderByID(db, orderUID, cache)
		if err != nil {
			log.Printf("заказ не найден: %v. ", err)
			c.JSON(404, gin.H{
				"error": "order not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"order": order,
		})
	})

	router.Run(":" + os.Getenv("HTTP_PORT"))
}
