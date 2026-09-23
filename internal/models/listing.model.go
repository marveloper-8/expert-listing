package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListingType string

const (
	ListingTypeRent     ListingType = "rent"
	ListingTypeSale     ListingType = "sale"
	ListingTypeShortlet ListingType = "shortlet"
)

func (lt ListingType) Valid() bool {
	switch lt {
	case ListingTypeRent, ListingTypeSale, ListingTypeShortlet:
		return true
	default:
		return false
	}
}

type Location struct {
	Address   string  `gorm:"size:255" json:"address,omitempty" example:"12 Alexander Avenue, Ikoyi"`
	City      string  `gorm:"size:100;index" json:"city,omitempty" example:"Lagos"`
	Latitude  float64 `gorm:"type:decimal(10,7);not null;index" json:"latitude" binding:"required,latitude" example:"6.4520"`
	Longitude float64 `gorm:"type:decimal(10,7);not null;index" json:"longitude" binding:"required,longitude" example:"3.4350"`
}

type Listing struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id" example:"11111111-1111-1111-1111-111111111111"`
	Title       string      `gorm:"size:255;not null;index" json:"title" example:"Luxury 3-Bedroom Penthouse in Ikoyi"`
	Description string      `gorm:"type:text" json:"description,omitempty" example:"Modern penthouse with ocean view and automated lighting."`
	Price       float64     `gorm:"type:decimal(14,2);not null;index" json:"price" example:"185000000.00"`
	Type        ListingType `gorm:"type:varchar(20);not null;index" json:"type" example:"sale"`
	Bedrooms    int         `gorm:"not null;index" json:"bedrooms" example:"3"`
	Bathrooms   int         `gorm:"default:1" json:"bathrooms" example:"3"`
	Location    Location    `gorm:"embedded" json:"location"`
	AgentID     string      `gorm:"size:100;not null;index" json:"agent_id" example:"agent_1001"`
	DistanceKm  *float64    `gorm:"-" json:"distance_km,omitempty" example:"2.45"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"created_at" example:"2026-03-23T10:00:00Z"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updated_at" example:"2026-03-23T10:00:00Z"`
}

func (l *Listing) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

type CreateListingRequest struct {
	Title       string      `json:"title" binding:"required,min=3,max=255" example:"Luxury 3-Bedroom Penthouse in Ikoyi"`
	Description string      `json:"description" binding:"max=2000" example:"Modern penthouse with ocean view."`
	Price       float64     `json:"price" binding:"required,gte=0" example:"185000000.00"`
	Type        ListingType `json:"type" binding:"required,oneof=rent sale shortlet" example:"sale"`
	Bedrooms    int         `json:"bedrooms" binding:"required,gte=0" example:"3"`
	Bathrooms   int         `json:"bathrooms" binding:"gte=0" example:"3"`
	Location    Location    `json:"location" binding:"required"`
	AgentID     string      `json:"agent_id" binding:"required,min=1,max=100" example:"agent_1001"`
}

type UpdateLocationRequest struct {
	Address   *string  `json:"address,omitempty" example:"14 Alexander Avenue, Ikoyi"`
	City      *string  `json:"city,omitempty" example:"Lagos"`
	Latitude  *float64 `json:"latitude,omitempty" binding:"omitempty,latitude" example:"6.4525"`
	Longitude *float64 `json:"longitude,omitempty" binding:"omitempty,longitude" example:"3.4355"`
}

type UpdateListingRequest struct {
	Title       *string                `json:"title" binding:"omitempty,min=3,max=255" example:"Updated Penthouse Title"`
	Description *string                `json:"description" binding:"omitempty,max=2000" example:"Updated description."`
	Price       *float64               `json:"price" binding:"omitempty,gte=0" example:"175000000.00"`
	Type        *ListingType           `json:"type" binding:"omitempty,oneof=rent sale shortlet" example:"sale"`
	Bedrooms    *int                   `json:"bedrooms" binding:"omitempty,gte=0" example:"3"`
	Bathrooms   *int                   `json:"bathrooms" binding:"omitempty,gte=0" example:"3"`
	Location    *UpdateLocationRequest `json:"location,omitempty"`
	AgentID     *string                `json:"agent_id" binding:"omitempty,min=1,max=100" example:"agent_1001"`
}

type SearchListingQuery struct {
	Type        *ListingType `form:"type" binding:"omitempty,oneof=rent sale shortlet"`
	MinPrice    *float64     `form:"min_price" binding:"omitempty,gte=0"`
	MaxPrice    *float64     `form:"max_price" binding:"omitempty,gte=0"`
	Bedrooms    *int         `form:"bedrooms" binding:"omitempty,gte=0"`
	MinBedrooms *int         `form:"min_bedrooms" binding:"omitempty,gte=0"`
	City        *string      `form:"city"`
	Lat         *float64     `form:"lat" binding:"omitempty,latitude"`
	Lng         *float64     `form:"lng" binding:"omitempty,longitude"`
	RadiusKm    *float64     `form:"radius_km" binding:"omitempty,gt=0"`
	Page        int          `form:"page,default=1" binding:"omitempty,min=1"`
	Limit       int          `form:"limit,default=10" binding:"omitempty,min=1,max=100"`
	SortBy      string       `form:"sort_by,default=created_at" binding:"omitempty,oneof=price bedrooms created_at distance"`
	Order       string       `form:"order,default=desc" binding:"omitempty,oneof=asc desc ASC DESC"`
}

type ListListingsQuery struct {
	Page   int    `form:"page,default=1" binding:"omitempty,min=1"`
	Limit  int    `form:"limit,default=10" binding:"omitempty,min=1,max=100"`
	SortBy string `form:"sort_by,default=created_at" binding:"omitempty,oneof=price bedrooms created_at"`
	Order  string `form:"order,default=desc" binding:"omitempty,oneof=asc desc ASC DESC"`
}
