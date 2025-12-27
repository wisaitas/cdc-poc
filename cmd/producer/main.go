package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	App *fiber.App
	DB  *gorm.DB
}

func init() {
	if err := godotenv.Load("./domain.env"); err != nil {
		log.Println(err)
	}

	if err := env.Parse(&Config); err != nil {
		log.Fatalln(err)
	}
}

var Config struct {
	Service struct {
		Port string `env:"PORT" envDefault:"8080"`
	} `envPrefix:"SERVICE_"`
	Postgres struct {
		Host     string `env:"HOST" envDefault:"localhost"`
		Port     string `env:"PORT" envDefault:"5432"`
		User     string `env:"USER" envDefault:"postgres"`
		Password string `env:"PASSWORD" envDefault:"postgres"`
		Name     string `env:"NAME" envDefault:"mydb"`
	} `envPrefix:"POSTGRES_"`
}

type Message struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:uuidv7()"`
	Content   string    `gorm:"column:content;type:text;not null"`
	Status    string    `gorm:"column:status;type:text;not null;default:'pending'"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:now()"`
}

func main() {
	app := New()

	app.App.Post("/api/v1/messages", func(c fiber.Ctx) error {
		request := struct {
			Message string `json:"message"`
		}{}
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		message := Message{
			Content: request.Message,
		}

		if err := app.DB.Create(&message).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Message created successfully",
			"data":    message,
		})
	})

	app.Start()

	gracefulShutdownChan := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdownChan, os.Interrupt, syscall.SIGTERM)

	<-gracefulShutdownChan

	app.Shutdown()
}

func New() *App {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Config.Postgres.Host,
		Config.Postgres.Port,
		Config.Postgres.User,
		Config.Postgres.Password,
		Config.Postgres.Name,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if err := db.AutoMigrate(&Message{}); err != nil {
		log.Fatalf("error migrating database: %v", err)
	}

	return &App{
		App: fiber.New(),
		DB:  db,
	}
}

func (a *App) Start() error {
	go func() {
		if err := a.App.Listen(fmt.Sprintf(":%s", Config.Service.Port)); err != nil {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	return nil
}

func (a *App) Shutdown() error {
	return a.App.Shutdown()
}
