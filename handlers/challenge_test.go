package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"coding-challenge/middleware"
	"coding-challenge/store"

	"github.com/gin-gonic/gin"
)

// setupChallengeRouter creates a fresh router for challenge tests
func setupChallengeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	s := store.New()
	booksHandler := NewBooksHandler(s)
	r := gin.New()
	r.GET("/ping", Ping)
	r.POST("/echo", Echo)
	r.POST("/auth/token", AuthToken)
	r.POST("/books", booksHandler.CreateBook)
	r.GET("/books", middleware.AuthOptional, booksHandler.ListBooks)
	r.GET("/books/:id", booksHandler.GetBookByID)
	r.PUT("/books/:id", booksHandler.UpdateBook)
	r.DELETE("/books/:id", booksHandler.DeleteBook)
	return r
}

// Level 1: Ping
func TestLevel1_Ping(t *testing.T) {
	r := setupChallengeRouter()

	t.Run("GET /ping returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("Response contains { success: true }", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if success, ok := result["success"].(bool); !ok || !success {
			t.Errorf("Expected {success: true}, got %v", result)
		}
	})
}

// Level 2: Echo
func TestLevel2_Echo(t *testing.T) {
	r := setupChallengeRouter()

	t.Run("POST /echo returns 200", func(t *testing.T) {
		body := map[string]interface{}{"test": "data"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("Response echoes the sent JSON body", func(t *testing.T) {
		input := map[string]interface{}{
			"message": "hello",
			"nested":  map[string]interface{}{"key": "val"},
			"number":  float64(42),
		}
		raw, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if !reflect.DeepEqual(input, result) {
			t.Errorf("Expected %v, got %v", input, result)
		}
	})

	t.Run("POST /echo with empty object returns {}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Body.String() != "{}" {
			t.Errorf("Expected {}, got %s", w.Body.String())
		}
	})
}

// Level 3: CRUD: Create & Read
func TestLevel3_CRUD_CreateRead(t *testing.T) {
	r := setupChallengeRouter()

	t.Run("POST /books returns 201", func(t *testing.T) {
		body := map[string]interface{}{"title": "Test Book", "author": "Test Author", "year": 2024}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("Expected 201, got %d", w.Code)
		}
	})

	t.Run("Created book has an id", func(t *testing.T) {
		body := map[string]interface{}{"title": "Test Book", "author": "Test Author"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if _, ok := result["id"]; !ok {
			t.Error("Response missing 'id' field")
		}
	})

	t.Run("GET /books returns an array", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		var result []interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Response is not an array: %v", err)
		}
	})

	t.Run("GET /books/:id returns the book", func(t *testing.T) {
		// Create a book first
		body := map[string]interface{}{"title": "Specific Book", "author": "Specific Author"}
		raw, _ := json.Marshal(body)
		createReq := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)
		var created map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &created)
		id := int(created["id"].(float64))

		// Get the book
		req := httptest.NewRequest(http.MethodGet, "/books/"+string(rune(id+48)), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if result["title"] != "Specific Book" {
			t.Errorf("Expected 'Specific Book', got %v", result["title"])
		}
	})
}

// Level 4: CRUD: Update & Delete
func TestLevel4_CRUD_UpdateDelete(t *testing.T) {
	r := setupChallengeRouter()

	// Setup: Create a book for testing
	createBook := func() int {
		body := map[string]interface{}{"title": "Original", "author": "Author"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		return int(result["id"].(float64))
	}

	t.Run("PUT /books/:id returns 200", func(t *testing.T) {
		id := createBook()
		body := map[string]interface{}{"title": "Updated", "author": "Author"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/books/"+string(rune(id+48)), bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("Updated book has new title", func(t *testing.T) {
		id := createBook()
		body := map[string]interface{}{"title": "New Title", "author": "Author"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/books/"+string(rune(id+48)), bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		if result["title"] != "New Title" {
			t.Errorf("Expected 'New Title', got %v", result["title"])
		}
	})

	t.Run("DELETE /books/:id returns 200 or 204", func(t *testing.T) {
		id := createBook()
		req := httptest.NewRequest(http.MethodDelete, "/books/"+string(rune(id+48)), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
			t.Errorf("Expected 200 or 204, got %d", w.Code)
		}
	})

	t.Run("GET deleted book returns 404", func(t *testing.T) {
		id := createBook()
		// Delete
		delReq := httptest.NewRequest(http.MethodDelete, "/books/"+string(rune(id+48)), nil)
		delW := httptest.NewRecorder()
		r.ServeHTTP(delW, delReq)
		// Try to get
		req := httptest.NewRequest(http.MethodGet, "/books/"+string(rune(id+48)), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})
}

// Level 5: Auth Guard
func TestLevel5_AuthGuard(t *testing.T) {
	r := setupChallengeRouter()

	t.Run("POST /auth/token returns 200", func(t *testing.T) {
		body := map[string]string{"username": "admin", "password": "password"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("Response contains a token", func(t *testing.T) {
		body := map[string]string{"username": "admin", "password": "password"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		if _, ok := result["token"]; !ok {
			t.Error("Response missing 'token' field")
		}
	})

	t.Run("GET /books without token returns 401 (with empty Bearer)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", w.Code)
		}
	})

	t.Run("GET /books with token returns 200", func(t *testing.T) {
		// Get token
		body := map[string]string{"username": "admin", "password": "password"}
		raw, _ := json.Marshal(body)
		authReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(raw))
		authReq.Header.Set("Content-Type", "application/json")
		authW := httptest.NewRecorder()
		r.ServeHTTP(authW, authReq)
		var authResult map[string]interface{}
		json.Unmarshal(authW.Body.Bytes(), &authResult)
		token := authResult["token"].(string)

		// Use token
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("Wrong credentials returns 401", func(t *testing.T) {
		body := map[string]string{"username": "wrong", "password": "wrong"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", w.Code)
		}
	})
}

// Level 6: Search & Paginate
func TestLevel6_SearchPaginate(t *testing.T) {
	r := setupChallengeRouter()

	// Setup: Create test books
	books := []map[string]interface{}{
		{"title": "Book A", "author": "Author X"},
		{"title": "Book B", "author": "Author Y"},
		{"title": "Book C", "author": "Author X"},
		{"title": "Book D", "author": "Author Z"},
	}
	for _, book := range books {
		raw, _ := json.Marshal(book)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	t.Run("GET /books?author=X filters by author", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books?author=Author+X", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		var result []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		for _, book := range result {
			if book["author"] != "Author X" {
				t.Errorf("Expected author 'Author X', got %v", book["author"])
			}
		}
	})

	t.Run("GET /books?page=1&limit=2 paginates", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books?page=1&limit=2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		var result []interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		if len(result) > 2 {
			t.Errorf("Expected at most 2 items, got %d", len(result))
		}
	})
}

// Level 7: Error Handling
func TestLevel7_ErrorHandling(t *testing.T) {
	r := setupChallengeRouter()

	t.Run("POST /books without title returns 400", func(t *testing.T) {
		body := map[string]string{"author": "Author"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("POST /books without author returns 400", func(t *testing.T) {
		body := map[string]string{"title": "Title"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("GET /books/nonexistent returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books/nonexistent", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})

	t.Run("PUT /books/nonexistent returns 404", func(t *testing.T) {
		body := map[string]string{"title": "Title", "author": "Author"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/books/nonexistent", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})

	t.Run("DELETE /books/nonexistent returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/books/nonexistent", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})
}

// Level 8: Boss - Speed Run (all previous tests should pass)
func TestLevel8_SpeedRun(t *testing.T) {
	t.Run("All levels", func(t *testing.T) {
		TestLevel1_Ping(t)
		TestLevel2_Echo(t)
		TestLevel3_CRUD_CreateRead(t)
		TestLevel4_CRUD_UpdateDelete(t)
		TestLevel5_AuthGuard(t)
		TestLevel6_SearchPaginate(t)
		TestLevel7_ErrorHandling(t)
	})
}
