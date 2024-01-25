package db_module

import (
	"database/sql"
	"myproject/internal/models"
)

func AddOrder(db *sql.DB, order models.Order) {
	var idxDelivary = addDelivary(db, order.Delivery)

	addPayment(db, order.Payment)

	_, err := db.Exec(`INSERT INTO public."ORDERS" (order_uid,track_number,entry,locale,internal_signature,
	customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard,delivery_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard, idxDelivary)
	if err != nil {
		panic(err)
	}

	addItems(db, order.Items)

}

func addDelivary(db *sql.DB, delivery models.Delivery) int {
	var id int

	err := db.QueryRow(`INSERT INTO public."DELIVERIES" (name, phone, zip, city, address, region, email) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) returning id`,
		delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region,
		delivery.Email).Scan(&id)
	if err != nil {
		panic(err)
	}

	return id
}

func addPayment(db *sql.DB, payment models.Payment) {
	_, err := db.Exec(`INSERT INTO public."PAYMENTS" (transaction,request_id,currency,provider,amount,payment_dt,bank,delivery_cost,goods_total,custom_fee)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDT,
		payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee)
	if err != nil {
		panic(err)
	}
}

func addItems(db *sql.DB, items []models.Item) {
	for _, item := range items {

		_, err := db.Exec(`INSERT INTO public."ITEMS" (chrt_id,track_number,price,rid,name,sale,size,total_price,nm_id,brand,status)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale,
			item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			panic(err)
		}

	}
}
