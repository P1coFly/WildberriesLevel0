drop table if exists public."DELIVERIES" CASCADE;

CREATE TABLE public."DELIVERIES"
(
    id bigserial NOT NULL,
	name text,
	phone text,
	zip text,
	city text,
	address text,
	region text,
	email text,
    PRIMARY KEY (id)
);

ALTER TABLE public."DELIVERIES"
    OWNER to postgres;		

---------------------------------------------------------
drop table if exists public."PAYMENTS" CASCADE;

CREATE TABLE public."PAYMENTS"
(
	transaction text NOT NULL,
    request_id text,							
    currency text,							
    provider text,						
    amount real,								
    payment_dt INTEGER,
    bank text,							
    delivery_cost real,					
    goods_total real,							
    custom_fee real,
    PRIMARY KEY (transaction)
);

ALTER TABLE public."PAYMENTS"
    OWNER to postgres;		

---------------------------------------------------------
drop table if exists public."ORDERS" CASCADE;

CREATE TABLE public."ORDERS"
(
	order_uid text NOT NULL,
	track_number text UNIQUE,
	entry text,
	locale text,
	internal_signature text,
	customer_id text,
	delivery_service text,
	shardkey text,
	sm_id integer,
	date_created timestamp,
	oof_shard text,
	delivery_id integer,
	FOREIGN KEY (delivery_id) REFERENCES "DELIVERIES"(id),
	FOREIGN KEY (order_uid) REFERENCES "PAYMENTS"(transaction),
	PRIMARY KEY (order_uid)
);

ALTER TABLE public."ORDERS"
    OWNER to postgres;				

---------------------------------------------------------
drop table if exists public."ITEMS" CASCADE;

CREATE TABLE public."ITEMS"
(
    id bigserial NOT NULL,
	chrt_id integer,
	track_number text,
	price real,
	rid text,
	name text,
	sale integer,
	size text,
	total_price real,
	nm_id integer,
	brand text,
	status integer,
	FOREIGN KEY (track_number) REFERENCES "ORDERS"(track_number),
    PRIMARY KEY (id)
);

ALTER TABLE public."ITEMS"
    OWNER to postgres;

	
-- INSERT INTO public."DELIVERIES" (name, phone, zip, city, address, region, email)
-- VALUES
--     ('Test Testov', '+9720000000','2639809','Kiryat Mozkin', 'Ploshad Mira 15', 'Kraiot','test@gmail.com');

-- INSERT INTO public."PAYMENTS" (transaction,request_id,currency,provider,amount,payment_dt,bank,delivery_cost,goods_total,custom_fee)
-- VALUES
--     ('b563feb7b2b84b6test','','USD','wbpay',1817,1637907727,'alpha',1500,317,0);

-- INSERT INTO public."ORDERS" (order_uid,track_number,entry,locale,internal_signature,customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard,delivery_id)
-- VALUES
--     ('b563feb7b2b84b6test','WBILMTESTTRACK','WBIL','en','','test','meest','9',99,'2021-11-26T06:22:19Z','1',1);

-- INSERT INTO public."ITEMS" (chrt_id,track_number,price,rid,name,sale,size,total_price,nm_id,brand,status)
-- VALUES
--     (9934930,'WBILMTESTTRACK',453,'ab4219087a764ae0btest','Mascaras',30,'0',317,2389212,'Vivienne Sabo',202);
GRANT CONNECT ON DATABASE "WB" TO wb_service;
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA public TO wb_service;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO wb_service;