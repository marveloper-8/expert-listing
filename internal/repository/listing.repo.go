package repository

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"expertlisting/internal/models"
	"expertlisting/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrListingNotFound = errors.New("listing not found")
)

type ListingRepository interface {
	Create(listing *models.Listing) error
	GetByID(id uuid.UUID) (*models.Listing, error)
	Update(listing *models.Listing) error
	Delete(id uuid.UUID) error
	List(query models.ListListingsQuery) ([]models.Listing, int64, error)
	Search(query models.SearchListingQuery) ([]models.Listing, int64, error)
}

type listingRepository struct {
	db *gorm.DB
}

func NewListingRepository(db *gorm.DB) ListingRepository {
	return &listingRepository{db: db}
}

func (r *listingRepository) Create(listing *models.Listing) error {
	return r.db.Create(listing).Error
}

func (r *listingRepository) GetByID(id uuid.UUID) (*models.Listing, error) {
	var listing models.Listing
	err := r.db.First(&listing, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	return &listing, nil
}

func (r *listingRepository) Update(listing *models.Listing) error {
	return r.db.Save(listing).Error
}

func (r *listingRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&models.Listing{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrListingNotFound
	}
	return nil
}

func (r *listingRepository) List(query models.ListListingsQuery) ([]models.Listing, int64, error) {
	var listings []models.Listing
	var total int64

	db := r.db.Model(&models.Listing{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	sortBy := "created_at"
	if query.SortBy != "" {
		sortBy = query.SortBy
	}
	order := "desc"
	if strings.ToLower(query.Order) == "asc" {
		order = "asc"
	}

	orderClause := fmt.Sprintf("%s %s", sortBy, order)
	err := db.Order(orderClause).Limit(limit).Offset(offset).Find(&listings).Error
	if err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}

func (r *listingRepository) Search(query models.SearchListingQuery) ([]models.Listing, int64, error) {
	db := r.db.Model(&models.Listing{})

	if query.Type != nil && *query.Type != "" {
		db = db.Where("type = ?", *query.Type)
	}
	if query.MinPrice != nil {
		db = db.Where("price >= ?", *query.MinPrice)
	}
	if query.MaxPrice != nil {
		db = db.Where("price <= ?", *query.MaxPrice)
	}
	if query.Bedrooms != nil {
		db = db.Where("bedrooms = ?", *query.Bedrooms)
	} else if query.MinBedrooms != nil {
		db = db.Where("bedrooms >= ?", *query.MinBedrooms)
	}
	if query.City != nil && *query.City != "" {
		db = db.Where("LOWER(city) = LOWER(?)", *query.City)
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 || limit > 100 {
		limit = 10
	}

	isGeoSearch := query.Lat != nil && query.Lng != nil && query.RadiusKm != nil

	if isGeoSearch {
		centerLat := *query.Lat
		centerLng := *query.Lng
		radiusKm := *query.RadiusKm

		bbox := utils.CalculateBoundingBox(centerLat, centerLng, radiusKm)
		db = db.Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
			bbox.MinLat, bbox.MaxLat, bbox.MinLon, bbox.MaxLon)

		var candidates []models.Listing
		if err := db.Find(&candidates).Error; err != nil {
			return nil, 0, err
		}

		var matches []models.Listing
		for i := range candidates {
			dist := utils.HaversineDistance(centerLat, centerLng, candidates[i].Location.Latitude, candidates[i].Location.Longitude)
			if dist <= radiusKm {
				roundedDist := utils.RoundFloat(dist, 2)
				candidates[i].DistanceKm = &roundedDist
				matches = append(matches, candidates[i])
			}
		}

		isAsc := strings.ToLower(query.Order) == "asc"
		switch query.SortBy {
		case "distance":
			sort.Slice(matches, func(i, j int) bool {
				d1, d2 := *matches[i].DistanceKm, *matches[j].DistanceKm
				if isAsc {
					return d1 < d2
				}
				return d1 > d2
			})
		case "price":
			sort.Slice(matches, func(i, j int) bool {
				if isAsc {
					return matches[i].Price < matches[j].Price
				}
				return matches[i].Price > matches[j].Price
			})
		case "bedrooms":
			sort.Slice(matches, func(i, j int) bool {
				if isAsc {
					return matches[i].Bedrooms < matches[j].Bedrooms
				}
				return matches[i].Bedrooms > matches[j].Bedrooms
			})
		default:
			sort.Slice(matches, func(i, j int) bool {
				return *matches[i].DistanceKm < *matches[j].DistanceKm
			})
		}

		total := int64(len(matches))
		offset := (page - 1) * limit
		if offset >= len(matches) {
			return []models.Listing{}, total, nil
		}

		end := offset + limit
		if end > len(matches) {
			end = len(matches)
		}

		return matches[offset:end], total, nil
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	if query.SortBy != "" && query.SortBy != "distance" {
		sortBy = query.SortBy
	}
	order := "desc"
	if strings.ToLower(query.Order) == "asc" {
		order = "asc"
	}

	offset := (page - 1) * limit
	orderClause := fmt.Sprintf("%s %s", sortBy, order)

	var listings []models.Listing
	err := db.Order(orderClause).Limit(limit).Offset(offset).Find(&listings).Error
	if err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}
