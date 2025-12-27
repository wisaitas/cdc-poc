package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

// DebeziumMessage represents the Debezium CDC message format
type DebeziumMessage struct {
	Payload struct {
		Before *MessageData `json:"before"`
		After  *MessageData `json:"after"`
		Op     string       `json:"op"` // c=create, u=update, d=delete, r=read
	} `json:"payload"`
}

type MessageData struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Kafka topic name follows Debezium naming convention:
	// <connector.name>.<schema>.<table>
	topic := "postgres-connector.public.messages"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		Topic:       topic,
		GroupID:     "cdc-consumer-group",
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	log.Printf("Consumer started, listening to topic: %s", topic)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down consumer...")
		cancel()
	}()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Consumer stopped")
				return
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		var debeziumMsg DebeziumMessage
		if err := json.Unmarshal(msg.Value, &debeziumMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Process the message based on operation type
		switch debeziumMsg.Payload.Op {
		case "c": // Create
			if debeziumMsg.Payload.After != nil {
				data := debeziumMsg.Payload.After
				if data.Status == "pending" {
					fmt.Printf("\n========== NEW PENDING MESSAGE ==========\n")
					fmt.Printf("ID:      %s\n", data.ID)
					fmt.Printf("Content: %s\n", data.Content)
					fmt.Printf("Status:  %s\n", data.Status)
					fmt.Printf("==========================================\n\n")
				}
			}
		case "u": // Update
			if debeziumMsg.Payload.After != nil {
				data := debeziumMsg.Payload.After
				fmt.Printf("\n========== MESSAGE UPDATED ==========\n")
				fmt.Printf("ID:      %s\n", data.ID)
				fmt.Printf("Content: %s\n", data.Content)
				fmt.Printf("Status:  %s\n", data.Status)
				fmt.Printf("======================================\n\n")
			}
		case "d": // Delete
			if debeziumMsg.Payload.Before != nil {
				data := debeziumMsg.Payload.Before
				fmt.Printf("\n========== MESSAGE DELETED ==========\n")
				fmt.Printf("ID:      %s\n", data.ID)
				fmt.Printf("======================================\n\n")
			}
		}
	}
}
