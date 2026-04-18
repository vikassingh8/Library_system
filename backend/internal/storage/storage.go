package storage

import "github.com/library_system/internal/models"

type Storage interface {
	// User methods
	CreateUser(name string, email string, password string) (int, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	UserExists(id int) (bool, error)

	// Book methods
	CreateBook(book *models.Book) (int, error)
	GetBookByID(id int) (*models.Book, error)
	GetAllBooks() ([]*models.Book, error)
	SearchBooks(genre, author, title string) ([]*models.Book, error)
	UpdateBook(book *models.Book) error
	DeleteBook(id int) error
	UpdateBookCopies(bookID int, availableCopies int) error

	// Borrow methods
	CreateBorrow(userID, bookID int, dueDate string) (int, error)
	BorrowBookAtomic(userID, bookID int, dueDate string) (int, error)
	GetBorrowByID(id int) (*models.Borrow, error)
	GetUserBorrows(userID int) ([]*models.Borrow, error)
	UpdateBorrowReturn(borrowID int, returnedAt string) error
}
