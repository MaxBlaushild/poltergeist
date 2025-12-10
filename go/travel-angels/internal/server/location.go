package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *server) SearchLocation(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "query parameter 'q' is required",
		})
		return
	}

	candidates, err := s.googleMapsClient.FindCandidatesByQuery(query)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("Error searching locations for query '%s': %v\n", query, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Transform candidates to response format
	results := make([]map[string]interface{}, 0, len(candidates))
	for _, candidate := range candidates {
		result := map[string]interface{}{
			"placeId":          candidate.PlaceID,
			"name":             candidate.Name,
			"formattedAddress": candidate.FormattedAddress,
			"latitude":         candidate.Geometry.Location.Lat,
			"longitude":        candidate.Geometry.Location.Lng,
		}
		results = append(results, result)
	}

	fmt.Printf("Location search for '%s' returned %d results\n", query, len(results))
	ctx.JSON(http.StatusOK, results)
}
