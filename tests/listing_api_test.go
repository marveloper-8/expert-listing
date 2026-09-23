package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"expertlisting/internal/cache"
	"expertlisting/internal/config"
	"expertlisting/internal/handlers"
	"expertlisting/internal/middleware"
	"expertlisting/internal/models"
	"expertlisting/internal/repository"
	"expertlisting/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Port:        "8080",
		DBDriver:    "sqlite",
		DatabaseURL: fmt.Sprintf("file:test_%s?mode=memory&cache=shared", uuid.New().String()),
		Environment: "test",
	}

	db, err := config.InitDB(cfg)
	require.NoError(t, err)

	memCache := cache.NewMemoryCache()
	listingRepo := repository.NewListingRepository(db)
	listingService := services.NewListingService(listingRepo, memCache)
	listingHandler := handlers.NewListingHandler(listingService)
	healthHandler := handlers.NewHealthHandler()

	router := gin.New()
	router.Use(middleware.PanicRecovery())

	router.GET("/health", healthHandler.Check)

	api := router.Group("/api")
	{
		listings := api.Group("/listings")
		{
			listings.GET("/search", listingHandler.SearchListings)
			listings.GET("", listingHandler.GetListings)
			listings.POST("", listingHandler.CreateListing)
			listings.GET("/:id", listingHandler.GetListingByID)
			listings.PUT("/:id", listingHandler.UpdateListing)
			listings.PATCH("/:id", listingHandler.PatchListing)
			listings.DELETE("/:id", listingHandler.DeleteListing)
		}
	}

	return router
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res models.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
}

func TestCreateListing_Success(t *testing.T) {
	router := setupTestRouter(t)

	payload := models.CreateListingRequest{
		Title:       "Luxury 3-Bed Apartment in Ikoyi",
		Description: "Fully serviced with ocean view",
		Price:       120000000.0,
		Type:        models.ListingTypeSale,
		Bedrooms:    3,
		Bathrooms:   3,
		Location: models.Location{
			Address:   "10 Bourdillon Rd, Ikoyi",
			City:      "Lagos",
			Latitude:  6.4500,
			Longitude: 3.4300,
		},
		AgentID: "agent_778",
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/listings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res models.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)

	dataMap := res.Data.(map[string]interface{})
	assert.NotEmpty(t, dataMap["id"])
	assert.Equal(t, "Luxury 3-Bed Apartment in Ikoyi", dataMap["title"])
	assert.Equal(t, "sale", dataMap["type"])

	locMap := dataMap["location"].(map[string]interface{})
	assert.Equal(t, "10 Bourdillon Rd, Ikoyi", locMap["address"])
	assert.Equal(t, 6.45, locMap["latitude"])
	assert.Equal(t, 3.43, locMap["longitude"])
}

func TestCreateListing_ValidationFailure(t *testing.T) {
	router := setupTestRouter(t)

	payload := map[string]interface{}{
		"price":    -500,
		"type":     "invalid_type",
		"bedrooms": -1,
		"location": map[string]interface{}{
			"latitude":  195.0,
			"longitude": 3.4300,
		},
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/listings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var res models.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.NotEmpty(t, res.Errors)
}

func TestGetListingByID_NotFoundAndInvalidUUID(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/listings/invalid-uuid", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	randomUUID := uuid.New().String()
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/listings/"+randomUUID, nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestCRUDLifecycle(t *testing.T) {
	router := setupTestRouter(t)

	createPayload := models.CreateListingRequest{
		Title:     "Cozy 1-Bedroom Studio in Lekki",
		Price:     3500000.0,
		Type:      models.ListingTypeRent,
		Bedrooms:  1,
		Bathrooms: 1,
		Location: models.Location{
			Address:   "Admiralty Way, Lekki Phase 1",
			City:      "Lagos",
			Latitude:  6.4474,
			Longitude: 3.4723,
		},
		AgentID: "agent_002",
	}

	body, _ := json.Marshal(createPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/listings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createRes models.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &createRes)
	createdID := createRes.Data.(map[string]interface{})["id"].(string)

	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/listings/"+createdID, nil)
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)

	newTitle := "Updated Studio in Lekki Phase 1"
	newPrice := 4000000.0
	newAddr := "15 Admiralty Way, Lekki Phase 1"
	updatePayload := models.UpdateListingRequest{
		Title: &newTitle,
		Price: &newPrice,
		Location: &models.UpdateLocationRequest{
			Address: &newAddr,
		},
	}
	updateBody, _ := json.Marshal(updatePayload)
	wPut := httptest.NewRecorder()
	reqPut, _ := http.NewRequest(http.MethodPut, "/api/listings/"+createdID, bytes.NewBuffer(updateBody))
	reqPut.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPut, reqPut)
	assert.Equal(t, http.StatusOK, wPut.Code)

	wGetUpdated := httptest.NewRecorder()
	reqGetUpdated, _ := http.NewRequest(http.MethodGet, "/api/listings/"+createdID, nil)
	router.ServeHTTP(wGetUpdated, reqGetUpdated)
	assert.Equal(t, http.StatusOK, wGetUpdated.Code)
	var updatedRes models.StandardResponse
	json.Unmarshal(wGetUpdated.Body.Bytes(), &updatedRes)
	updData := updatedRes.Data.(map[string]interface{})
	assert.Equal(t, "Updated Studio in Lekki Phase 1", updData["title"])
	updLoc := updData["location"].(map[string]interface{})
	assert.Equal(t, "15 Admiralty Way, Lekki Phase 1", updLoc["address"])

	wDel := httptest.NewRecorder()
	reqDel, _ := http.NewRequest(http.MethodDelete, "/api/listings/"+createdID, nil)
	router.ServeHTTP(wDel, reqDel)
	assert.Equal(t, http.StatusOK, wDel.Code)

	wGetAfter := httptest.NewRecorder()
	reqGetAfter, _ := http.NewRequest(http.MethodGet, "/api/listings/"+createdID, nil)
	router.ServeHTTP(wGetAfter, reqGetAfter)
	assert.Equal(t, http.StatusNotFound, wGetAfter.Code)
}

func TestSearchAndGeospatialProximity(t *testing.T) {
	router := setupTestRouter(t)

	seedListings := []models.CreateListingRequest{
		{
			Title:     "Ikoyi Waterfront Villa",
			Price:     350000000.0,
			Type:      models.ListingTypeSale,
			Bedrooms:  5,
			Bathrooms: 5,
			Location: models.Location{
				City:      "Lagos",
				Latitude:  6.4520, // Ikoyi (~2.8 km from Victoria Island center)
				Longitude: 3.4350,
			},
			AgentID: "agent_001",
		},
		{
			Title:     "Victoria Island Executive Flat",
			Price:     8500000.0,
			Type:      models.ListingTypeRent,
			Bedrooms:  2,
			Bathrooms: 2,
			Location: models.Location{
				City:      "Lagos",
				Latitude:  6.4281, // VI Center
				Longitude: 3.4219,
			},
			AgentID: "agent_001",
		},
		{
			Title:     "Lekki Phase 1 Shortlet Apartment",
			Price:     80000.0,
			Type:      models.ListingTypeShortlet,
			Bedrooms:  2,
			Bathrooms: 2,
			Location: models.Location{
				City:      "Lagos",
				Latitude:  6.4470, // Lekki (~6 km from VI)
				Longitude: 3.4750,
			},
			AgentID: "agent_002",
		},
		{
			Title:     "Ikeja GRA Family Home",
			Price:     180000000.0,
			Type:      models.ListingTypeSale,
			Bedrooms:  4,
			Bathrooms: 4,
			Location: models.Location{
				City:      "Lagos",
				Latitude:  6.5925, // Ikeja Mainland (~20 km from VI)
				Longitude: 3.3560,
			},
			AgentID: "agent_003",
		},
		{
			Title:     "Abuja Maitama Mansion",
			Price:     900000000.0,
			Type:      models.ListingTypeSale,
			Bedrooms:  6,
			Bathrooms: 7,
			Location: models.Location{
				City:      "Abuja",
				Latitude:  9.0882, // Abuja (~535 km away)
				Longitude: 7.4934,
			},
			AgentID: "agent_004",
		},
	}

	for _, item := range seedListings {
		body, _ := json.Marshal(item)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/listings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	wRent := httptest.NewRecorder()
	reqRent, _ := http.NewRequest(http.MethodGet, "/api/listings/search?type=rent", nil)
	router.ServeHTTP(wRent, reqRent)
	assert.Equal(t, http.StatusOK, wRent.Code)
	var resRent models.StandardResponse
	json.Unmarshal(wRent.Body.Bytes(), &resRent)
	itemsRent := resRent.Data.([]interface{})
	assert.Len(t, itemsRent, 1)

	wPrice := httptest.NewRecorder()
	reqPrice, _ := http.NewRequest(http.MethodGet, "/api/listings/search?min_price=100000000&max_price=400000000", nil)
	router.ServeHTTP(wPrice, reqPrice)
	assert.Equal(t, http.StatusOK, wPrice.Code)
	var resPrice models.StandardResponse
	json.Unmarshal(wPrice.Body.Bytes(), &resPrice)
	assert.Len(t, resPrice.Data.([]interface{}), 2)

	wGeo := httptest.NewRecorder()
	reqGeo, _ := http.NewRequest(http.MethodGet, "/api/listings/search?lat=6.4281&lng=3.4219&radius_km=5.0", nil)
	router.ServeHTTP(wGeo, reqGeo)
	assert.Equal(t, http.StatusOK, wGeo.Code)

	var resGeo models.StandardResponse
	json.Unmarshal(wGeo.Body.Bytes(), &resGeo)
	itemsGeo := resGeo.Data.([]interface{})
	assert.Len(t, itemsGeo, 2)

	firstItem := itemsGeo[0].(map[string]interface{})
	assert.Equal(t, "Victoria Island Executive Flat", firstItem["title"])
	assert.Equal(t, float64(0), firstItem["distance_km"])

	wInvalidGeo := httptest.NewRecorder()
	reqInvalidGeo, _ := http.NewRequest(http.MethodGet, "/api/listings/search?lat=6.4281", nil)
	router.ServeHTTP(wInvalidGeo, reqInvalidGeo)
	assert.Equal(t, http.StatusBadRequest, wInvalidGeo.Code)
}

func TestPaginationMetadata(t *testing.T) {
	router := setupTestRouter(t)

	for i := 1; i <= 15; i++ {
		payload := models.CreateListingRequest{
			Title:     fmt.Sprintf("Apartment #%d", i),
			Price:     float64(i * 1000000),
			Type:      models.ListingTypeRent,
			Bedrooms:  2,
			Bathrooms: 1,
			Location: models.Location{
				Latitude:  6.4000 + float64(i)*0.001,
				Longitude: 3.4000 + float64(i)*0.001,
			},
			AgentID: "agent_batch",
		}
		body, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/listings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	wP1 := httptest.NewRecorder()
	reqP1, _ := http.NewRequest(http.MethodGet, "/api/listings?page=1&limit=5", nil)
	router.ServeHTTP(wP1, reqP1)
	assert.Equal(t, http.StatusOK, wP1.Code)

	var resP1 models.StandardResponse
	json.Unmarshal(wP1.Body.Bytes(), &resP1)
	assert.Len(t, resP1.Data.([]interface{}), 5)
	assert.Equal(t, 1, resP1.Pagination.CurrentPage)
	assert.Equal(t, 5, resP1.Pagination.PageSize)
	assert.Equal(t, int64(15), resP1.Pagination.TotalItems)
	assert.Equal(t, 3, resP1.Pagination.TotalPages)
	assert.True(t, resP1.Pagination.HasNextPage)
	assert.False(t, resP1.Pagination.HasPrevPage)
}
