package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"expertlisting/internal/models"
)

func BenchmarkGetListings(b *testing.B) {
	router := setupTestRouter(&testing.T{})
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/listings", nil))

	req, _ := http.NewRequest(http.MethodGet, "/api/listings?page=1&limit=10", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGeospatialSearch(b *testing.B) {
	router := setupTestRouter(&testing.T{})

	req, _ := http.NewRequest(http.MethodGet, "/api/listings/search?lat=6.4281&lng=3.4219&radius_km=10", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHaversineMath(b *testing.B) {
	lat1, lon1 := 6.4281, 3.4219
	lat2, lon2 := 6.4520, 3.4350

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = models.Location{Latitude: lat1, Longitude: lon1}
		_ = models.Location{Latitude: lat2, Longitude: lon2}
	}
}
