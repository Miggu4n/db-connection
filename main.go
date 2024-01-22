package main

import (
	"db-connection/storage"
	"db-connection/types"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/books", r.GetBooks)
	api.Post("/books", r.CreateBook)
	api.Delete("/books/:id", r.DeleteBook)
	api.Get("/books/:id", r.GetBookById)

}

func (r *Repository) CreateBook(ctx *fiber.Ctx) error {
	book := Book{}
	err := ctx.BodyParser(&book)

	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "bad request"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{"message": "book created"})

	return nil
}

func (r *Repository) GetBooks(ctx *fiber.Ctx) error {
	bookModels := &[]types.Books{}

	err := r.DB.Find(bookModels).Error

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "bad request"})
	}

	ctx.Status(http.StatusOK).JSON(
		&fiber.Map{
			"message": "books fetched",
			"data":    bookModels,
		})
	return nil
}

func (r *Repository) GetBookById(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	bookModel := &types.Books{}

	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id can't be empty",
		})

	}

	err := r.DB.Where("id = ?", id).First(bookModel).Error

	if err != nil {

		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get the book",
		})

		return err
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book fetched",
		"data":    bookModel,
	})
	return nil
}

func (r *Repository) DeleteBook(ctx *fiber.Ctx) error {
	bookModel := types.Books{}
	id := ctx.Params("id")

	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id can't be empty",
		})
	}

	err := r.DB.Delete(bookModel, id)
	if err.Error != nil {

		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})

	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book deleted",
	})

	return nil
}

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database", err)
	}

	err = types.MigrateBooks(db)
	if err != nil {
		log.Fatal(err)
	}

	r := Repository{
		DB: db,
	}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
}
