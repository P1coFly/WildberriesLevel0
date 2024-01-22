package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	stan "github.com/nats-io/stan.go"
)

func main() {

	log.Println("Start program")
	// Параметры подключения к NATS Streaming
	clusterID := "orders-cluster"
	clientID := "json-consumer"
	channelName := "json-channel"
	natsURL := "nats://localhost:4222"

	// Подключение к NATS Streaming
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	// Подписка на канал
	sub, err := sc.Subscribe(channelName, func(msg *stan.Msg) {
		// Обработка полученного JSON
		log.Printf("Received message: %s\n", string(msg.Data))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()
	defer sub.Close()

	log.Println("The connection and subscription were successful")
	log.Println("Waiting for a message...")

	// Ожидание сигналов завершения работы программы
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Exiting...")
}
