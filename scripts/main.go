package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	stan "github.com/nats-io/stan.go"
)

func init() {

	if err := godotenv.Load(); err != nil {
		log.Panic("No .env file found")
	}
}

func main() {
	// Параметры подключения к NATS Streaming
	clusterID := os.Getenv("CLUSTER_ID")
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
	jsonFiles, err := getJSONFiles("scripts/")
	if err != nil {
		log.Fatal(err)
	}

	for _, jsonFile := range jsonFiles {
		// Чтение JSON файла
		jsonData, err := readJSONFile(jsonFile)
		if err != nil {
			log.Fatal(err)
		}

		// Отправка JSON данных в канал
		err = sc.Publish(channelName, jsonData)
		if err != nil {
			log.Fatal(err)
		}

		parts := strings.Split(jsonFile, "/")
		log.Printf("JSON data: %s - sent to channel: %s\n", parts[len(parts)-1], channelName)
	}
}

// readJSONFile считывает содержимое JSON файла по указанному пути.
// Возвращает считанные данные в виде среза байтов.
func readJSONFile(jsonFile string) ([]byte, error) {
	file, err := os.Open(jsonFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// getJSONFiles возвращает список путей ко всем файлам с расширением .json в указанной директории.
// Использует ioutil.ReadDir для чтения только файлов в указанной директории.
// Возвращаемый срез содержит абсолютные пути ко всем найденным JSON файлам.
func getJSONFiles(dirPath string) ([]string, error) {
	var jsonFiles []string

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, filepath.Join(dirPath, file.Name()))
		}
	}

	return jsonFiles, nil
}
