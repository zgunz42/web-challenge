package models

// Book represents a book entity.
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year,omitempty"`
}

// CreateBookInput is the request body for creating a book.
type CreateBookInput struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
	Year   int    `json:"year"`
}

// UpdateBookInput is the request body for updating a book (all fields optional for partial updates).
// Year is a pointer so we can distinguish "not provided" (nil) from "explicitly set to 0".
type UpdateBookInput struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   *int   `json:"year"`
}
