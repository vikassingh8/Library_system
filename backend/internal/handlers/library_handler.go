package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/library_system/internal/models"
	"github.com/library_system/internal/storage"
	"github.com/library_system/internal/utils/response"
)

// CreateBook creates a new book in the catalog
func CreateBook(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var book models.Book
		err := json.NewDecoder(r.Body).Decode(&book)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("invalid request body"))
			return
		}

		// Set available copies equal to total copies initially
		book.AvailableCopies = book.TotalCopies

		bookID, err := db.CreateBook(&book)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		response.ApiResponse(w, http.StatusCreated, map[string]interface{}{
			"message": "Book created successfully",
			"book_id": bookID,
		})
	})
}

// GetAllBooks returns all books in the catalog
func GetAllBooks(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		books, err := db.GetAllBooks()
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]interface{}{
			"books": books,
			"count": len(books),
		})
	})
}

// GetBookByID returns a specific book by ID
func GetBookByID(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookID := r.PathValue("id")
		id, err := strconv.Atoi(bookID)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("invalid book ID"))
			return
		}

		book, err := db.GetBookByID(id)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusNotFound, errors.New("book not found"))
			return
		}

		response.ApiResponse(w, http.StatusOK, book)
	})
}

// SearchBooks searches books by genre or author
func SearchBooks(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		genre := r.URL.Query().Get("genre")
		author := r.URL.Query().Get("author")
		title := r.URL.Query().Get("title")

		books, err := db.SearchBooks(genre, author, title)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]interface{}{
			"books": books,
			"count": len(books),
		})
	})
}

// UpdateBook updates an existing book
func UpdateBook(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookID := r.PathValue("id")
		id, err := strconv.Atoi(bookID)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("invalid book ID"))
			return
		}

		var book models.Book
		err = json.NewDecoder(r.Body).Decode(&book)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("invalid request body"))
			return
		}

		book.ID = id
		err = db.UpdateBook(&book)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]string{
			"message": "Book updated successfully",
		})
	})
}

// DeleteBook deletes a book from the catalog
func DeleteBook(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookID := r.PathValue("id")
		id, err := strconv.Atoi(bookID)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("invalid book ID"))
			return
		}

		err = db.DeleteBook(id)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]string{
			"message": "Book deleted successfully",
		})
	})
}

// BorrowBook allows a user to borrow a book
func BorrowBook(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)
		bookID := r.PathValue("id")
		id, err := strconv.Atoi(bookID)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("invalid book ID"))
			return
		}

		// Check if book exists first (optional, but good for specific error message)
		_, err = db.GetBookByID(id)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusNotFound, errors.New("book not found"))
			return
		}

		// Calculate due date (14 days loan period)
		dueDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02 15:04:05")

		// Perform atomic borrow
		borrowID, err := db.BorrowBookAtomic(userId, id, dueDate)
		if err != nil {
			if err == sql.ErrNoRows {
				response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("book not available"))
			} else {
				response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			}
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]interface{}{
			"message":   "Book borrowed successfully",
			"borrow_id": borrowID,
			"due_date":  dueDate,
		})
	})
}

// ReturnBook allows a user to return a borrowed book
func ReturnBook(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)
		borrowID := r.PathValue("id")
		id, err := strconv.Atoi(borrowID)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("invalid borrow ID"))
			return
		}

		// Get borrow record
		borrow, err := db.GetBorrowByID(id)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusNotFound, errors.New("borrow record not found"))
			return
		}

		// Verify the borrow belongs to this user
		if borrow.UserID != userId {
			response.ApiErrorResponse(w, http.StatusForbidden, errors.New("unauthorized"))
			return
		}

		// Check if already returned
		if borrow.ReturnedAt != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("book already returned"))
			return
		}

		// Mark as returned
		returnedAt := time.Now().Format("2006-01-02 15:04:05")
		err = db.UpdateBorrowReturn(id, returnedAt)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		// Update available copies
		book, err := db.GetBookByID(borrow.BookID)
		if err == nil {
			db.UpdateBookCopies(borrow.BookID, book.AvailableCopies+1)
		}

		response.ApiResponse(w, http.StatusOK, map[string]string{
			"message": "Book returned successfully",
		})
	})
}

// GetMyBorrows returns all borrows for the authenticated user
func GetMyBorrows(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		borrows, err := db.GetUserBorrows(userId)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]interface{}{
			"borrows": borrows,
			"count":   len(borrows),
		})
	})
}
