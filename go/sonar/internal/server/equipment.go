package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type equipmentSlotResponse struct {
	Slot                string               `json:"slot"`
	OwnedInventoryItemID *uuid.UUID          `json:"ownedInventoryItemId,omitempty"`
	InventoryItemID     *int                 `json:"inventoryItemId,omitempty"`
	InventoryItem       *models.InventoryItem `json:"inventoryItem,omitempty"`
}

func (s *server) getUserEquipment(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	equipment, err := s.buildEquipmentResponse(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, equipment)
}

func (s *server) equipInventoryItem(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		OwnedInventoryItemID string `json:"ownedInventoryItemId" binding:"required"`
		Slot                 string `json:"slot"`
	}
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownedID, err := uuid.Parse(strings.TrimSpace(requestBody.OwnedInventoryItemID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid owned inventory item ID"})
		return
	}

	owned, err := s.dbClient.InventoryItem().FindByID(ctx, ownedID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if owned == nil || owned.UserID == nil || *owned.UserID != user.ID {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "owned inventory item not found"})
		return
	}
	if owned.Quantity <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "owned inventory item has no remaining quantity"})
		return
	}

	item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, owned.InventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	equipSlot := ""
	if item.EquipSlot != nil {
		equipSlot = strings.TrimSpace(*item.EquipSlot)
	}
	if equipSlot == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "item is not equippable"})
		return
	}
	if !models.IsValidInventoryEquipSlot(equipSlot) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid equip slot on item"})
		return
	}

	requestedSlot := strings.TrimSpace(requestBody.Slot)
	if requestedSlot == "" {
		if equipSlot == string(models.EquipmentSlotRing) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ring items require a slot"})
			return
		}
		requestedSlot = equipSlot
	}
	if !models.IsValidEquipmentSlot(requestedSlot) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid equipment slot"})
		return
	}
	if equipSlot == string(models.EquipmentSlotRing) {
		if !models.IsRingSlot(requestedSlot) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ring items must be equipped to a ring slot"})
			return
		}
	} else if equipSlot != requestedSlot {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "item cannot be equipped to that slot"})
		return
	}

	if _, err := s.dbClient.UserEquipment().Equip(ctx, user.ID, requestedSlot, owned.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	equipment, err := s.buildEquipmentResponse(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, equipment)
}

func (s *server) unequipInventoryItem(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		OwnedInventoryItemID string `json:"ownedInventoryItemId"`
		Slot                 string `json:"slot"`
	}
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slot := strings.TrimSpace(requestBody.Slot)
	ownedIDRaw := strings.TrimSpace(requestBody.OwnedInventoryItemID)
	if slot == "" && ownedIDRaw == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "slot or owned inventory item ID is required"})
		return
	}

	if slot != "" {
		if !models.IsValidEquipmentSlot(slot) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid equipment slot"})
			return
		}
		if err := s.dbClient.UserEquipment().UnequipSlot(ctx, user.ID, slot); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		ownedID, err := uuid.Parse(ownedIDRaw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid owned inventory item ID"})
			return
		}
		if err := s.dbClient.UserEquipment().UnequipOwnedItem(ctx, user.ID, ownedID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	equipment, err := s.buildEquipmentResponse(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, equipment)
}

func (s *server) buildEquipmentResponse(ctx *gin.Context, userID uuid.UUID) ([]equipmentSlotResponse, error) {
	equipment, err := s.dbClient.UserEquipment().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]equipmentSlotResponse, 0, len(equipment))
	for _, slot := range equipment {
		owned, err := s.dbClient.InventoryItem().FindByID(ctx, slot.OwnedInventoryItemID)
		if err != nil || owned == nil {
			_ = s.dbClient.UserEquipment().UnequipOwnedItem(ctx, userID, slot.OwnedInventoryItemID)
			continue
		}
		if owned.UserID == nil || *owned.UserID != userID || owned.Quantity <= 0 {
			_ = s.dbClient.UserEquipment().UnequipOwnedItem(ctx, userID, slot.OwnedInventoryItemID)
			continue
		}
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, owned.InventoryItemID)
		if err != nil || item == nil {
			_ = s.dbClient.UserEquipment().UnequipOwnedItem(ctx, userID, slot.OwnedInventoryItemID)
			continue
		}
		ownedID := slot.OwnedInventoryItemID
		itemID := owned.InventoryItemID
		response = append(response, equipmentSlotResponse{
			Slot:                slot.Slot,
			OwnedInventoryItemID: &ownedID,
			InventoryItemID:     &itemID,
			InventoryItem:       item,
		})
	}

	return response, nil
}
