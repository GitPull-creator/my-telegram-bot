package storage

import (
	"database/sql"
	"my-telegram-bot/internal/database"
)

func GetCategoryID(db *sql.DB, userID int64, categoryName string) (int, error) {
	var categoryID int
	err := db.QueryRow("SELECT id FROM categories WHERE user_id = ? AND name = ?", userID, categoryName).Scan(&categoryID)
	if err != nil {
		return 0, err
	}
	return categoryID, nil
}

func GetUserCategories(db *sql.DB, userID int64) ([]database.Category, error) {
	query := "SELECT id, name, user_id FROM categories WHERE user_id = ?"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []database.Category
	for rows.Next() {
		var category database.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.UserID); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func AddCard(db *sql.DB, card *database.Card) error {
	query := "INSERT INTO cards (photo_file_id, title, link, category_id, subcategory_id, user_id) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := db.Exec(query, card.PhotoFileID, card.Title, card.Link, card.CategoryID, card.SubcategoryID, card.UserID)
	if err != nil {
		return err
	}
	return nil
}

func GetCategoryCards(db *sql.DB, userID int64, categoryID int) ([]database.Card, error) {
	query := "SELECT id, photo_file_id, title, link, category_id, subcategory_id, user_id FROM cards WHERE user_id = ? AND category_id = ?"
	rows, err := db.Query(query, userID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []database.Card

	for rows.Next() {
		var card database.Card
		if err := rows.Scan(&card.ID, &card.PhotoFileID, &card.Title, &card.Link, &card.CategoryID, &card.SubcategoryID, &card.UserID); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func GetCategoryByID(db *sql.DB, userID int64, categoryID int) (database.Category, error) {
	query := "SELECT id, name, user_id FROM categories WHERE user_id = ? AND id = ?"
	var category database.Category
	err := db.QueryRow(query, userID, categoryID).Scan(&category.ID, &category.Name, &category.UserID)
	if err != nil {
		return database.Category{}, err
	}
	return category, nil
}

func AddSubcategory(db *sql.DB, subcategory *database.Subcategory) error {
	query := "INSERT INTO subcategories (name, category_id, user_id) VALUES (?, ?, ?)"
	_, err := db.Exec(query, subcategory.Name, subcategory.CategoryID, subcategory.UserID)
	return err
}

func GetSubcategories(db *sql.DB, userID int64, categoryID int) ([]database.Subcategory, error) {
	query := "SELECT id, name, category_id, user_id FROM subcategories WHERE user_id = ? AND category_id = ?"
	rows, err := db.Query(query, userID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subcategories []database.Subcategory
	for rows.Next() {
		var subcategory database.Subcategory
		if err := rows.Scan(&subcategory.ID, &subcategory.Name, &subcategory.CategoryID, &subcategory.UserID); err != nil {
			return nil, err
		}
		subcategories = append(subcategories, subcategory)
	}
	return subcategories, nil
}
