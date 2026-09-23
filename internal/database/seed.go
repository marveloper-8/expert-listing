package database

import (
	"log"

	"expertlisting/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedDemoData(db *gorm.DB) {
	var count int64
	if err := db.Model(&models.Listing{}).Count(&count).Error; err != nil {
		log.Printf("seed check error: %v", err)
		return
	}

	if count > 0 {
		return
	}

	sampleListings := []models.Listing{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Title:       "Luxury 3-Bedroom Penthouse in Ikoyi",
			Description: "Penthouse with waterfront views, private elevator, and 24/7 serviced power.",
			Price:       185000000.00,
			Type:        models.ListingTypeSale,
			Bedrooms:    3,
			Bathrooms:   3,
			Location: models.Location{
				Address:   "12 Alexander Avenue, Ikoyi",
				City:      "Lagos",
				Latitude:  6.4520,
				Longitude: 3.4350,
			},
			AgentID: "agent_alpha_101",
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Title:       "Executive 2-Bed Flat in Victoria Island",
			Description: "Serviced apartment in commercial VI. 24/7 security and dedicated parking.",
			Price:       8500000.00,
			Type:        models.ListingTypeRent,
			Bedrooms:    2,
			Bathrooms:   2,
			Location: models.Location{
				Address:   "45 Ahmadu Bello Way, Victoria Island",
				City:      "Lagos",
				Latitude:  6.4281,
				Longitude: 3.4219,
			},
			AgentID: "agent_alpha_101",
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Title:       "Waterfront Shortlet Studio",
			Description: "Cozy serviced studio with fiber internet, smart TV, and pool access.",
			Price:       75000.00,
			Type:        models.ListingTypeShortlet,
			Bedrooms:    1,
			Bathrooms:   1,
			Location: models.Location{
				Address:   "8 Ocean View Drive, Lekki Phase 1",
				City:      "Lagos",
				Latitude:  6.4470,
				Longitude: 3.4750,
			},
			AgentID: "agent_beta_202",
		},
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			Title:       "4-Bedroom Detached Duplex in Ikeja GRA",
			Description: "Family home with private compound, fitted kitchen, and BQ.",
			Price:       220000000.00,
			Type:        models.ListingTypeSale,
			Bedrooms:    4,
			Bathrooms:   4,
			Location: models.Location{
				Address:   "19 Isaac John Street, Ikeja GRA",
				City:      "Lagos",
				Latitude:  6.5925,
				Longitude: 3.3560,
			},
			AgentID: "agent_gamma_303",
		},
		{
			ID:          uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			Title:       "2-Bedroom Shortlet Suite in Lekki",
			Description: "Shortlet suite with workstation, generator backup, and daily housekeeping.",
			Price:       110000.00,
			Type:        models.ListingTypeShortlet,
			Bedrooms:    2,
			Bathrooms:   2,
			Location: models.Location{
				Address:   "22 Admiralty Way, Lekki Phase 1",
				City:      "Lagos",
				Latitude:  6.4485,
				Longitude: 3.4710,
			},
			AgentID: "agent_beta_202",
		},
		{
			ID:          uuid.MustParse("66666666-6666-6666-6666-666666666666"),
			Title:       "3-Bed Townhouse in Oniru",
			Description: "Gated estate townhouse with communal gym and 24/7 security.",
			Price:       12000000.00,
			Type:        models.ListingTypeRent,
			Bedrooms:    3,
			Bathrooms:   3,
			Location: models.Location{
				Address:   "5 Oniru Chieftaincy Estate",
				City:      "Lagos",
				Latitude:  6.4350,
				Longitude: 3.4410,
			},
			AgentID: "agent_alpha_101",
		},
	}

	for _, item := range sampleListings {
		if err := db.Create(&item).Error; err != nil {
			log.Printf("seed record error %s: %v", item.Title, err)
		}
	}

	log.Printf("Seeded %d demo listings", len(sampleListings))
}
