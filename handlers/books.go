package handlers

import (
	"net/http"
	"strconv"

	"coding-challenge/models"
	"coding-challenge/store"

	"github.com/gin-gonic/gin"
)

// BooksHandler holds the store for book operations.
type BooksHandler struct {
	Store *store.Store
}

// NewBooksHandler creates a BooksHandler.
func NewBooksHandler(s *store.Store) *BooksHandler {
	return &BooksHandler{Store: s}
}

// CreateBook handles POST /books (Level 3, 7 error handling).
func (h *BooksHandler) CreateBook(c *gin.Context) {
	var input models.CreateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	book := &models.Book{Title: input.Title, Author: input.Author, Year: input.Year}
	created := h.Store.Create(book)
	c.JSON(http.StatusCreated, created)
}

// ListBooks handles GET /books with optional author filter and pagination (Levels 3, 5, 6).
func (h *BooksHandler) ListBooks(c *gin.Context) {
	author := c.Query("author")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))
	list := h.Store.List(author, page, limit)
	if list == nil {
		list = []models.Book{}
	}
	c.JSON(http.StatusOK, list)
}

// GetBookByID handles GET /books/:id (Level 3, 7 not found).
func (h *BooksHandler) GetBookByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	book := h.Store.GetByID(id)
	if book == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	c.JSON(http.StatusOK, book)
}

// UpdateBook handles PUT /books/:id (Level 4). Supports partial updates: only non-empty fields override existing values.
func (h *BooksHandler) UpdateBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	existing := h.Store.GetByID(id)
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	var input models.UpdateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	title, author, year := existing.Title, existing.Author, existing.Year
	if input.Title != "" {
		title = input.Title
	}
	if input.Author != "" {
		author = input.Author
	}
	if input.Year != nil {
		year = *input.Year
	}
	if !h.Store.Update(id, title, author, year) {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	book := h.Store.GetByID(id)
	if book == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "book state inconsistent"})
		return
	}
	c.JSON(http.StatusOK, book)
}

// DeleteBook handles DELETE /books/:id (Level 4).
func (h *BooksHandler) DeleteBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	if !h.Store.Delete(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
