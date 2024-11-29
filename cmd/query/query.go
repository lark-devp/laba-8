package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq" // подключение пакета для работы с PostgreSQL
)

// Константы для подключения к базе данных
const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "lark_dev"
	dbPassword = "Annapetrovna2005"
	dbName     = "query"
)

// User представляет пользователя в системе
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DatabaseProvider содержит соединение с базой данных
type DatabaseProvider struct {
	db *sql.DB
}

// NewDatabaseProvider создает новый экземпляр DatabaseProvider
func NewDatabaseProvider() (*DatabaseProvider, error) {
	// Формируем строку подключения
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName) // sslmode будет проверять подлинность сервера, проверяя цепочку доверия до корневого сертификата,

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DatabaseProvider{db: db}, nil
}

// InsertUser добавляет нового пользователя в базу данных
func (dp *DatabaseProvider) InsertUser(name string) (int, error) {
	var id int
	err := dp.db.QueryRow("INSERT INTO users(name) VALUES($1) RETURNING id", name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUser извлекает пользователя из базы данных по ID
func (dp *DatabaseProvider) GetUser(id int) (User, error) {
	var user User
	err := dp.db.QueryRow("SELECT id, name FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// addUserHandler обрабатывает добавление пользователя
func (dp *DatabaseProvider) addUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}

	id, err := dp.InsertUser(user.Name)
	if err != nil {
		http.Error(w, "Не удалось добавить пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Создан пользователь с ID: %d", id)
}

// getUserHandler обрабатывает извлечение пользователя по ID
func (dp *DatabaseProvider) getUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Отсутствует ID пользователя", http.StatusBadRequest)
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id) //Функция fmt.Sscanf() в языке Go сканирует указанную строку и сохраняет последовательные значения, разделенные пробелами, в последовательные аргументы, как определено форматом.
	if err != nil {
		http.Error(w, "Некорректный формат ID", http.StatusBadRequest)
		return
	}

	user, err := dp.GetUser(id)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func main() {
	// Подключение к базе данных
	dbProvider, err := NewDatabaseProvider()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err) //используется для записи сообщений об ошибках в резервный журнал и завершения работы программы.
	}
	defer dbProvider.db.Close()

	// Регистрация обработчиков
	http.HandleFunc("/api/user/post", dbProvider.addUserHandler)
	http.HandleFunc("/api/user/get", dbProvider.getUserHandler)

	// Запуск сервера

	log.Println("Сервер запущен на порту 9000...")
	err = http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
