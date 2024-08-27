package quartermaster

import "context"

func (q *client) ApplyItemEffectByID(ctx context.Context, itemID int) string {
	switch itemID {
	case 1:
		// Deploy to sow confusion among your rivals by warping their clue texts into bewildering riddles.
		return "Effect applied: Confusion sown among rivals."
	case 2:
		// Instantly reveal a hidden point on the map.
		return "Effect applied: Hidden point revealed."
	case 3:
		// Instantly capture a tier one challenge.
		return "Effect applied: Tier one challenge captured."
	case 4:
		// Instantly capture a tier two challenge.
		return "Effect applied: Tier two challenge captured."
	case 5:
		// Instantly capture a tier three challenge.
		return "Effect applied: Tier three challenge captured."
	case 6:
		// Steal all of another team's items. Must be within a 100 meter radius of the target team to use.
		return "Effect applied: Items stolen from another team."
	case 7:
		// Destroy one of another team's items at random. Can be used from any distance.
		return "Effect applied: Random item destroyed from another team."
	case 8:
		// Hold in your inventory to increase your score by 1.
		return "Effect applied: Score increased by 1."
	default:
		return "No effect found for this item."
	}
}
