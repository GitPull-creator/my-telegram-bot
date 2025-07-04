package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    user_id INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS subcategories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    photo_file_id TEXT NOT NULL,
    category_id INTEGER,
    subcategory_id INTEGER,
    user_id INTEGER NOT NULL
);`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("ошибка при создании таблиц: %w", err)
	}

	return nil
}

func CreateUserCategories(db *sql.DB, userID int64) error {
	const checkQuery = `SELECT COUNT(*) FROM categories WHERE user_id = ?`
	var cnt int
	if err := db.QueryRow(checkQuery, userID).Scan(&cnt); err != nil {
		return fmt.Errorf("проверка категорий пользователя: %w", err)
	}

	if cnt > 0 {
		return nil
	}

	return createDefaultCategories(db, userID)
}

func ResetUserCategories(db *sql.DB, userID int64) error {
	// Удаляем все старые категории пользователя
	_, err := db.Exec("DELETE FROM categories WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления категорий: %w", err)
	}

	// Создаем новые категории
	return createDefaultCategories(db, userID)
}

func createDefaultCategories(db *sql.DB, userID int64) error {
	defaults := []string{"Косметика", "Маникюр", "Педикюр"}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO categories (name, user_id) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, name := range defaults {
		_, err := stmt.Exec(name, userID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
