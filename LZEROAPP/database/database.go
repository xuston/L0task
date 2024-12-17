package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"main.go/models"
)

type DB struct {
	conn *sql.DB
}

func NewDB(user, password, dbname, host string, port int) (*DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable", user, password, dbname, host, port)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Подключение с БД установлено")
	return &DB{conn: db}, nil
}

// Сохраняет информацию о заказе и связанную с ним информацию
func (db *DB) SaveOrder(order models.Orders) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Сохраняет заказ
	_, err = tx.Exec(`
		INSERT INTO orders (order_uid, track_number, entry, locate, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.Order_uid, order.Track_number, order.Entry, order.Locate, order.Internal_signature,
		order.Custoner_id, order.Delivery_service, order.Shardkey, order.Sm_id, order.Date_created, order.Oof_shard,
	)
	if err != nil {
		return err
	}

	// Сохраняет информацию о доставке
	_, err = tx.Exec(`
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.Order_uid, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	// Сохраняет информацию об оплате
	_, err = tx.Exec(`
		INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.Order_uid, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	// Сохраняет информацию о товарах
	for _, item := range order.Items {
		_, err = tx.Exec(`
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			order.Order_uid, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetAllOrders() ([]models.Orders, error) {
	// Запрос для получения всех заказов
	rows, err := db.conn.Query(`
		SELECT order_uid, track_number, entry, locate, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Orders
	for rows.Next() {
		var order models.Orders
		// Преобразует строку в структуру Order
		if err := rows.Scan(
			&order.Order_uid, &order.Track_number, &order.Entry, &order.Locate, &order.Internal_signature,
			&order.Custoner_id, &order.Delivery_service, &order.Shardkey, &order.Sm_id, &order.Date_created, &order.Oof_shard,
		); err != nil {
			return nil, err
		}

		// Получает данные о доставке этого заказа
		err = db.conn.QueryRow(`
			SELECT name, phone, zip, city, address, region, email
			FROM delivery WHERE order_uid = $1`, order.Order_uid).
			Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
		if err != nil {
			return nil, err
		}

		// Получает данные об оплате этого заказа
		err = db.conn.QueryRow(`
			SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
			FROM payment WHERE order_uid = $1`, order.Order_uid).
			Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
		if err != nil {
			return nil, err
		}

		// Получает товары для этого заказа
		rowsItems, err := db.conn.Query(`
			SELECT chrt_id, track_number, price, rid, name, sale, total_price, nm_id, brand, status
			FROM items WHERE order_uid = $1`, order.Order_uid)
		if err != nil {
			return nil, err
		}
		defer rowsItems.Close()

		for rowsItems.Next() {
			var item models.Items
			if err := rowsItems.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
				return nil, err
			}
			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	// Проверка на ошибки при обработке строк
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (db *DB) GetOrderByID(orderID string) (models.Orders, error) {
	var order models.Orders
	// Пример SQL-запроса для получения данных из таблицы orders
	query := `SELECT order_uid, track_number, name, phone, address, city, amount, currency FROM orders WHERE order_uid = $1`
	row := db.conn.QueryRow(query, orderID)

	// Сканируем полученные данные
	err := row.Scan(&order.Order_uid, &order.Track_number, &order.Delivery.Name, &order.Delivery.Phone,
		&order.Delivery.Address, &order.Delivery.City, &order.Payment.Amount, &order.Payment.Currency)
	if err != nil {
		if err == sql.ErrNoRows {
			return order, errors.New("заказ не найден")
		}
		return order, err
	}

	// Можно добавить дополнительные запросы для items
	return order, nil
}
