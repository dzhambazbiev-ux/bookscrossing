package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dasler-fw/bookcrossing/internal/models"
)

const (
	batchSize = 1000

	// Production-like volumes
	genresCount    = 200
	usersCount     = 200000
	booksCount     = 500000
	reviewsCount   = 1000000
	exchangesCount = 1000000
	// book_genres junction: ~4 genres per book on average
	bookGenresCount = 2000000
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Build DSN
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbMode := os.Getenv("DB_SSLMODE")
	if dbMode == "" {
		dbMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPass, dbName, dbPort, dbMode)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations to ensure tables exist
	fmt.Println("Running migrations...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Genre{},
		&models.Exchange{},
		&models.Review{},
	); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	fmt.Println("Migrations completed ✓")

	// Clear existing data (order from most dependent to least dependent)
	fmt.Println("Clearing existing data...")
	// TRUNCATE all tables in correct order (most dependent first)
	// Ignore errors if tables don't exist yet (they will be created by AutoMigrate above)
	_ = db.Exec("TRUNCATE TABLE book_genres, exchanges, reviews, books, users, genres RESTART IDENTITY CASCADE")
	fmt.Println("Data cleared ✓")
	fmt.Println()

	// Initialize fake data generator
	gofakeit.Seed(0)

	// Seed in dependency order
	startTime := time.Now()

	genreIDs := seedGenres(db)
	userIDs := seedUsers(db)
	bookIDs := seedBooks(db, userIDs, genreIDs)
	reviewIDs := seedReviews(db, userIDs, bookIDs)
	exchangeIDs := seedExchanges(db, userIDs, bookIDs)
	seedBookGenres(db, bookIDs, genreIDs)

	elapsed := time.Since(startTime)

	// Print summary
	fmt.Println("\n=== Seeding completed ===")
	fmt.Printf("Genres:     %d\n", len(genreIDs))
	fmt.Printf("Users:      %d\n", len(userIDs))
	fmt.Printf("Books:      %d\n", len(bookIDs))
	fmt.Printf("Reviews:    %d\n", len(reviewIDs))
	fmt.Printf("Exchanges:  %d\n", len(exchangeIDs))
	fmt.Printf("BookGenres: %d\n", bookGenresCount)
	fmt.Printf("\nTotal time: %v\n", elapsed)
}

func seedGenres(db *gorm.DB) []uint {
	const total = genresCount
	genres := make([]models.Genre, 0, total)
	ids := make([]uint, 0, total)

	fmt.Printf("Seeding genres... 0/%d", total)

	// Use a set to ensure unique genre names
	genreNames := make(map[string]bool)
	for i := 0; i < total; i++ {
		var name string
		for {
			name = gofakeit.BookGenre()
			if !genreNames[name] {
				genreNames[name] = true
				break
			}
			// If duplicate, append index
			name = fmt.Sprintf("%s_%d", name, i)
			if !genreNames[name] {
				genreNames[name] = true
				break
			}
		}

		genre := models.Genre{
			Name: name,
		}

		genres = append(genres, genre)
	}

	// Insert all genres at once (small batch)
	err := db.Session(&gorm.Session{SkipHooks: true}).Create(&genres).Error
	if err != nil {
		log.Fatal("Failed to seed genres:", err)
	}

	// Collect IDs after insertion
	for _, g := range genres {
		if g.ID != 0 {
			ids = append(ids, g.ID)
		}
	}

	// If IDs are still not populated, query them from database
	if len(ids) == 0 {
		var dbGenres []models.Genre
		if err := db.Order("id").Find(&dbGenres).Error; err != nil {
			log.Fatal("Failed to query genres:", err)
		}
		for _, g := range dbGenres {
			ids = append(ids, g.ID)
		}
	}

	fmt.Printf("\rSeeding genres... %d/%d", len(ids), total)
	fmt.Println(" ✓")
	return ids
}

func seedUsers(db *gorm.DB) []uint {
	const total = usersCount
	users := make([]models.User, 0, batchSize)
	ids := make([]uint, 0, total)

	fmt.Printf("Seeding users... 0/%d", total)

	deletedCount := total * 5 / 100

	for i := 0; i < total; i++ {
		user := models.User{
			Name:         gofakeit.Name(),
			Email:        fmt.Sprintf("user_%06d@test.com", i),
			PasswordHash: gofakeit.Password(true, true, true, true, false, 32),
			City:         gofakeit.City(),
			Address:      gofakeit.Address().Address,
		}

		// ~5% soft deleted
		if i < deletedCount {
			deletedAt := gofakeit.DateRange(
				time.Now().AddDate(-1, 0, 0),
				time.Now(),
			)
			user.DeletedAt = gorm.DeletedAt{Time: deletedAt, Valid: true}
		}

		users = append(users, user)

		if len(users) >= batchSize {
			db.Session(&gorm.Session{SkipHooks: true}).Create(&users)
			for _, u := range users {
				ids = append(ids, u.ID)
			}
			users = users[:0]
			fmt.Printf("\rSeeding users... %d/%d", i+1, total)
		}
	}

	// Handle remaining
	if len(users) > 0 {
		db.Session(&gorm.Session{SkipHooks: true}).Create(&users)
		for _, u := range users {
			ids = append(ids, u.ID)
		}
	}

	fmt.Println(" ✓")
	return ids
}

func seedBooks(db *gorm.DB, userIDs []uint, genreIDs []uint) []uint {
	const total = booksCount
	books := make([]models.Book, 0, batchSize)
	ids := make([]uint, 0, total)

	fmt.Printf("Seeding books... 0/%d", total)

	statuses := []string{"available", "reserved"}
	deletedCount := total * 5 / 100

	for i := 0; i < total; i++ {
		book := models.Book{
			Title:       gofakeit.BookTitle(),
			Author:      gofakeit.Name(),
			Description: gofakeit.Paragraph(1, 3, 10, " "),
			AISummary:   gofakeit.Paragraph(1, 2, 5, " "),
			Status:      statuses[gofakeit.Number(0, len(statuses)-1)],
			UserID:      userIDs[gofakeit.Number(0, len(userIDs)-1)],
		}

		// ~5% soft deleted
		if i < deletedCount {
			deletedAt := gofakeit.DateRange(
				time.Now().AddDate(-1, 0, 0),
				time.Now(),
			)
			book.DeletedAt = gorm.DeletedAt{Time: deletedAt, Valid: true}
		}

		books = append(books, book)

		if len(books) >= batchSize {
			db.Session(&gorm.Session{SkipHooks: true}).Create(&books)
			for _, b := range books {
				ids = append(ids, b.ID)
			}
			books = books[:0]
			fmt.Printf("\rSeeding books... %d/%d", i+1, total)
		}
	}

	// Handle remaining
	if len(books) > 0 {
		db.Session(&gorm.Session{SkipHooks: true}).Create(&books)
		for _, b := range books {
			ids = append(ids, b.ID)
		}
	}

	fmt.Println(" ✓")
	return ids
}

func seedReviews(db *gorm.DB, userIDs []uint, bookIDs []uint) []uint {
	const total = reviewsCount
	reviews := make([]models.Review, 0, batchSize)
	ids := make([]uint, 0, total)

	fmt.Printf("Seeding reviews... 0/%d", total)

	for i := 0; i < total; i++ {
		authorID := userIDs[gofakeit.Number(0, len(userIDs)-1)]
		targetUserID := userIDs[gofakeit.Number(0, len(userIDs)-1)]
		targetBookID := bookIDs[gofakeit.Number(0, len(bookIDs)-1)]

		review := models.Review{
			AuthorID:     authorID,
			TargetUserID: targetUserID,
			TargetBookID: targetBookID,
			Text:         gofakeit.Paragraph(1, 3, 10, " "),
			Rating:       gofakeit.Number(1, 5),
		}

		reviews = append(reviews, review)

		if len(reviews) >= batchSize {
			db.Session(&gorm.Session{SkipHooks: true}).Create(&reviews)
			for _, r := range reviews {
				ids = append(ids, r.ID)
			}
			reviews = reviews[:0]
			fmt.Printf("\rSeeding reviews... %d/%d", i+1, total)
		}
	}

	// Handle remaining
	if len(reviews) > 0 {
		db.Session(&gorm.Session{SkipHooks: true}).Create(&reviews)
		for _, r := range reviews {
			ids = append(ids, r.ID)
		}
	}

	fmt.Println(" ✓")
	return ids
}

func seedExchanges(db *gorm.DB, userIDs []uint, bookIDs []uint) []uint {
	const total = exchangesCount
	exchanges := make([]models.Exchange, 0, batchSize)
	ids := make([]uint, 0, total)

	fmt.Printf("Seeding exchanges... 0/%d", total)

	statuses := []string{"pending", "accepted", "completed", "cancelled"}
	maxAttempts := 100 // Limit attempts to avoid infinite loops

	for i := 0; i < total; i++ {
		initiatorID := userIDs[gofakeit.Number(0, len(userIDs)-1)]
		recipientID := userIDs[gofakeit.Number(0, len(userIDs)-1)]

		// Ensure initiator and recipient are different (with retry limit)
		attempts := 0
		for recipientID == initiatorID && attempts < maxAttempts {
			recipientID = userIDs[gofakeit.Number(0, len(userIDs)-1)]
			attempts++
		}
		// If still same after max attempts, just use next user
		if recipientID == initiatorID && len(userIDs) > 1 {
			// Find a different user by index
			initiatorIdx := 0
			for idx, uid := range userIDs {
				if uid == initiatorID {
					initiatorIdx = idx
					break
				}
			}
			nextIdx := (initiatorIdx + 1) % len(userIDs)
			recipientID = userIDs[nextIdx]
		}

		initiatorBookID := bookIDs[gofakeit.Number(0, len(bookIDs)-1)]
		recipientBookID := bookIDs[gofakeit.Number(0, len(bookIDs)-1)]

		// Ensure books are different (with retry limit)
		attempts = 0
		for recipientBookID == initiatorBookID && attempts < maxAttempts {
			recipientBookID = bookIDs[gofakeit.Number(0, len(bookIDs)-1)]
			attempts++
		}
		// If still same after max attempts, just use next book
		if recipientBookID == initiatorBookID && len(bookIDs) > 1 {
			// Find a different book by index
			initiatorBookIdx := 0
			for idx, bid := range bookIDs {
				if bid == initiatorBookID {
					initiatorBookIdx = idx
					break
				}
			}
			nextBookIdx := (initiatorBookIdx + 1) % len(bookIDs)
			recipientBookID = bookIDs[nextBookIdx]
		}

		status := statuses[gofakeit.Number(0, len(statuses)-1)]

		exchange := models.Exchange{
			InitiatorID:     initiatorID,
			RecipientID:     recipientID,
			InitiatorBookID: initiatorBookID,
			RecipientBookID: recipientBookID,
			Status:          status,
		}

		// Set CompletedAt for completed exchanges
		if status == "completed" {
			completedAt := gofakeit.DateRange(
				time.Now().AddDate(-1, 0, 0),
				time.Now(),
			)
			exchange.CompletedAt = &completedAt
		}

		exchanges = append(exchanges, exchange)

		if len(exchanges) >= batchSize {
			err := db.Session(&gorm.Session{SkipHooks: true}).Create(&exchanges).Error
			if err != nil {
				log.Fatalf("Failed to seed exchanges at batch %d: %v", i/batchSize, err)
			}
			for _, e := range exchanges {
				ids = append(ids, e.ID)
			}
			exchanges = exchanges[:0]
			fmt.Printf("\rSeeding exchanges... %d/%d", i+1, total)
		}
	}

	// Handle remaining
	if len(exchanges) > 0 {
		err := db.Session(&gorm.Session{SkipHooks: true}).Create(&exchanges).Error
		if err != nil {
			log.Fatal("Failed to seed remaining exchanges:", err)
		}
		for _, e := range exchanges {
			ids = append(ids, e.ID)
		}
		fmt.Printf("\rSeeding exchanges... %d/%d", total, total)
	}

	fmt.Println(" ✓")
	return ids
}

func seedBookGenres(db *gorm.DB, bookIDs []uint, genreIDs []uint) {
	const total = bookGenresCount
	fmt.Printf("Seeding book_genres... 0/%d", total)

	// Create a map to track existing book-genre pairs to avoid duplicates
	bookGenreMap := make(map[string]bool)
	pairs := make([]struct {
		BookID  uint
		GenreID uint
	}, 0, batchSize)

	for i := 0; i < total; i++ {
		bookID := bookIDs[gofakeit.Number(0, len(bookIDs)-1)]
		genreID := genreIDs[gofakeit.Number(0, len(genreIDs)-1)]

		// Create unique key for book-genre pair
		key := fmt.Sprintf("%d-%d", bookID, genreID)

		// Skip if pair already exists
		if bookGenreMap[key] {
			i-- // Retry this iteration
			continue
		}

		bookGenreMap[key] = true
		pairs = append(pairs, struct {
			BookID  uint
			GenreID uint
		}{BookID: bookID, GenreID: genreID})

		if len(pairs) >= batchSize {
			// Build batch insert query for PostgreSQL
			query := "INSERT INTO book_genres (book_id, genre_id) VALUES "
			args := make([]interface{}, 0, len(pairs)*2)

			for idx, pair := range pairs {
				if idx > 0 {
					query += ", "
				}
				query += fmt.Sprintf("($%d, $%d)", idx*2+1, idx*2+2)
				args = append(args, pair.BookID, pair.GenreID)
			}

			query += " ON CONFLICT DO NOTHING"
			db.Exec(query, args...)
			pairs = pairs[:0]
			fmt.Printf("\rSeeding book_genres... %d/%d", i+1, total)
		}
	}

	// Handle remaining
	if len(pairs) > 0 {
		query := "INSERT INTO book_genres (book_id, genre_id) VALUES "
		args := make([]interface{}, 0, len(pairs)*2)

		for idx, pair := range pairs {
			if idx > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d)", idx*2+1, idx*2+2)
			args = append(args, pair.BookID, pair.GenreID)
		}

		query += " ON CONFLICT DO NOTHING"
		db.Exec(query, args...)
	}

	fmt.Println(" ✓")
}
