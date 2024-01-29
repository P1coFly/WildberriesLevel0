package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	modelsDB "myproject/internal/db"
	models "myproject/internal/models"

	_ "github.com/lib/pq"
	stan "github.com/nats-io/stan.go"
	"github.com/patrickmn/go-cache"
)

func main() {

	log.Println("Start program")

	// Подключение к бд
	connStr := "user=wb_service password=12345678 dbname=WB sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	orderCache := cache.New(-1, -1)

	initializingCache(orderCache, db)

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
		var order models.Order

		if err := json.Unmarshal(msg.Data, &order); err != nil {
			log.Fatal(err)
		}
		log.Printf("Received message: %+v\n", order)
		modelsDB.AddOrder(db, order)
		orderCache.Set(order.OrderUID, order, cache.NoExpiration)

		ord, found := orderCache.Get(order.OrderUID)
		if found {
			log.Print("FROM CACHE:")
			log.Println(ord)
		}

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

func initializingCache(orderCache *cache.Cache, db *sql.DB) {
	uids, err := modelsDB.GetUids(db)
	if err != nil {
		log.Println(err)
		return
	}
	for _, uid := range uids {
		order, err := modelsDB.GetOrderByUID(db, uid)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println(order)

		orderCache.Set(order.OrderUID, order, cache.NoExpiration)

		ord, found := orderCache.Get(order.OrderUID)
		if found {
			log.Print("FROM CACHE:")
			log.Println(ord)
		}
	}

	log.Printf("Восстановление данных из БД в кэш ... записей найдено (%v)\n", len(orderCache.Items()))

}
