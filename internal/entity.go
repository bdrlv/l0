package internal

import (
	"database/sql"
	"time"
)

type Order struct {
	OrderUID    string `json:"order_uid"`
	TrackNumber string `json:"track_number"`
	Entry       string `json:"entry"`
	Delivery    struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     string `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	} `json:"delivery"`
	Payment struct {
		Transaction  string `json:"transaction"`
		RequestID    string `json:"request_id"`
		Currency     string `json:"currency"`
		Provider     string `json:"provider"`
		Amount       int    `json:"amount"`
		PaymentDt    int64  `json:"payment_dt"`
		Bank         string `json:"bank"`
		DeliveryCost int    `json:"delivery_cost"`
		GoodsTotal   int    `json:"goods_total"`
		CustomFee    int    `json:"custom_fee"`
	} `json:"payment"`
	Items []struct {
		ChrtID      int    `json:"chrt_id"`
		TrackNumber string `json:"track_number"`
		Price       int    `json:"price"`
		Rid         string `json:"rid"`
		Name        string `json:"name"`
		Sale        int    `json:"sale"`
		Size        string `json:"size"`
		TotalPrice  int    `json:"total_price"`
		NmID        int    `json:"nm_id"`
		Brand       string `json:"brand"`
		Status      int    `json:"status"`
	} `json:"items"`
	Locale            string `json:"locale"`
	InternalSignature string `json:"internal_signature"`
	CustomerID        string `json:"customer_id"`
	DeliveryService   string `json:"delivery_service"`
	Shardkey          string `json:"shardkey"`
	SmID              int    `json:"sm_id"`
	DateCreated       string `json:"date_created"`
	OofShard          string `json:"oof_shard"`
}

type dbRow struct {
	OrderUID          string
	TrackNumber       string
	Entry             string
	Locale            string
	InternalSignature sql.NullString
	CustomerID        string
	DeliveryService   string
	Shardkey          string
	SmID              int
	DateCreated       time.Time
	OofShard          string
	DeliveryName      sql.NullString
	DeliveryPhone     sql.NullString
	DeliveryZip       sql.NullString
	DeliveryCity      sql.NullString
	DeliveryAddress   sql.NullString
	DeliveryRegion    sql.NullString
	DeliveryEmail     sql.NullString
	Transaction       sql.NullString
	RequestID         sql.NullString
	Currency          sql.NullString
	Provider          sql.NullString
	Amount            sql.NullFloat64
	PaymentDt         sql.NullTime
	Bank              sql.NullString
	DeliveryCost      sql.NullFloat64
	GoodsTotal        sql.NullFloat64
	CustomFee         sql.NullFloat64
	ChrtID            sql.NullInt64
	ItemName          sql.NullString
	ItemSize          sql.NullString
	NmID              sql.NullInt64
	Brand             sql.NullString
	ItemTrackNumber   sql.NullString
	Price             sql.NullFloat64
	Sale              sql.NullInt64
	ItemTotalPrice    sql.NullFloat64
	Rid               sql.NullString
	Status            sql.NullInt64
}
