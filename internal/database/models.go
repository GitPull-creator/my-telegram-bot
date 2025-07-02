package database

type Category struct {
	ID     int
	Name   string
	UserID int64
}

type Subcategory struct {
	ID         int
	Name       string
	CategoryID int
	UserID     int64
}

type Card struct {
	ID            int
	PhotoFileID   string
	Title         string
	Link          string
	CategoryID    int
	SubcategoryID *int
	UserID        int64
}
