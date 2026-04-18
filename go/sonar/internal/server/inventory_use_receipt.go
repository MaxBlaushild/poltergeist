package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/gameengine"
	"github.com/gin-gonic/gin"
)

func usedInventoryItemReceiptPayload(
	item *models.InventoryItem,
	ownedItem *models.OwnedInventoryItem,
	learnedRecipes []gin.H,
	message string,
) gin.H {
	if item == nil {
		return gin.H{}
	}

	remainingQuantity := 0
	ownedItemID := ""
	if ownedItem != nil {
		remainingQuantity = ownedItem.Quantity
		if remainingQuantity < 0 {
			remainingQuantity = 0
		}
		ownedItemID = ownedItem.ID.String()
	}

	statusesRemoved := cloneInventoryReceiptStrings([]string(item.ConsumeStatusesToRemove))
	spellIDs := cloneInventoryReceiptStrings([]string(item.ConsumeSpellIDs))
	teachRecipeIDs := cloneInventoryReceiptStrings([]string(item.ConsumeTeachRecipeIDs))

	return gin.H{
		"id":                                ownedItemID,
		"inventoryItemId":                   item.ID,
		"name":                              item.Name,
		"imageUrl":                          item.ImageURL,
		"flavorText":                        item.FlavorText,
		"effectText":                        item.EffectText,
		"message":                           strings.TrimSpace(message),
		"usedAt":                            time.Now().UTC().Format(time.RFC3339),
		"consumedQuantity":                  1,
		"remainingQuantity":                 remainingQuantity,
		"depleted":                          remainingQuantity <= 0,
		"healthDelta":                       item.ConsumeHealthDelta,
		"manaDelta":                         item.ConsumeManaDelta,
		"revivePartyMemberHealth":           item.ConsumeRevivePartyMemberHealth,
		"reviveAllDownedPartyMembersHealth": item.ConsumeReviveAllDownedPartyMembersHealth,
		"dealDamage":                        item.ConsumeDealDamage,
		"dealDamageHits":                    item.ConsumeDealDamageHits,
		"dealDamageAllEnemies":              item.ConsumeDealDamageAllEnemies,
		"dealDamageAllEnemiesHits":          item.ConsumeDealDamageAllEnemiesHits,
		"createBase":                        item.ConsumeCreateBase,
		"statusesAdded":                     item.ConsumeStatusesToAdd,
		"statusesRemoved":                   statusesRemoved,
		"spellIds":                          spellIDs,
		"teachRecipeIds":                    teachRecipeIDs,
		"learnedRecipes":                    learnedRecipes,
		"isCaptureType":                     item.IsCaptureType,
		"effectSummary":                     usedInventoryItemEffectSummary(item, learnedRecipes),
	}
}

func usedInventoryItemEffectSummary(
	item *models.InventoryItem,
	learnedRecipes []gin.H,
) []string {
	if item == nil {
		return nil
	}

	summary := make([]string, 0, 10)
	if item.ConsumeHealthDelta != 0 {
		summary = append(summary, fmt.Sprintf("HP %s%d", signedPrefix(item.ConsumeHealthDelta), item.ConsumeHealthDelta))
	}
	if item.ConsumeManaDelta != 0 {
		summary = append(summary, fmt.Sprintf("MP %s%d", signedPrefix(item.ConsumeManaDelta), item.ConsumeManaDelta))
	}
	if item.ConsumeRevivePartyMemberHealth > 0 {
		summary = append(summary, fmt.Sprintf("Revives one ally with %d HP", item.ConsumeRevivePartyMemberHealth))
	}
	if item.ConsumeReviveAllDownedPartyMembersHealth > 0 {
		summary = append(summary, fmt.Sprintf("Revives all downed allies with %d HP", item.ConsumeReviveAllDownedPartyMembersHealth))
	}
	if item.ConsumeDealDamage > 0 {
		if item.ConsumeDealDamageHits > 1 {
			summary = append(summary, fmt.Sprintf("Deals %d damage %d times", item.ConsumeDealDamage, item.ConsumeDealDamageHits))
		} else {
			summary = append(summary, fmt.Sprintf("Deals %d damage", item.ConsumeDealDamage))
		}
	}
	if item.ConsumeDealDamageAllEnemies > 0 {
		if item.ConsumeDealDamageAllEnemiesHits > 1 {
			summary = append(summary, fmt.Sprintf("Deals %d damage to all enemies %d times", item.ConsumeDealDamageAllEnemies, item.ConsumeDealDamageAllEnemiesHits))
		} else {
			summary = append(summary, fmt.Sprintf("Deals %d damage to all enemies", item.ConsumeDealDamageAllEnemies))
		}
	}
	if item.ConsumeCreateBase {
		summary = append(summary, "Creates a home base")
	}
	if len(item.ConsumeStatusesToAdd) > 0 {
		names := make([]string, 0, len(item.ConsumeStatusesToAdd))
		for _, status := range item.ConsumeStatusesToAdd {
			name := strings.TrimSpace(status.Name)
			if name == "" {
				continue
			}
			names = append(names, name)
		}
		if len(names) > 0 {
			summary = append(summary, fmt.Sprintf("Adds %s", strings.Join(names, ", ")))
		}
	}
	if len(item.ConsumeStatusesToRemove) > 0 {
		summary = append(summary, fmt.Sprintf("Removes %s", strings.Join(cloneInventoryReceiptStrings([]string(item.ConsumeStatusesToRemove)), ", ")))
	}
	if len(item.ConsumeSpellIDs) > 0 {
		summary = append(summary, fmt.Sprintf("Grants %d spell%s", len(item.ConsumeSpellIDs), pluralSuffix(len(item.ConsumeSpellIDs))))
	}

	learnedNames := make([]string, 0, len(learnedRecipes))
	for _, recipe := range learnedRecipes {
		name := strings.TrimSpace(fmt.Sprint(recipe["itemName"]))
		if name == "" {
			continue
		}
		learnedNames = append(learnedNames, name)
	}
	if len(learnedNames) > 0 {
		summary = append(summary, fmt.Sprintf("Learned %s", strings.Join(learnedNames, ", ")))
	} else if len(item.ConsumeTeachRecipeIDs) > 0 {
		summary = append(summary, fmt.Sprintf("Can teach %d recipe%s", len(item.ConsumeTeachRecipeIDs), pluralSuffix(len(item.ConsumeTeachRecipeIDs))))
	}

	if len(summary) == 0 {
		if effectText := strings.TrimSpace(item.EffectText); effectText != "" {
			summary = append(summary, effectText)
		}
	}

	return summary
}

func submissionResultResponsePayload(result *gameengine.SubmissionResult) gin.H {
	if result == nil {
		return gin.H{}
	}

	payload := gin.H{}
	encoded, err := json.Marshal(result)
	if err != nil {
		return payload
	}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return gin.H{}
	}
	return payload
}

func cloneInventoryReceiptStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	cloned := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		cloned = append(cloned, trimmed)
	}
	return cloned
}

func signedPrefix(value int) string {
	if value > 0 {
		return "+"
	}
	return ""
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
