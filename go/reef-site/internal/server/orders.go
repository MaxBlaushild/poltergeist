package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type orderItemResponse struct {
	models.ReefOrderItem
	ProductName string `json:"productName"`
}

type orderResponse struct {
	models.ReefOrder
	Items []orderItemResponse `json:"items"`
}

// GET /api/reef/orders/:token (R-8.1, R-8.2: no-login order status lookup).
// Enriches each line item with its product name — ReefOrderItem only stores
// ProductID, and the order page has no other way to show a customer what
// they actually bought.
func (s *server) getOrder(c *gin.Context) {
	ctx := c.Request.Context()
	order, err := s.deps.DbClient.ReefOrder().FindByToken(ctx, c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	items := make([]orderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		resp := orderItemResponse{ReefOrderItem: item}
		if product, err := s.deps.DbClient.ReefProduct().FindByID(ctx, item.ProductID); err == nil {
			resp.ProductName = product.Name
		}
		items = append(items, resp)
	}

	c.JSON(http.StatusOK, orderResponse{ReefOrder: *order, Items: items})
}
