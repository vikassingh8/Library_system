package models

import "time"

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Role      string `json:"role"` // "user" | "admin"
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Book struct {
	ID              int       `json:"id"`
	Title           string    `json:"title" validate:"required"`
	Author          string    `json:"author" validate:"required"`
	ISBN            string    `json:"isbn" validate:"required"`
	Genre           string    `json:"genre"`
	TotalCopies     int       `json:"total_copies" validate:"required,min=1"`
	AvailableCopies int       `json:"available_copies"`
	CoverImageURL   string    `json:"cover_image_url"`
	CreatedAt       time.Time `json:"created_at"`
}

type Borrow struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	BookID     int        `json:"book_id"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueDate    time.Time  `json:"due_date"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
	Status     string     `json:"status"` // borrowed, returned, overdue
	BookTitle  string     `json:"book_title,omitempty"`
	BookAuthor string     `json:"book_author,omitempty"`
}
