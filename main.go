package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

//
// 1. Абстракции (интерфейсы) - RepositoryWriter и Notifier
//

// RepositoryWriter - интерфейс для записи заказов в хранилище (ISP)
type RepositoryWriter interface {
	SaveOrder(customer string, products string, total float64) error
}

// Notifier - базовый интерфейс для отправки уведомлений (LSP)
type Notifier interface {
	Send(customer string) error
}

//
// 2. Адаптация существующих структур под интерфейсы
//

// EmailSender реализует интерфейс Notifier
type EmailSender struct{}

func (e *EmailSender) Send(customer string) error {
	fmt.Printf("[Email] Уведомление отправлено клиенту %s\n", customer)
	return nil
}

//
// 3. Новая функциональность - SMS Sender
//

// SMSSender также реализует интерфейс Notifier
type SMSSender struct{}

func (s *SMSSender) Send(customer string) error {
	fmt.Printf("[SMS] Уведомление отправлено клиенту %s\n", customer)
	return nil
}

// Реализация RepositoryWriter для SQLite
type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) (*SQLiteRepo, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer TEXT NOT NULL,
			products TEXT NOT NULL,
			total REAL NOT NULL,
			status TEXT NOT NULL
		)`)
	if err != nil {
		return nil, err
	}
	return &SQLiteRepo{db: db}, nil
}

func (r *SQLiteRepo) SaveOrder(customer string, products string, total float64) error {
	_, err := r.db.Exec(
		"INSERT INTO orders (customer, products, total, status) VALUES (?, ?, ?, ?)",
		customer, products, total, "pending",
	)
	return err
}

//
// 4. Рефакторинг основного сервиса
//    Теперь он зависит от интерфейсов, а не от конкретных реализаций (DIP)
//

type OrderService struct {
	repo     RepositoryWriter
	notifier Notifier
}

func NewOrderService(repo RepositoryWriter, notifier Notifier) *OrderService {
	return &OrderService{
		repo:     repo,
		notifier: notifier,
	}
}

func (s *OrderService) CreateOrder(customer string, products []string, total float64) error {
	productsStr := fmt.Sprintf("%v", products)

	if err := s.repo.SaveOrder(customer, productsStr, total); err != nil {
		return err
	}

	return s.notifier.Send(customer)
}

//
// 5. Демонстрация работы с разными отправителями
//

func main() {
	db, err := sql.Open("sqlite3", "orders.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo, err := NewSQLiteRepo(db)
	if err != nil {
		log.Fatal(err)
	}

	// Демонстрация с EmailSender
	emailNotifier := &EmailSender{}
	orderServiceEmail := NewOrderService(repo, emailNotifier)

	err = orderServiceEmail.CreateOrder("Иван", []string{"apple", "banana"}, 10.5)
	if err != nil {
		log.Fatal(err)
	}

	// Демонстрация с SMSSender
	smsNotifier := &SMSSender{}
	orderServiceSMS := NewOrderService(repo, smsNotifier)

	err = orderServiceSMS.CreateOrder("Пётр", []string{"orange", "milk"}, 20.0)
	if err != nil {
		log.Fatal(err)
	}
}
