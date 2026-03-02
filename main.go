package main

import (
	"log"
	"os"

	"coding-challenge/handlers"
	"coding-challenge/middleware"
	"coding-challenge/store"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.New()
	booksHandler := handlers.NewBooksHandler(s)

	r := gin.Default()

	// Add detailed request/response logging
	r.Use(middleware.RequestLogger)

	r.GET("/ping", handlers.Ping)
	r.POST("/echo", handlers.Echo)
	r.POST("/auth/token", handlers.AuthToken)

	r.POST("/books", booksHandler.CreateBook)
	r.GET("/books", middleware.AuthRequired, booksHandler.ListBooks) // Level 5: requires auth
	r.GET("/books/:id", booksHandler.GetBookByID)
	r.PUT("/books/:id", booksHandler.UpdateBook)
	r.DELETE("/books/:id", booksHandler.DeleteBook)

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
// Force rebuild - Tue Mar  3 00:06:39 WITA 2026
