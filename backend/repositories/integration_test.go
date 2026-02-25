package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/Dailiduzhou/library_manage_sys/models"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	// Import testcontainers-go mysql module
	mysqlmodule "github.com/testcontainers/testcontainers-go/modules/mysql"
)

// setupTestDB sets up a MySQL testcontainer and returns a GORM DB connection
func setupTestDB(t *testing.T) (*mysqlmodule.MySQLContainer, *gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()

	// Create MySQL container using the module
	mysqlContainer, err := mysqlmodule.Run(ctx, "mysql:8.0",
		mysqlmodule.WithDatabase("library_test"),
		mysqlmodule.WithUsername("testuser"),
		mysqlmodule.WithPassword("testpass"),
	)
	require.NoError(t, err, "Failed to start MySQL container")

	// Get connection string
	connStr, err := mysqlContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Ensure MySQL DATETIME/TIMESTAMP columns are scanned into time.Time.
	dsnCfg, err := mysqlDriver.ParseDSN(connStr)
	require.NoError(t, err, "Failed to parse DSN from testcontainer")
	dsnCfg.ParseTime = true
	dsnCfg.Loc = time.Local
	connStr = dsnCfg.FormatDSN()

	// Connect to database
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Run migrations
	err = db.AutoMigrate(&models.Book{}, &models.User{}, &models.BorrowRecord{})
	require.NoError(t, err, "Failed to run migrations")

	// Cleanup function
	cleanup := func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return mysqlContainer, db, cleanup
}

// TestBookRepositoryIntegration_Create tests creating a book in the repository
func TestBookRepositoryIntegration_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGormBookRepository(db)

	// Create test book
	book := &models.Book{
		Title:        "Go Programming",
		Author:       "John Doe",
		Summary:      "A comprehensive guide to Go",
		CoverPath:    "uploads/go.png",
		InitialStock: 10,
		Stock:        10,
		TotalStock:   10,
	}

	// Execute
	err := repo.CreateBook(book)

	// Assert
	require.NoError(t, err)
	assert.NotZero(t, book.ID, "Book ID should be auto-generated")
	assert.Equal(t, "Go Programming", book.Title)
}

// TestBookRepositoryIntegration_GetByID tests retrieving a book by ID
func TestBookRepositoryIntegration_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGormBookRepository(db)

	// Create test book first
	book := &models.Book{
		Title:        "Go Programming",
		Author:       "John Doe",
		Summary:      "A comprehensive guide",
		CoverPath:    "uploads/go.png",
		InitialStock: 5,
		Stock:        5,
		TotalStock:   5,
	}
	err := repo.CreateBook(book)
	require.NoError(t, err)

	// Execute - retrieve by ID
	retrievedBook, err := repo.GetBookByID(book.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, book.ID, retrievedBook.ID)
	assert.Equal(t, "Go Programming", retrievedBook.Title)
	assert.Equal(t, "John Doe", retrievedBook.Author)
}

// TestBookRepositoryIntegration_Update tests updating a book
func TestBookRepositoryIntegration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGormBookRepository(db)

	// Create test book
	book := &models.Book{
		Title:        "Original Title",
		Author:       "Original Author",
		Summary:      "Original Summary",
		CoverPath:    "uploads/original.png",
		InitialStock: 5,
		Stock:        5,
		TotalStock:   5,
	}
	err := repo.CreateBook(book)
	require.NoError(t, err)

	// Update the book
	book.Title = "Updated Title"
	book.Author = "Updated Author"
	book.Stock = 3

	err = repo.UpdateBook(book)
	require.NoError(t, err)

	// Retrieve and verify
	retrievedBook, err := repo.GetBookByID(book.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrievedBook.Title)
	assert.Equal(t, "Updated Author", retrievedBook.Author)
	assert.Equal(t, 3, retrievedBook.Stock)
}

// TestBookRepositoryIntegration_Delete tests deleting a book
func TestBookRepositoryIntegration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGormBookRepository(db)

	// Create test book
	book := &models.Book{
		Title:        "To Be Deleted",
		Author:       "Author",
		Summary:      "Summary",
		CoverPath:    "uploads/delete.png",
		InitialStock: 5,
		Stock:        5,
		TotalStock:   5,
	}
	err := repo.CreateBook(book)
	require.NoError(t, err)
	bookID := book.ID

	// Delete the book
	err = repo.DeleteBook(bookID)
	require.NoError(t, err)

	// Try to retrieve - should fail
	_, err = repo.GetBookByID(bookID)
	assert.Error(t, err, "Expected error when retrieving deleted book")
}

// TestBookRepositoryIntegration_FindAll tests finding all books
func TestBookRepositoryIntegration_FindAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGormBookRepository(db)

	// Create multiple books
	books := []*models.Book{
		{
			Title:        "Book One",
			Author:       "Author A",
			Summary:      "Summary 1",
			CoverPath:    "uploads/book1.png",
			InitialStock: 5,
			Stock:        5,
			TotalStock:   5,
		},
		{
			Title:        "Book Two",
			Author:       "Author B",
			Summary:      "Summary 2",
			CoverPath:    "uploads/book2.png",
			InitialStock: 3,
			Stock:        3,
			TotalStock:   3,
		},
		{
			Title:        "Book Three",
			Author:       "Author A",
			Summary:      "Summary 3",
			CoverPath:    "uploads/book3.png",
			InitialStock: 7,
			Stock:        7,
			TotalStock:   7,
		},
	}

	for _, b := range books {
		err := repo.CreateBook(b)
		require.NoError(t, err)
	}

	// Execute - find all
	allBooks, err := repo.ListBooks()

	// Assert
	require.NoError(t, err)
	assert.Len(t, allBooks, 3, "Should find all 3 books")
}

// TestBookRepositoryIntegration_FindByTitle tests finding books by title
func TestBookRepositoryIntegration_FindByTitle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGormBookRepository(db)

	// Create books with different titles
	books := []*models.Book{
		{
			Title:        "Go Programming Guide",
			Author:       "Author A",
			Summary:      "Summary 1",
			CoverPath:    "uploads/go.png",
			InitialStock: 5,
			Stock:        5,
			TotalStock:   5,
		},
		{
			Title:        "Python Programming Guide",
			Author:       "Author B",
			Summary:      "Summary 2",
			CoverPath:    "uploads/python.png",
			InitialStock: 3,
			Stock:        3,
			TotalStock:   3,
		},
		{
			Title:        "Go Web Development",
			Author:       "Author A",
			Summary:      "Summary 3",
			CoverPath:    "uploads/goweb.png",
			InitialStock: 7,
			Stock:        7,
			TotalStock:   7,
		},
	}

	for _, b := range books {
		err := repo.CreateBook(b)
		require.NoError(t, err)
	}

	// Execute - find by title
	foundBooks, err := repo.FindBooksByTitle("Go")

	// Assert
	require.NoError(t, err)
	assert.Len(t, foundBooks, 2, "Should find 2 books with 'Go' in title")
}

// TestBookRepositoryIntegration_FindByAuthor tests finding books by author
func TestBookRepositoryIntegration_FindByAuthor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGormBookRepository(db)

	// Create books with different authors
	books := []*models.Book{
		{
			Title:        "Book One",
			Author:       "John Doe",
			Summary:      "Summary 1",
			CoverPath:    "uploads/book1.png",
			InitialStock: 5,
			Stock:        5,
			TotalStock:   5,
		},
		{
			Title:        "Book Two",
			Author:       "Jane Smith",
			Summary:      "Summary 2",
			CoverPath:    "uploads/book2.png",
			InitialStock: 3,
			Stock:        3,
			TotalStock:   3,
		},
		{
			Title:        "Book Three",
			Author:       "John Doe",
			Summary:      "Summary 3",
			CoverPath:    "uploads/book3.png",
			InitialStock: 7,
			Stock:        7,
			TotalStock:   7,
		},
	}

	for _, b := range books {
		err := repo.CreateBook(b)
		require.NoError(t, err)
	}

	// Execute - find by author
	foundBooks, err := repo.FindBooksByAuthor("John")

	// Assert
	require.NoError(t, err)
	assert.Len(t, foundBooks, 2, "Should find 2 books by 'John'")
}
