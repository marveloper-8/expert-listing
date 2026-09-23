package handlers

import (
	"errors"
	"net/http"

	"expertlisting/internal/models"
	"expertlisting/internal/repository"
	"expertlisting/internal/services"
	"expertlisting/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListingHandler struct {
	service services.ListingService
}

func NewListingHandler(service services.ListingService) *ListingHandler {
	return &ListingHandler{service: service}
}

// CreateListing godoc
// @Summary Create a property listing
// @Description Creates a new listing with nested location details, pricing, and agent metadata
// @Tags Listings
// @Accept json
// @Produce json
// @Param listing body models.CreateListingRequest true "Listing payload"
// @Success 201 {object} models.StandardResponse{data=models.Listing}
// @Failure 400 {object} models.StandardResponse
// @Failure 422 {object} models.StandardResponse{errors=[]models.FieldError}
// @Failure 500 {object} models.StandardResponse
// @Router /listings [post]
func (h *ListingHandler) CreateListing(c *gin.Context) {
	var req models.CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fieldErrors := utils.FormatValidationError(err)
		c.JSON(http.StatusUnprocessableEntity, models.NewValidationErrorResponse("Validation failed", fieldErrors))
		return
	}

	listing, err := h.service.CreateListing(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(http.StatusInternalServerError, "Failed to create listing", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(http.StatusCreated, "Listing created successfully", listing))
}

// GetListingByID godoc
// @Summary Fetch a listing by ID
// @Description Returns the listing with matching UUID
// @Tags Listings
// @Produce json
// @Param id path string true "Listing UUID" format(uuid)
// @Success 200 {object} models.StandardResponse{data=models.Listing}
// @Failure 400 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Failure 500 {object} models.StandardResponse
// @Router /listings/{id} [get]
func (h *ListingHandler) GetListingByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(http.StatusBadRequest, "Invalid listing ID format", "id must be a valid UUID"))
		return
	}

	listing, err := h.service.GetListingByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(http.StatusNotFound, "Listing not found", "No property listing matches the provided ID"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch listing", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(http.StatusOK, "Listing retrieved successfully", listing))
}

// GetListings godoc
// @Summary List property listings
// @Description Returns paginated listings with optional sorting
// @Tags Listings
// @Produce json
// @Param page query int false "Page number (default: 1)" default(1) minimum(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param sort_by query string false "Sort field" default(created_at) Enums(created_at, price, bedrooms)
// @Param order query string false "Sort order" default(desc) Enums(asc, desc)
// @Success 200 {object} models.StandardResponse{data=[]models.Listing,pagination=models.PaginationMetadata}
// @Failure 400 {object} models.StandardResponse
// @Failure 500 {object} models.StandardResponse
// @Router /listings [get]
func (h *ListingHandler) GetListings(c *gin.Context) {
	var query models.ListListingsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		fieldErrors := utils.FormatValidationError(err)
		c.JSON(http.StatusUnprocessableEntity, models.NewValidationErrorResponse("Invalid query parameters", fieldErrors))
		return
	}

	listings, pagination, err := h.service.ListListings(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve listings", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewPaginatedResponse(http.StatusOK, "Listings retrieved successfully", listings, pagination))
}

// UpdateListing godoc
// @Summary Update a listing
// @Description Updates fields of an existing property listing
// @Tags Listings
// @Accept json
// @Produce json
// @Param id path string true "Listing UUID" format(uuid)
// @Param listing body models.UpdateListingRequest true "Listing update payload"
// @Success 200 {object} models.StandardResponse{data=models.Listing}
// @Failure 400 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Failure 422 {object} models.StandardResponse{errors=[]models.FieldError}
// @Failure 500 {object} models.StandardResponse
// @Router /listings/{id} [put]
func (h *ListingHandler) UpdateListing(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(http.StatusBadRequest, "Invalid listing ID format", "id must be a valid UUID"))
		return
	}

	var req models.UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fieldErrors := utils.FormatValidationError(err)
		c.JSON(http.StatusUnprocessableEntity, models.NewValidationErrorResponse("Validation failed", fieldErrors))
		return
	}

	listing, err := h.service.UpdateListing(id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(http.StatusNotFound, "Listing not found", "No property listing matches the provided ID"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(http.StatusInternalServerError, "Failed to update listing", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(http.StatusOK, "Listing updated successfully", listing))
}

// PatchListing godoc
// @Summary Partially update a listing
// @Description Updates specific fields of an existing property listing
// @Tags Listings
// @Accept json
// @Produce json
// @Param id path string true "Listing UUID" format(uuid)
// @Param listing body models.UpdateListingRequest true "Partial update payload"
// @Success 200 {object} models.StandardResponse{data=models.Listing}
// @Failure 400 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Failure 422 {object} models.StandardResponse{errors=[]models.FieldError}
// @Failure 500 {object} models.StandardResponse
// @Router /listings/{id} [patch]
func (h *ListingHandler) PatchListing(c *gin.Context) {
	h.UpdateListing(c)
}

// DeleteListing godoc
// @Summary Delete a listing
// @Description Deletes a property listing by UUID
// @Tags Listings
// @Produce json
// @Param id path string true "Listing UUID" format(uuid)
// @Success 200 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Failure 500 {object} models.StandardResponse
// @Router /listings/{id} [delete]
func (h *ListingHandler) DeleteListing(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(http.StatusBadRequest, "Invalid listing ID format", "id must be a valid UUID"))
		return
	}

	err = h.service.DeleteListing(id)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(http.StatusNotFound, "Listing not found", "No property listing matches the provided ID"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(http.StatusInternalServerError, "Failed to delete listing", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(http.StatusOK, "Listing deleted successfully", nil))
}

// SearchListings godoc
// @Summary Search listings by filters and location radius
// @Description Filters by type, price range, bedrooms, and returns listings within radius_km of a given coordinate
// @Tags Listings
// @Produce json
// @Param type query string false "Listing type" Enums(rent, sale, shortlet)
// @Param min_price query number false "Minimum price" minimum(0)
// @Param max_price query number false "Maximum price" minimum(0)
// @Param bedrooms query int false "Exact bedroom count" minimum(0)
// @Param min_bedrooms query int false "Minimum bedroom count" minimum(0)
// @Param city query string false "City name"
// @Param lat query number false "Latitude for radius search (-90 to 90)"
// @Param lng query number false "Longitude for radius search (-180 to 180)"
// @Param radius_km query number false "Radius in kilometers" minimum(0.1)
// @Param page query int false "Page number (default: 1)" default(1) minimum(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param sort_by query string false "Sort field" default(created_at) Enums(created_at, price, bedrooms, distance)
// @Param order query string false "Sort order" default(desc) Enums(asc, desc)
// @Success 200 {object} models.StandardResponse{data=[]models.Listing,pagination=models.PaginationMetadata}
// @Failure 400 {object} models.StandardResponse
// @Failure 422 {object} models.StandardResponse{errors=[]models.FieldError}
// @Failure 500 {object} models.StandardResponse
// @Router /listings/search [get]
func (h *ListingHandler) SearchListings(c *gin.Context) {
	var query models.SearchListingQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		fieldErrors := utils.FormatValidationError(err)
		c.JSON(http.StatusUnprocessableEntity, models.NewValidationErrorResponse("Invalid query parameters", fieldErrors))
		return
	}

	listings, pagination, err := h.service.SearchListings(query)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPriceRange) || errors.Is(err, services.ErrIncompleteGeoParams) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(http.StatusBadRequest, "Invalid search criteria", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(http.StatusInternalServerError, "Search failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewPaginatedResponse(http.StatusOK, "Search results retrieved successfully", listings, pagination))
}
