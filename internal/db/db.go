package db_module

import (
	"database/sql"
	"myproject/internal/models"
)

// AddOrder добавляет в бд сущность заказа
// Если происходит ошибка при добавлении данных в таблицы Payment, Items, Delivery, или Orders,
// функция возвращает эту ошибку. В противном случае возвращается nil.
func AddOrder(db *sql.DB, order models.Order) error {

	//Добавляем  данные в таблицу Payment
	err := addPayment(db, order.Payment)
	if err != nil {
		return err
	}

	//Добавляем  данные в таблицу Delivery
	var idxDelivery int
	idxDelivery, err = addDelivery(db, order.Delivery)
	if err != nil {
		return err
	}
	//Добавляем  данные в таблицу Orders
	_, err = db.Exec(`INSERT INTO public."ORDERS" (order_uid,track_number,entry,locale,internal_signature,
	customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard,delivery_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard, idxDelivery)
	if err != nil {
		return err
	}

	//Добавляем  данные в таблицу Items
	err = addItems(db, order.Items)
	return err

}

// addDelivary добавляет данные в таблицу Delivery
// Возвращает идентификатор добавленной записи и ошибку.
func addDelivery(db *sql.DB, delivery models.Delivery) (int, error) {
	var id int

	err := db.QueryRow(`INSERT INTO public."DELIVERIES" (name, phone, zip, city, address, region, email) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) returning id`,
		delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region,
		delivery.Email).Scan(&id)

	return id, err
}

// addPayment добавляет данные в таблицу Payment
// Если происходит ошибка, то возвращает эту ошибку
func addPayment(db *sql.DB, payment models.Payment) error {
	_, err := db.Exec(`INSERT INTO public."PAYMENTS" (transaction,request_id,currency,provider,amount,payment_dt,bank,delivery_cost,goods_total,custom_fee)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDT,
		payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee)

	return err
}

// addItems добавляет данные в таблицу Items
// Если происходит ошибка, то возвращает эту ошибку
func addItems(db *sql.DB, items []models.Item) error {
	var err error
	for _, item := range items {

		_, err = db.Exec(`INSERT INTO public."ITEMS" (chrt_id,track_number,price,rid,name,sale,size,total_price,nm_id,brand,status)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale,
			item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}

	}
	return err
}

// GetUids возвращает срез Uids из бд
// Если происходит ошибка, то возвращает эту ошибку
func GetUids(db *sql.DB) ([]string, error) {
	var uids []string

	rows, err := db.Query(`SELECT order_uid	FROM public."ORDERS"`)
	if err != nil {
		return uids, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid string
		err := rows.Scan(&uid)
		if err != nil {
			continue
		}
		uids = append(uids, uid)
	}

	return uids, err

}

// GetOrderNyUID возвращает заказ по Uid
// Если происходит ошибка, то возвращает эту ошибку
func GetOrderByUID(db *sql.DB, uid string) (models.Order, error) {
	var order models.Order

	row := db.QueryRow(`SELECT 
	o.order_uid,
    o.track_number,
    o.entry,
    o.locale,
    o.internal_signature,
    o.customer_id,
    o.delivery_service,
    o.shardkey,
    o.sm_id,
    o.date_created,
    o.oof_shard,
    d.name,
    d.phone,
    d.zip,
    d.city,
    d.address,
    d.region,
    d.email,
    p.transaction,
    p.request_id,
    p.currency,
    p.provider,
    p.amount,
    p.payment_dt,
    p.bank,
    p.delivery_cost,
    p.goods_total,
    p.custom_fee
	FROM public."ORDERS" o
	JOIN public."DELIVERIES" d ON o.delivery_id = d.id
	JOIN public."PAYMENTS" p ON o.order_uid = p.transaction
	WHERE o.order_uid = $1`, uid)
	err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email, &order.Payment.Transaction, &order.Payment.RequestID,
		&order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return order, err
	}

	order.Items, err = getItemsByTrackNumber(db, order.TrackNumber)

	return order, err

}

// getItemsByTrackNumber возвращает срез Item для конкретного закза по TrackNumber
// Если происходит ошибка, то возвращает эту ошибку
func getItemsByTrackNumber(db *sql.DB, trackNumber string) ([]models.Item, error) {
	var items []models.Item

	rows, err := db.Query(`SELECT
	chrt_id,
	track_number,
	price,
	rid,
	name,
	sale,
	size,
	total_price,
	nm_id,
	brand,
	status 
	FROM public."ITEMS"
	WHERE track_number = $1`, trackNumber)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, err
}
