package azuresql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/library_system/internal/config"
	"github.com/library_system/internal/models"
	_ "github.com/microsoft/go-mssqldb"
)

type AzureSQL struct {
	db *sql.DB
}

// New creates a connection to Azure SQL Database and ensures the schema exists.
func New(cfg *config.Config) (*AzureSQL, error) {
	connString := fmt.Sprintf(
		"server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=true;TrustServerCertificate=false;",
		cfg.Database.Server,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Retry connection with exponential backoff
	maxRetries := 5
	baseDelay := 2 * time.Second
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("error connecting to Azure SQL after %d retries: %w", maxRetries, err)
			}
			delay := time.Duration(1<<uint(i)) * baseDelay // exponential backoff
			time.Sleep(delay)
			continue
		}
		break
	}

	store := &AzureSQL{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return store, nil
}

// migrate creates the required tables if they don't exist.
// Uses SQL Server / Azure SQL syntax (IDENTITY instead of AUTOINCREMENT, NVARCHAR, etc.)
func (a *AzureSQL) migrate() error {
	stmts := []string{
		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='users' AND xtype='U')
		CREATE TABLE users (
			id         INT IDENTITY(1,1) PRIMARY KEY,
			name       NVARCHAR(255) NOT NULL,
			email      NVARCHAR(255) NOT NULL UNIQUE,
			password   NVARCHAR(255) NOT NULL,
			role       NVARCHAR(20)  NOT NULL DEFAULT 'user',
			created_at DATETIME2     NOT NULL,
			updated_at DATETIME2     NULL
		)`,

		// Add role column to existing tables that pre-date this migration
		`IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID(N'users') AND name = N'role')
		ALTER TABLE users ADD role NVARCHAR(20) NOT NULL DEFAULT 'user'`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='books' AND xtype='U')
		CREATE TABLE books (
			id               INT IDENTITY(1,1) PRIMARY KEY,
			title            NVARCHAR(500) NOT NULL,
			author           NVARCHAR(255) NOT NULL,
			isbn             NVARCHAR(50)  UNIQUE,
			genre            NVARCHAR(100),
			total_copies     INT           NOT NULL DEFAULT 1,
			available_copies INT           NOT NULL DEFAULT 1,
			cover_image_url  NVARCHAR(1000),
			created_at       DATETIME2     DEFAULT GETUTCDATE()
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='borrows' AND xtype='U')
		CREATE TABLE borrows (
			id          INT IDENTITY(1,1) PRIMARY KEY,
			user_id     INT           NOT NULL REFERENCES users(id),
			book_id     INT           NOT NULL REFERENCES books(id),
			borrowed_at DATETIME2     DEFAULT GETUTCDATE(),
			due_date    DATETIME2     NOT NULL,
			returned_at DATETIME2     NULL,
			status      NVARCHAR(20)  NOT NULL DEFAULT 'borrowed'
		)`,
	}

	for _, stmt := range stmts {
		if _, err := a.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ─── User methods ────────────────────────────────────────────────────────────

func (a *AzureSQL) CreateUser(name, email, password string) (int, error) {
	// SQL Server does not support LastInsertId via result; use OUTPUT INSERTED.id
	var id int
	query := `INSERT INTO users (name, email, password, created_at)
	          OUTPUT INSERTED.id
	          VALUES (@p1, @p2, @p3, @p4)`
	err := a.db.QueryRow(query, name, email, password, time.Now().UTC()).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *AzureSQL) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, name, email, password, role, created_at FROM users WHERE email = @p1`
	err := a.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *AzureSQL) GetUserByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, name, email, password, role, created_at FROM users WHERE id = @p1`
	err := a.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *AzureSQL) UserExists(id int) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM users WHERE id = @p1`
	err := a.db.QueryRow(query, id).Scan(&count)
	return count > 0, err
}
