package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"main.go/cache"
	"main.go/database"
	"main.go/kafka"
	"main.go/models"
)

var db *database.DB
var c *cache.Cache

// Обрабатывает запросы для отображения заказа по ID
func OrderPageHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		http.Error(w, "Заказа с таким id не существует", http.StatusBadRequest)
		return
	}

	// Получает заказ из кэша или БД
	order, found := c.Get(orderID)
	if !found {
		var err error
		order, err = db.GetOrderByID(orderID)
		if err != nil {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		c.Set(orderID, order)
	}

	// Загружает HTML-шаблон
	tmpl, err := template.ParseFiles("templates/order.html")
	if err != nil {
		http.Error(w, "Не удалось загрузить шаблон", http.StatusInternalServerError)
		return
	}

	// Отображает данные в шаблоне
	if err := tmpl.Execute(w, order); err != nil {
		http.Error(w, "Не удалось отобразить шаблон", http.StatusInternalServerError)
	}
}

func main() {
	// Подключение к базе данных
	var err error
	db, err = database.NewDB("xuston", "123", "l0base", "localhost", 5432)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	// Инициализация кэша
	c = cache.NewCache()

	// Загрузка данных из базы в кэш
	orders, err := db.GetAllOrders()
	if err != nil {
		log.Println("Не удалось загрузить заказ из базы данных:", err)
	}
	c.LoadFromDB(orders)

	// Запуск Kafka консюмер
	go kafka.Consume(context.Background(), "orders", "localhost:9092", func(order models.Orders) {
		if err := db.SaveOrder(order); err != nil {
			log.Println("Не удалось сохранить заказ:", err)
		}
		c.Save(order)
	})

	// HTTP-сервер
	http.HandleFunc("/order", OrderPageHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
