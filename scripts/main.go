package main

import (
	"io"
	"log"
	"os"
	"strings"

	stan "github.com/nats-io/stan.go"
)

func main() {
	// Параметры подключения к NATS Streaming
	clusterID := "orders-cluster"
	clientID := "json-producer"
	channelName := "json-channel"
	natsURL := "nats://localhost:4222"

	// Подключение к NATS Streaming
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	// Путь к JSON файлу
	filePath := "scripts/model.json"

	// Чтение JSON файла
	jsonData, err := readJSONFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Отправка JSON данных в канал
	err = sc.Publish(channelName, jsonData)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(filePath, "/")
	log.Printf("JSON data: %s - sent to channel: %s\n", parts[len(parts)-1], channelName)
}

func readJSONFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Используем ioutil.ReadAll для чтения файла
	jsonData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
