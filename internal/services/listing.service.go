package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"expertlisting/internal/cache"
	"expertlisting/internal/models"
	"expertlisting/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrInvalidPriceRange   = errors.New("min_price cannot be greater than max_price")
	ErrIncompleteGeoParams = errors.New("lat, lng, and radius_km must all be provided together for proximity search")
)

type ListingService interface {
	CreateListing(req *models.CreateListingRequest) (*models.Listing, error)
	GetListingByID(id uuid.UUID) (*models.Listing, error)
	UpdateListing(id uuid.UUID, req *models.UpdateListingRequest) (*models.Listing, error)
	DeleteListing(id uuid.UUID) error
	ListListings(query models.ListListingsQuery) ([]models.Listing, *models.PaginationMetadata, error)
	SearchListings(query models.SearchListingQuery) ([]models.Listing, *models.PaginationMetadata, error)
}

type listingService struct {
	repo  repository.ListingRepository
	cache *cache.MemoryCache
}

func NewListingService(repo repository.ListingRepository, cache *cache.MemoryCache) ListingService {
	return &listingService{
		repo:  repo,
		cache: cache,
	}
}

func (s *listingService) CreateListing(req *models.CreateListingRequest) (*models.Listing, error) {
	listing := &models.Listing{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Type:        req.Type,
		Bedrooms:    req.Bedrooms,
		Bathrooms:   req.Bathrooms,
		Location: models.Location{
			Address:   req.Location.Address,
			City:      req.Location.City,
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
		},
		AgentID: req.AgentID,
	}

	if err := s.repo.Create(listing); err != nil {
		return nil, err
	}

	s.cache.Flush()
	return listing, nil
}

func (s *listingService) GetListingByID(id uuid.UUID) (*models.Listing, error) {
	cacheKey := fmt.Sprintf("listing:%s", id.String())
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*models.Listing), nil
	}

	listing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, listing, 2*time.Minute)
	return listing, nil
}

func (s *listingService) UpdateListing(id uuid.UUID, req *models.UpdateListingRequest) (*models.Listing, error) {
	listing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		listing.Title = *req.Title
	}
	if req.Description != nil {
		listing.Description = *req.Description
	}
	if req.Price != nil {
		listing.Price = *req.Price
	}
	if req.Type != nil {
		listing.Type = *req.Type
	}
	if req.Bedrooms != nil {
		listing.Bedrooms = *req.Bedrooms
	}
	if req.Bathrooms != nil {
		listing.Bathrooms = *req.Bathrooms
	}
	if req.Location != nil {
		if req.Location.Address != nil {
			listing.Location.Address = *req.Location.Address
		}
		if req.Location.City != nil {
			listing.Location.City = *req.Location.City
		}
		if req.Location.Latitude != nil {
			listing.Location.Latitude = *req.Location.Latitude
		}
		if req.Location.Longitude != nil {
			listing.Location.Longitude = *req.Location.Longitude
		}
	}
	if req.AgentID != nil {
		listing.AgentID = *req.AgentID
	}

	if err := s.repo.Update(listing); err != nil {
		return nil, err
	}

	s.cache.Flush()
	return listing, nil
}

func (s *listingService) DeleteListing(id uuid.UUID) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	s.cache.Flush()
	return nil
}

type listCacheEntry struct {
	Listings   []models.Listing
	Pagination *models.PaginationMetadata
}

func (s *listingService) ListListings(query models.ListListingsQuery) ([]models.Listing, *models.PaginationMetadata, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 10
	}

	cacheKey := fmt.Sprintf("list:%d:%d:%s:%s", query.Page, query.Limit, query.SortBy, query.Order)
	if cached, found := s.cache.Get(cacheKey); found {
		entry := cached.(*listCacheEntry)
		return entry.Listings, entry.Pagination, nil
	}

	listings, total, err := s.repo.List(query)
	if err != nil {
		return nil, nil, err
	}

	pagination := buildPagination(query.Page, query.Limit, total)
	s.cache.Set(cacheKey, &listCacheEntry{Listings: listings, Pagination: pagination}, 1*time.Minute)
	return listings, pagination, nil
}

func (s *listingService) SearchListings(query models.SearchListingQuery) ([]models.Listing, *models.PaginationMetadata, error) {
	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		return nil, nil, ErrInvalidPriceRange
	}

	hasLat := query.Lat != nil
	hasLng := query.Lng != nil
	hasRadius := query.RadiusKm != nil

	if (hasLat || hasLng || hasRadius) && !(hasLat && hasLng && hasRadius) {
		return nil, nil, ErrIncompleteGeoParams
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 10
	}

	cacheKey := fmt.Sprintf("search:%v:%v:%v:%v:%v:%v:%v:%v:%v:%d:%d:%s:%s",
		valOrEmpty(query.Type),
		valOrZero(query.MinPrice),
		valOrZero(query.MaxPrice),
		valOrZero(query.Bedrooms),
		valOrZero(query.MinBedrooms),
		valOrEmpty(query.City),
		valOrZero(query.Lat),
		valOrZero(query.Lng),
		valOrZero(query.RadiusKm),
		query.Page,
		query.Limit,
		query.SortBy,
		query.Order,
	)

	if cached, found := s.cache.Get(cacheKey); found {
		entry := cached.(*listCacheEntry)
		return entry.Listings, entry.Pagination, nil
	}

	listings, total, err := s.repo.Search(query)
	if err != nil {
		return nil, nil, err
	}

	pagination := buildPagination(query.Page, query.Limit, total)
	s.cache.Set(cacheKey, &listCacheEntry{Listings: listings, Pagination: pagination}, 1*time.Minute)
	return listings, pagination, nil
}

func valOrEmpty[T ~string](s *T) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func valOrZero[T any](v *T) interface{} {
	if v == nil {
		return ""
	}
	return *v
}

func buildPagination(page, limit int, totalItems int64) *models.PaginationMetadata {
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	if totalPages == 0 && totalItems == 0 {
		totalPages = 0
	}

	return &models.PaginationMetadata{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}
}
