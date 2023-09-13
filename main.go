package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Valgard/godotenv"
	"github.com/basicgodemo/go-react-app/models"
	"github.com/basicgodemo/go-react-app/storage"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Book struct {
	Authors   string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(c *fiber.Ctx) error {
	book := Book{}

	err := c.BodyParser(&book)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"message": "Request fail"})
		return err
	}
	err = r.DB.Create(&book).Error

	if err != nil {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not create a book"})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"message": "book has been added"})
	return nil
}
func (r *Repository) GetBooks(c *fiber.Ctx) error {
	bookModels := &[]models.Books{}
	err := r.DB.Find(bookModels).Error

	if err != nil {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not get books", "data": bookModels})
		return err
	}

	c.Status(http.StatusOK).JSON(&fiber.Map{"message": "book fetch successfully"})
	return nil
}

func (r *Repository) DeleteBook(c *fiber.Ctx) error {
	bookModel := &models.Books{}
	id := c.Params("id")
	if id == "" {
		c.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "id should not be empty"})
		return nil
	}
	err := r.DB.Delete(bookModel, id)
	if err.Error != nil {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not delete book"})
		return err.Error
	}
	return nil
}

func (r *Repository) GetBookByID(c *fiber.Ctx) error {
	bookModel := &models.Books{}
	id := c.Params("id")
	if id == "" {
		c.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "id should not be empty"})
		return nil
	}
	fmt.Println("id is ", id)
	err := r.DB.Where("id = ?", id).First(bookModel).Error

	if err != nil {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not get the book"})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"message": "book id fetch successfully", "data": bookModel})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_books/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookByID)
	api.Get("/get_books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("Could not Load the Database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}
