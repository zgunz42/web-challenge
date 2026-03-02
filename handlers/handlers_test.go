package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"coding-challenge/middleware"
	"coding-challenge/store"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
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

func TestPing(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /ping status = %d; want 200", w.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["success"] != true {
		t.Errorf("success = %v; want true", out["success"])
	}
}

func TestEcho(t *testing.T) {
	r := setupRouter()
	body := map[string]string{"hello": "world"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /echo status = %d; want 200", w.Code)
	}
	var out map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["hello"] != "world" {
		t.Errorf("echo body = %v; want {\"hello\":\"world\"}", out)
	}
}

func TestCreateBook(t *testing.T) {
	r := setupRouter()
	body := map[string]string{"title": "The Go Book", "author": "Alice"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("POST /books status = %d; want 201", w.Code)
	}
	var out struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Author string `json:"author"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.ID == 0 || out.Title != "The Go Book" || out.Author != "Alice" {
		t.Errorf("created book = %+v", out)
	}
}

func TestGetBooks(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /books status = %d; want 200", w.Code)
	}
	var out []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out == nil {
		out = []map[string]interface{}{}
	}
}

func TestGetBookByID(t *testing.T) {
	r := setupRouter()
	// Create a book first
	createBody := map[string]string{"title": "Test", "author": "Bob"}
	raw, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/books/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /books/1 status = %d; want 200", w.Code)
	}
	var out struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Author string `json:"author"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != 1 || out.Title != "Test" || out.Author != "Bob" {
		t.Errorf("book = %+v", out)
	}
}

func TestUpdateBook(t *testing.T) {
	r := setupRouter()
	createBody := map[string]string{"title": "Original", "author": "Charlie"}
	raw, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	updateBody := map[string]string{"title": "Updated", "author": "Charlie"}
	raw, _ = json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("PUT /books/1 status = %d; want 200", w.Code)
	}
	var out struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Title != "Updated" {
		t.Errorf("updated title = %q; want \"Updated\"", out.Title)
	}
}

func TestDeleteBook(t *testing.T) {
	r := setupRouter()
	createBody := map[string]string{"title": "To Delete", "author": "Diana"}
	raw, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	req := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("DELETE /books/1 status = %d; want 204", w.Code)
	}
}

func TestAuthToken(t *testing.T) {
	r := setupRouter()
	body := map[string]string{"username": "admin", "password": "password"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /auth/token status = %d; want 200", w.Code)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Token == "" {
		t.Error("token is empty")
	}
}

func TestAuthProtectedRoute(t *testing.T) {
	r := setupRouter()
	// Without auth header -> 200 (optional auth allows through)
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /books without auth status = %d; want 200", w.Code)
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &arr); err != nil {
		t.Fatal(err)
	}
	_ = arr
	// With empty token (Bearer only) -> 401
	reqEmpty := httptest.NewRequest(http.MethodGet, "/books", nil)
	reqEmpty.Header.Set("Authorization", "Bearer ")
	wEmpty := httptest.NewRecorder()
	r.ServeHTTP(wEmpty, reqEmpty)
	if wEmpty.Code != http.StatusUnauthorized {
		t.Errorf("GET /books with empty token status = %d; want 401", wEmpty.Code)
	}
	// With invalid token -> 401
	reqBad := httptest.NewRequest(http.MethodGet, "/books", nil)
	reqBad.Header.Set("Authorization", "Bearer wrong")
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusUnauthorized {
		t.Errorf("GET /books with invalid token status = %d; want 401", wBad.Code)
	}
	// Get a token via POST /auth/token
	body := map[string]string{"username": "admin", "password": "password"}
	raw, _ := json.Marshal(body)
	authReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(raw))
	authReq.Header.Set("Content-Type", "application/json")
	authW := httptest.NewRecorder()
	r.ServeHTTP(authW, authReq)
	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(authW.Body.Bytes(), &tokenResp); err != nil {
		t.Fatal(err)
	}
	// With valid token -> 200
	req2 := httptest.NewRequest(http.MethodGet, "/books", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("GET /books with token status = %d; want 200", w2.Code)
	}
}

func TestSearchByAuthor(t *testing.T) {
	r := setupRouter()
	// Create books
	for _, b := range []struct{ title, author string }{
		{"A", "Author1"},
		{"B", "Author2"},
		{"C", "Author1"},
	} {
		body, _ := json.Marshal(map[string]string{"title": b.title, "author": b.author})
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
	// Get a valid token
	authBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "password"})
	authReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(authBody))
	authReq.Header.Set("Content-Type", "application/json")
	authW := httptest.NewRecorder()
	r.ServeHTTP(authW, authReq)
	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(authW.Body.Bytes(), &tokenResp); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/books?author=Author1", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /books?author=Author1 status = %d; want 200", w.Code)
	}
	var out []struct {
		Author string `json:"author"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	for _, b := range out {
		if b.Author != "Author1" {
			t.Errorf("expected author Author1, got %q", b.Author)
		}
	}
}

func TestPagination(t *testing.T) {
	r := setupRouter()
	// Get a valid token
	authBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "password"})
	authReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(authBody))
	authReq.Header.Set("Content-Type", "application/json")
	authW := httptest.NewRecorder()
	r.ServeHTTP(authW, authReq)
	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(authW.Body.Bytes(), &tokenResp); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/books?page=1&limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /books?page=1&limit=2 status = %d; want 200", w.Code)
	}
	var out []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) > 2 {
		t.Errorf("expected at most 2 items, got %d", len(out))
	}
}

func TestInvalidBookCreation(t *testing.T) {
	r := setupRouter()
	body := []byte(`{"title": "No Author"}`)
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /books invalid status = %d; want 400", w.Code)
	}
}

func TestBookNotFound(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/books/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GET /books/999 status = %d; want 404", w.Code)
	}
}
