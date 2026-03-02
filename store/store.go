package store

import (
	"coding-challenge/models"
	"sync"
)

// Store is a thread-safe in-memory store for books.
type Store struct {
	mu    sync.RWMutex
	books map[int]*models.Book
	nextID int
}

// New creates a new Store.
func New() *Store {
	return &Store{
		books:  make(map[int]*models.Book),
		nextID: 1,
	}
}

// Create adds a book and returns it with assigned ID.
func (s *Store) Create(book *models.Book) *models.Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	book.ID = s.nextID
	s.nextID++
	s.books[book.ID] = book
	// Return a copy so caller cannot mutate stored data
	cp := *book
	return &cp
}

// GetByID returns a book by ID or nil if not found.
func (s *Store) GetByID(id int) *models.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.books[id]
	if !ok {
		return nil
	}
	cp := *b
	return &cp
}

// List returns all books. Optional filter by author; page and limit for pagination.
// author empty means no filter. page 1-based; limit 0 means no limit.
func (s *Store) List(author string, page, limit int) []models.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []models.Book
	for _, b := range s.books {
		if author != "" && b.Author != author {
			continue
		}
		list = append(list, *b)
	}
	// Simple deterministic order by ID
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].ID < list[i].ID {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if page < 1 {
		page = 1
	}
	if limit > 0 {
		start := (page - 1) * limit
		end := start + limit
		if start >= len(list) {
			return []models.Book{}
		}
		if end > len(list) {
			end = len(list)
		}
		list = list[start:end]
	}
	return list
}

// Update updates a book by ID. Returns true if found and updated.
func (s *Store) Update(id int, title, author string, year int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.books[id]
	if !ok {
		return false
	}
	b.Title = title
	b.Author = author
	b.Year = year
	return true
}

// Delete removes a book by ID. Returns true if found and deleted.
func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return false
	}
	delete(s.books, id)
	return true
}
