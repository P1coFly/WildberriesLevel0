package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	modelsDB "myproject/internal/db"
	models "myproject/internal/models"

	_ "github.com/lib/pq"
	stan "github.com/nats-io/stan.go"
	"github.com/patrickmn/go-cache"
)

var orderCache *cache.Cache

func main() {

	log.Println("Start program")

	// Подключение к бд
	connStr := "user=wb_service password=12345678 dbname=WB sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("Connection to the database - successful")

	//инициализируем экземпляр объекта кеш
	orderCache = cache.New(-1, -1)
	log.Println("----Starting cache initialization from the database----")
	initializingCache(db)
	log.Printf("----%v records have been added to the cache from the database----\n", len(orderCache.Items()))

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

		//Десериализация полученных данных
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Println("Invalid  json data")
			return
		}
		//пытаемся добавить заказ в бд, если получилось, то и в кеш
		log.Printf("Trying to add an order with uid: %v to the db and to the cache\n", order.OrderUID)
		err = modelsDB.AddOrder(db, order)
		if err != nil {
			log.Printf("Failed to add an order to the db: %v\n", err)
			return
		}
		orderCache.Set(order.OrderUID, order, cache.NoExpiration)
		log.Printf("Order with uid: %v was added\n", order.OrderUID)

	})

	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()
	defer sub.Close()

	log.Println("The connection and subscription  to nats-streaming server - successful")
	log.Println("Waiting for a message...")

	//подниммаем http-сервер
	log.Print("Server start listening on localhost:8080")
	fileServer := http.FileServer(http.Dir("../static/"))
	http.Handle("/static/", http.StripPrefix("/static", fileServer))
	http.HandleFunc("/order", orderHandler)

	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}

func initializingCache(db *sql.DB) {
	//получаем срез uid
	uids, err := modelsDB.GetUids(db)
	if err != nil {
		log.Println(err)
		return
	}
	//получаем заказ по uid и добавляем кго в кеш
	for _, uid := range uids {
		order, err := modelsDB.GetOrderByUID(db, uid)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Trying to add an order with uid: %v to the cache\n", order.OrderUID)
		orderCache.Set(order.OrderUID, order, cache.NoExpiration)
		log.Printf("Order with uid: %v was added\n", order.OrderUID)
	}

}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getOrder(w, r)
	case http.MethodPost:
		postOrder(w, r)
	default:
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
	}
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../templates/index.html")
}

func postOrder(w http.ResponseWriter, r *http.Request) {

	var uid string
	err := json.NewDecoder(r.Body).Decode(&uid)
	log.Printf("Request to receive an order with a Uid: %v\n", uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Достаём из кеша заказ по uid
	ord, found := orderCache.Get(uid)
	if !found {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ord)

}
