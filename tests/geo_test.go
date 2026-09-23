package tests

import (
	"math"
	"testing"

	"expertlisting/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestDegreesAndRadiansConversion(t *testing.T) {
	deg := 180.0
	rad := utils.DegreesToRadians(deg)
	assert.InDelta(t, math.Pi, rad, 0.0001)

	convertedBack := utils.RadiansToDegrees(rad)
	assert.InDelta(t, deg, convertedBack, 0.0001)
}

func TestHaversineDistance_KnownLocations(t *testing.T) {
	// London to Paris (~343 km)
	londonLat, londonLng := 51.5074, -0.1278
	parisLat, parisLng := 48.8566, 2.3522

	dist := utils.HaversineDistance(londonLat, londonLng, parisLat, parisLng)
	assert.InDelta(t, 343.5, dist, 2.0)

	zeroDist := utils.HaversineDistance(londonLat, londonLng, londonLat, londonLng)
	assert.Equal(t, 0.0, zeroDist)

	// Lagos to Abuja (~535 km)
	lagosLat, lagosLng := 6.5244, 3.3792
	abujaLat, abujaLng := 9.0765, 7.3986
	lagosAbujaDist := utils.HaversineDistance(lagosLat, lagosLng, abujaLat, abujaLng)
	assert.InDelta(t, 535.0, lagosAbujaDist, 10.0)
}

func TestCalculateBoundingBox(t *testing.T) {
	centerLat, centerLng := 6.4474, 3.4357
	radiusKm := 10.0

	bbox := utils.CalculateBoundingBox(centerLat, centerLng, radiusKm)

	assert.True(t, bbox.MinLat < centerLat)
	assert.True(t, bbox.MaxLat > centerLat)
	assert.True(t, bbox.MinLon < centerLng)
	assert.True(t, bbox.MaxLon > centerLng)

	point5kmNorthLat := centerLat + 0.045
	assert.True(t, point5kmNorthLat <= bbox.MaxLat && point5kmNorthLat >= bbox.MinLat)
}

func TestRoundFloat(t *testing.T) {
	val := 12.3456789
	rounded := utils.RoundFloat(val, 2)
	assert.Equal(t, 12.35, rounded)

	rounded1 := utils.RoundFloat(val, 1)
	assert.Equal(t, 12.3, rounded1)
}
