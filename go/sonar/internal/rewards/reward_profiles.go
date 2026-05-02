package rewards

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func ApplyDefaultRewardProfiles(
	ctx context.Context,
	dbClient db.DbClient,
	rewardContext *models.RandomRewardContext,
) (*models.RandomRewardContext, error) {
	if rewardContext == nil || dbClient == nil {
		return rewardContext, nil
	}

	profileSlugs := models.DefaultRewardProfileSlugsForContext(rewardContext)
	if len(profileSlugs) == 0 {
		return rewardContext, nil
	}

	allProfiles, err := dbClient.RewardProfile().FindAll(ctx, false)
	if err != nil {
		return nil, err
	}
	if len(allProfiles) == 0 {
		return rewardContext, nil
	}

	profilesBySlug := make(map[string]models.RewardProfile, len(allProfiles))
	for _, rewardProfile := range allProfiles {
		slug := models.NormalizeRewardProfileSlug(rewardProfile.Slug)
		if slug == "" {
			continue
		}
		profilesBySlug[slug] = rewardProfile
	}

	selected := make([]models.RewardProfile, 0, len(profileSlugs))
	for _, rawSlug := range profileSlugs {
		slug := models.NormalizeRewardProfileSlug(rawSlug)
		if slug == "" {
			continue
		}
		rewardProfile, exists := profilesBySlug[slug]
		if !exists {
			continue
		}
		selected = append(selected, rewardProfile)
	}
	if len(selected) == 0 {
		return rewardContext, nil
	}

	rewardContext.ApplyRewardProfiles(selected)
	return rewardContext, nil
}
