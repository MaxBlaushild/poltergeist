package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type orderItemResponse struct {
	models.BgiOrderItem
	ProductName string `json:"productName"`
}

type orderResponse struct {
	models.BgiOrder
	Items []orderItemResponse `json:"items"`
}

// GET /api/bgi/orders/:token (R-8.1, R-8.2: no-login order status lookup).
func (s *server) getOrder(c *gin.Context) {
	ctx := c.Request.Context()
	order, err := s.deps.DbClient.BgiOrder().FindByToken(ctx, c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	items := make([]orderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		resp := orderItemResponse{BgiOrderItem: item}
		if product, err := s.deps.DbClient.BgiProduct().FindByID(ctx, item.ProductID); err == nil {
			resp.ProductName = product.Name
		}
		items = append(items, resp)
	}

	c.JSON(http.StatusOK, orderResponse{BgiOrder: *order, Items: items})
}
