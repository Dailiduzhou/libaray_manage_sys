package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"github.com/Dailiduzhou/library_manage_sys/services/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestGetBooksSuccess tests the GetBooks endpoint with successful retrieval
func TestGetBooksSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock book service
	mockBookService := mocks.NewMockBookService(ctrl)

	// Create handler with mock service
	handler := NewBookHandler(mockBookService)

	// Prepare expected books
	expectedBooks := []models.Book{
		{
			ID:         1,
			Title:      "Go Programming",
			Author:     "John Doe",
			Summary:    "A comprehensive guide to Go",
			CoverPath:  "uploads/go.png",
			Stock:      5,
			TotalStock: 10,
		},
		{
			ID:         2,
			Title:      "Go Web Development",
			Author:     "Jane Smith",
			Summary:    "Building web apps with Go",
			CoverPath:  "uploads/goweb.png",
			Stock:      3,
			TotalStock: 5,
		},
	}

	// Expect GetBooks to be called with empty filters
	mockBookService.EXPECT().
		GetBooks("", "", "").
		Return(expectedBooks, nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books", nil)

	// Call handler
	handler.GetBooks(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "查询成功", response.Msg)

	// Verify data
	data, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))
}

// TestGetBooksWithFilters tests the GetBooks endpoint with query filters
func TestGetBooksWithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookService := mocks.NewMockBookService(ctrl)
	handler := NewBookHandler(mockBookService)

	// Expected filtered results
	expectedBooks := []models.Book{
		{
			ID:         1,
			Title:      "Go Programming",
			Author:     "John Doe",
			Summary:    "Programming guide",
			CoverPath:  "uploads/go.png",
			Stock:      5,
			TotalStock: 10,
		},
	}

	// Expect GetBooks to be called with filters
	mockBookService.EXPECT().
		GetBooks("Go", "John", "").
		Return(expectedBooks, nil)

	// Create request with query parameters
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books?title=Go&author=John", nil)

	// Call handler
	handler.GetBooks(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.Code)
}

// TestGetBooksError tests the GetBooks endpoint when service returns error
func TestGetBooksError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookService := mocks.NewMockBookService(ctrl)
	handler := NewBookHandler(mockBookService)

	// Expect GetBooks to return error
	mockBookService.EXPECT().
		GetBooks("", "", "").
		Return(nil, assert.AnError)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books", nil)

	// Call handler
	handler.GetBooks(c)

	// Assert error response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 500, response.Code)
	assert.Equal(t, "数据库查询失败", response.Msg)
}

// TestGetBooksEmpty tests the GetBooks endpoint when no books are found
func TestGetBooksEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookService := mocks.NewMockBookService(ctrl)
	handler := NewBookHandler(mockBookService)

	// Expect empty result
	mockBookService.EXPECT().
		GetBooks("", "", "").
		Return([]models.Book{}, nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books", nil)

	// Call handler
	handler.GetBooks(c)

	// Assert successful but empty response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.Code)

	// Empty slice should be returned
	data, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(data))
}
