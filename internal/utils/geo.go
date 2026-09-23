package utils

import (
	"math"
)

const EarthRadiusKm = 6371.0

func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func RadiansToDegrees(radians float64) float64 {
	return radians * 180.0 / math.Pi
}

func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := DegreesToRadians(lat2 - lat1)
	dLon := DegreesToRadians(lon2 - lon1)

	lat1Rad := DegreesToRadians(lat1)
	lat2Rad := DegreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	if a > 1.0 {
		a = 1.0
	}

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadiusKm * c
}

type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

func CalculateBoundingBox(lat, lon, radiusKm float64) BoundingBox {
	latRad := DegreesToRadians(lat)
	radDist := radiusKm / EarthRadiusKm

	minLat := lat - RadiansToDegrees(radDist)
	maxLat := lat + RadiansToDegrees(radDist)

	if minLat > -90.0 && maxLat < 90.0 {
		deltaLon := math.Asin(math.Sin(radDist) / math.Cos(latRad))
		deltaLonDeg := RadiansToDegrees(deltaLon)

		return BoundingBox{
			MinLat: minLat,
			MaxLat: maxLat,
			MinLon: lon - deltaLonDeg,
			MaxLon: lon + deltaLonDeg,
		}
	}

	return BoundingBox{
		MinLat: math.Max(-90.0, minLat),
		MaxLat: math.Min(90.0, maxLat),
		MinLon: -180.0,
		MaxLon: 180.0,
	}
}

func RoundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
