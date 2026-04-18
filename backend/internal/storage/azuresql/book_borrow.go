package azuresql

import (
	"database/sql"
	"fmt"

	"github.com/library_system/internal/models"
)

// ─── Book methods ─────────────────────────────────────────────────────────────

func (a *AzureSQL) CreateBook(book *models.Book) (int, error) {
	var id int
	query := `INSERT INTO books (title, author, isbn, genre, total_copies, available_copies, cover_image_url)
	          OUTPUT INSERTED.id
	          VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`
	err := a.db.QueryRow(query,
		book.Title, book.Author, book.ISBN, book.Genre,
		book.TotalCopies, book.TotalCopies, book.CoverImageURL,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *AzureSQL) GetBookByID(id int) (*models.Book, error) {
	var book models.Book
	query := `SELECT id, title, author, isbn, genre, total_copies, available_copies, cover_image_url, created_at
	          FROM books WHERE id = @p1`
	err := a.db.QueryRow(query, id).Scan(
		&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre,
		&book.TotalCopies, &book.AvailableCopies, &book.CoverImageURL, &book.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (a *AzureSQL) GetAllBooks() ([]*models.Book, error) {
	query := `SELECT id, title, author, isbn, genre, total_copies, available_copies, cover_image_url, created_at
	          FROM books ORDER BY created_at DESC`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(
			&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre,
			&book.TotalCopies, &book.AvailableCopies, &book.CoverImageURL, &book.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, &book)
	}
	return books, nil
}

func (a *AzureSQL) SearchBooks(genre, author, title string) ([]*models.Book, error) {
	// Build parameterized query dynamically for SQL Server
	query := `SELECT id, title, author, isbn, genre, total_copies, available_copies, cover_image_url, created_at
	          FROM books WHERE 1=1`
	args := []interface{}{}
	paramIdx := 1

	if genre != "" {
		query += fmt.Sprintf(" AND genre LIKE @p%d", paramIdx)
		args = append(args, "%"+genre+"%")
		paramIdx++
	}
	if author != "" {
		query += fmt.Sprintf(" AND author LIKE @p%d", paramIdx)
		args = append(args, "%"+author+"%")
		paramIdx++
	}
	if title != "" {
		query += fmt.Sprintf(" AND title LIKE @p%d", paramIdx)
		args = append(args, "%"+title+"%")
		paramIdx++
	}

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(
			&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre,
			&book.TotalCopies, &book.AvailableCopies, &book.CoverImageURL, &book.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, &book)
	}
	return books, nil
}

func (a *AzureSQL) UpdateBook(book *models.Book) error {
	_, err := a.db.Exec(
		`UPDATE books SET title = @p1, author = @p2, isbn = @p3, genre = @p4,
		 total_copies = @p5, available_copies = @p6, cover_image_url = @p7
		 WHERE id = @p8`,
		book.Title, book.Author, book.ISBN, book.Genre,
		book.TotalCopies, book.AvailableCopies, book.CoverImageURL, book.ID,
	)
	return err
}

func (a *AzureSQL) DeleteBook(id int) error {
	_, err := a.db.Exec("DELETE FROM books WHERE id = @p1", id)
	return err
}

func (a *AzureSQL) UpdateBookCopies(bookID int, availableCopies int) error {
	_, err := a.db.Exec(
		"UPDATE books SET available_copies = @p1 WHERE id = @p2",
		availableCopies, bookID,
	)
	return err
}

// ─── Borrow methods ───────────────────────────────────────────────────────────

func (a *AzureSQL) CreateBorrow(userID, bookID int, dueDate string) (int, error) {
	var id int
	query := `INSERT INTO borrows (user_id, book_id, due_date, status)
	          OUTPUT INSERTED.id
	          VALUES (@p1, @p2, @p3, 'borrowed')`
	err := a.db.QueryRow(query, userID, bookID, dueDate).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// BorrowBookAtomic performs an atomic borrow: decrement copies and insert record in one transaction.
func (a *AzureSQL) BorrowBookAtomic(userID, bookID int, dueDate string) (int, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Atomically decrement available_copies only if > 0
	res, err := tx.Exec(
		"UPDATE books SET available_copies = available_copies - 1 WHERE id = @p1 AND available_copies > 0",
		bookID,
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, sql.ErrNoRows // no copies available
	}

	// 2. Insert borrow record and get the new ID
	var borrowID int
	err = tx.QueryRow(
		`INSERT INTO borrows (user_id, book_id, due_date, status)
		 OUTPUT INSERTED.id
		 VALUES (@p1, @p2, @p3, 'borrowed')`,
		userID, bookID, dueDate,
	).Scan(&borrowID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return borrowID, nil
}

func (a *AzureSQL) GetBorrowByID(id int) (*models.Borrow, error) {
	var borrow models.Borrow
	var returnedAt sql.NullTime

	query := `SELECT id, user_id, book_id, borrowed_at, due_date, returned_at, status
	          FROM borrows WHERE id = @p1`
	err := a.db.QueryRow(query, id).Scan(
		&borrow.ID, &borrow.UserID, &borrow.BookID, &borrow.BorrowedAt,
		&borrow.DueDate, &returnedAt, &borrow.Status,
	)
	if err != nil {
		return nil, err
	}

	if returnedAt.Valid {
		borrow.ReturnedAt = &returnedAt.Time
	}
	return &borrow, nil
}

func (a *AzureSQL) GetUserBorrows(userID int) ([]*models.Borrow, error) {
	// Mark any past-due active borrows as overdue before returning results
	_, _ = a.db.Exec(
		`UPDATE borrows SET status = 'overdue'
		 WHERE status = 'borrowed' AND due_date < GETUTCDATE()`,
	)

	query := `
		SELECT
			b.id, b.user_id, b.book_id, b.borrowed_at, b.due_date, b.returned_at, b.status,
			bk.title, bk.author
		FROM borrows b
		JOIN books bk ON b.book_id = bk.id
		WHERE b.user_id = @p1
		ORDER BY b.borrowed_at DESC`

	rows, err := a.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var borrows []*models.Borrow
	for rows.Next() {
		var borrow models.Borrow
		var returnedAt sql.NullTime

		err := rows.Scan(
			&borrow.ID, &borrow.UserID, &borrow.BookID, &borrow.BorrowedAt,
			&borrow.DueDate, &returnedAt, &borrow.Status,
			&borrow.BookTitle, &borrow.BookAuthor,
		)
		if err != nil {
			return nil, err
		}

		if returnedAt.Valid {
			borrow.ReturnedAt = &returnedAt.Time
		}
		borrows = append(borrows, &borrow)
	}
	return borrows, nil
}

func (a *AzureSQL) UpdateBorrowReturn(borrowID int, returnedAt string) error {
	_, err := a.db.Exec(
		`UPDATE borrows SET returned_at = @p1, status = 'returned' WHERE id = @p2`,
		returnedAt, borrowID,
	)
	return err
}
