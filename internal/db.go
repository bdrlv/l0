package internal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(connstring string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connstring)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func saveOrder(ctx context.Context, db *sql.DB, order Order) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("начало транзакции провалилось: %w. ", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING
	`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("ошибка сохранения заказа: %w. ", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO NOTHING
	`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("ошибка сохранения данных доставки: %w. ", err)
	}

	paymentTime := time.Unix(order.Payment.PaymentDt, 0)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING
	`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		paymentTime, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("ошибка сохранения данных оплаты: %w. ", err)
	}

	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, `
        INSERT INTO items (chrt_id, name, size, nm_id, brand)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (chrt_id) DO NOTHING
    	`, item.ChrtID, item.Name, item.Size, item.NmID, item.Brand,
		)
		if err != nil {
			return fmt.Errorf("ошибка сохранения товаров заказа: %w. ", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO order_items (
				order_uid, chrt_id, track_number, price, sale, total_price, rid, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (order_uid, chrt_id, rid) DO NOTHING 
		`,
			order.OrderUID, item.ChrtID, item.TrackNumber,
			item.Price, item.Sale, item.TotalPrice,
			item.Rid, item.Status,
		)
		if err != nil {
			return fmt.Errorf("ошибка вставки в order_items: %w. ", err)
		}
	}

	return tx.Commit()
}

func getOrderByIdFromDB(ctx context.Context, db *sql.DB, order_id string) (Order, error) {
	query := `
        SELECT
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
            d.name AS delivery_name,
            d.phone AS delivery_phone,
            d.zip AS delivery_zip,
            d.city AS delivery_city,
            d.address AS delivery_address,
            d.region AS delivery_region,
            d.email AS delivery_email,
            p.transaction,
            p.request_id,
            p.currency,
            p.provider,
            p.amount,
            p.payment_dt,
            p.bank,
            p.delivery_cost,
            p.goods_total,
            p.custom_fee,
            i.chrt_id,
            i.name AS item_name,
            i.size AS item_size,
            i.nm_id,
            i.brand,
            oi.track_number AS item_track_number,
            oi.price,
            oi.sale,
            oi.total_price AS item_total_price,
            oi.rid,
            oi.status
        FROM orders o
        LEFT JOIN delivery d ON o.order_uid = d.order_uid
        LEFT JOIN payment p ON o.order_uid = p.order_uid
        LEFT JOIN order_items oi ON o.order_uid = oi.order_uid
        LEFT JOIN items i ON oi.chrt_id = i.chrt_id
        WHERE o.order_uid = $1
        ORDER BY i.chrt_id, oi.rid;
    `

	rows, err := db.QueryContext(ctx, query, order_id)
	if err != nil {
		return Order{}, fmt.Errorf("ошибка выполнения запроса: %w. ", err)
	}
	defer rows.Close()

	var order Order
	var firstRow = true

	for rows.Next() {
		var r dbRow
		err := rows.Scan(
			&r.OrderUID, &r.TrackNumber, &r.Entry, &r.Locale, &r.InternalSignature,
			&r.CustomerID, &r.DeliveryService, &r.Shardkey, &r.SmID, &r.DateCreated, &r.OofShard,
			&r.DeliveryName, &r.DeliveryPhone, &r.DeliveryZip, &r.DeliveryCity,
			&r.DeliveryAddress, &r.DeliveryRegion, &r.DeliveryEmail,
			&r.Transaction, &r.RequestID, &r.Currency, &r.Provider,
			&r.Amount, &r.PaymentDt, &r.Bank, &r.DeliveryCost, &r.GoodsTotal, &r.CustomFee,
			&r.ChrtID, &r.ItemName, &r.ItemSize, &r.NmID, &r.Brand,
			&r.ItemTrackNumber, &r.Price, &r.Sale, &r.ItemTotalPrice, &r.Rid, &r.Status,
		)
		if err != nil {
			return Order{}, fmt.Errorf("ошибка сканирования строки: %w. ", err)
		}
		if firstRow {
			order.OrderUID = r.OrderUID
			order.TrackNumber = r.TrackNumber
			order.Entry = r.Entry
			order.Locale = r.Locale
			order.InternalSignature = nullStringOrEmpty(r.InternalSignature)
			order.CustomerID = r.CustomerID
			order.DeliveryService = r.DeliveryService
			order.Shardkey = r.Shardkey
			order.SmID = r.SmID
			order.DateCreated = r.DateCreated.Format(time.RFC3339)
			order.OofShard = r.OofShard
			order.Delivery = struct {
				Name    string `json:"name"`
				Phone   string `json:"phone"`
				Zip     string `json:"zip"`
				City    string `json:"city"`
				Address string `json:"address"`
				Region  string `json:"region"`
				Email   string `json:"email"`
			}{
				Name:    nullStringOrEmpty(r.DeliveryName),
				Phone:   nullStringOrEmpty(r.DeliveryPhone),
				Zip:     nullStringOrEmpty(r.DeliveryZip),
				City:    nullStringOrEmpty(r.DeliveryCity),
				Address: nullStringOrEmpty(r.DeliveryAddress),
				Region:  nullStringOrEmpty(r.DeliveryRegion),
				Email:   nullStringOrEmpty(r.DeliveryEmail),
			}
			order.Payment = struct {
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
			}{
				Transaction:  nullStringOrEmpty(r.Transaction),
				RequestID:    nullStringOrEmpty(r.RequestID),
				Currency:     nullStringOrEmpty(r.Currency),
				Provider:     nullStringOrEmpty(r.Provider),
				Amount:       int(nullFloat64OrZero(r.Amount)),
				PaymentDt:    r.PaymentDt.Time.Unix(),
				Bank:         nullStringOrEmpty(r.Bank),
				DeliveryCost: int(nullFloat64OrZero(r.DeliveryCost)),
				GoodsTotal:   int(nullFloat64OrZero(r.GoodsTotal)),
				CustomFee:    int(nullFloat64OrZero(r.CustomFee)),
			}

			firstRow = false
		}

		if r.ChrtID.Valid {
			item := struct {
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
			}{
				ChrtID:      int(r.ChrtID.Int64),
				TrackNumber: nullStringOrEmpty(r.ItemTrackNumber),
				Price:       int(nullFloat64OrZero(r.Price)),
				Rid:         nullStringOrEmpty(r.Rid),
				Name:        nullStringOrEmpty(r.ItemName),
				Sale:        int(nullInt64OrZero(r.Sale)),
				Size:        nullStringOrEmpty(r.ItemSize),
				TotalPrice:  int(nullFloat64OrZero(r.ItemTotalPrice)),
				NmID:        int(r.NmID.Int64),
				Brand:       nullStringOrEmpty(r.Brand),
				Status:      int(r.Status.Int64),
			}
			order.Items = append(order.Items, item)
		}
	}

	if err = rows.Err(); err != nil {
		return Order{}, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	if firstRow {
		return Order{}, fmt.Errorf("заказ с order_uid=%s не найден", order_id)
	}

	return order, nil
}

func getAlllOrders(ctx context.Context, db *sql.DB) ([]Order, error) {
	query := `
	SELECT
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

		d.name AS delivery_name,
		d.phone AS delivery_phone,
		d.zip AS delivery_zip,
		d.city AS delivery_city,
		d.address AS delivery_address,
		d.region AS delivery_region,
		d.email AS delivery_email,

		p.transaction,
		p.request_id,
		p.currency,
		p.provider,
		p.amount,
		p.payment_dt,
		p.bank,
		p.delivery_cost,
		p.goods_total,
		p.custom_fee,

		oi.chrt_id,
		i.name AS item_name,
		i.size,
		i.nm_id,
		i.brand,
		oi.track_number AS item_track_number,
		oi.price,
		oi.sale,
		oi.total_price AS item_total_price,
		oi.rid,
		oi.status
	FROM orders o
	LEFT JOIN delivery d ON o.order_uid = d.order_uid
	LEFT JOIN payment p ON o.order_uid = p.order_uid
	LEFT JOIN order_items oi ON o.order_uid = oi.order_uid
	LEFT JOIN items i ON oi.chrt_id = i.chrt_id
	ORDER BY o.order_uid;
`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	ordersMap := make(map[string]Order)

	for rows.Next() {
		var r dbRow
		err := rows.Scan(
			&r.OrderUID, &r.TrackNumber, &r.Entry, &r.Locale, &r.InternalSignature,
			&r.CustomerID, &r.DeliveryService, &r.Shardkey, &r.SmID, &r.DateCreated, &r.OofShard,
			&r.DeliveryName, &r.DeliveryPhone, &r.DeliveryZip, &r.DeliveryCity, &r.DeliveryAddress,
			&r.DeliveryRegion, &r.DeliveryEmail,
			&r.Transaction, &r.RequestID, &r.Currency, &r.Provider, &r.Amount, &r.PaymentDt,
			&r.Bank, &r.DeliveryCost, &r.GoodsTotal, &r.CustomFee,
			&r.ChrtID, &r.ItemName, &r.ItemSize, &r.NmID, &r.Brand, &r.ItemTrackNumber,
			&r.Price, &r.Sale, &r.ItemTotalPrice, &r.Rid, &r.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}

		order, exists := ordersMap[r.OrderUID]
		if !exists {
			order = Order{
				OrderUID:          r.OrderUID,
				TrackNumber:       r.TrackNumber,
				Entry:             r.Entry,
				Locale:            r.Locale,
				InternalSignature: nullStringOrEmpty(r.InternalSignature),
				CustomerID:        r.CustomerID,
				DeliveryService:   r.DeliveryService,
				Shardkey:          r.Shardkey,
				SmID:              r.SmID,
				DateCreated:       r.DateCreated.Format(time.RFC3339),
				OofShard:          r.OofShard,
				Items: make([]struct {
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
				}, 0),
			}

			if r.DeliveryName.Valid {
				order.Delivery = struct {
					Name    string `json:"name"`
					Phone   string `json:"phone"`
					Zip     string `json:"zip"`
					City    string `json:"city"`
					Address string `json:"address"`
					Region  string `json:"region"`
					Email   string `json:"email"`
				}{
					Name:    r.DeliveryName.String,
					Phone:   nullStringOrEmpty(r.DeliveryPhone),
					Zip:     nullStringOrEmpty(r.DeliveryZip),
					City:    nullStringOrEmpty(r.DeliveryCity),
					Address: nullStringOrEmpty(r.DeliveryAddress),
					Region:  nullStringOrEmpty(r.DeliveryRegion),
					Email:   nullStringOrEmpty(r.DeliveryEmail),
				}
			}

			if r.Transaction.Valid {
				order.Payment = struct {
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
				}{
					Transaction:  r.Transaction.String,
					RequestID:    nullStringOrEmpty(r.RequestID),
					Currency:     nullStringOrEmpty(r.Currency),
					Provider:     nullStringOrEmpty(r.Provider),
					Amount:       int(r.Amount.Float64),
					PaymentDt:    r.PaymentDt.Time.Unix(),
					Bank:         nullStringOrEmpty(r.Bank),
					DeliveryCost: int(nullFloat64OrZero(r.DeliveryCost)),
					GoodsTotal:   int(nullFloat64OrZero(r.GoodsTotal)),
					CustomFee:    int(nullFloat64OrZero(r.CustomFee)),
				}
			}

			ordersMap[r.OrderUID] = order
		}

		if r.ChrtID.Valid {
			order.Items = append(order.Items, struct {
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
			}{
				ChrtID:      int(r.ChrtID.Int64),
				TrackNumber: nullStringOrEmpty(r.ItemTrackNumber),
				Price:       int(r.Price.Float64),
				Rid:         nullStringOrEmpty(r.Rid),
				Name:        nullStringOrEmpty(r.ItemName),
				Sale:        int(r.Sale.Int64),
				Size:        nullStringOrEmpty(r.ItemSize),
				TotalPrice:  int(r.ItemTotalPrice.Float64),
				NmID:        int(r.NmID.Int64),
				Brand:       nullStringOrEmpty(r.Brand),
				Status:      int(r.Status.Int64),
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерирования по строкам: %w", err)
	}

	var allOrders []Order
	for _, order := range ordersMap {
		allOrders = append(allOrders, order)
	}

	return allOrders, nil
}
func nullStringOrEmpty(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}
func nullFloat64OrZero(f sql.NullFloat64) float64 {
	if f.Valid {
		return f.Float64
	}
	return 0
}
func nullInt64OrZero(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}
