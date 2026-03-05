package models

import "strings"

type RewardMode string

const (
	RewardModeExplicit RewardMode = "explicit"
	RewardModeRandom   RewardMode = "random"
)

type RandomRewardSize string

const (
	RandomRewardSizeSmall  RandomRewardSize = "small"
	RandomRewardSizeMedium RandomRewardSize = "medium"
	RandomRewardSizeLarge  RandomRewardSize = "large"
)

func NormalizeRewardMode(raw string) RewardMode {
	switch RewardMode(strings.ToLower(strings.TrimSpace(raw))) {
	case RewardModeExplicit:
		return RewardModeExplicit
	default:
		return RewardModeRandom
	}
}

func IsValidRewardMode(raw string) bool {
	switch RewardMode(strings.ToLower(strings.TrimSpace(raw))) {
	case RewardModeExplicit, RewardModeRandom:
		return true
	default:
		return false
	}
}

func NormalizeRandomRewardSize(raw string) RandomRewardSize {
	switch RandomRewardSize(strings.ToLower(strings.TrimSpace(raw))) {
	case RandomRewardSizeMedium:
		return RandomRewardSizeMedium
	case RandomRewardSizeLarge:
		return RandomRewardSizeLarge
	default:
		return RandomRewardSizeSmall
	}
}

func IsValidRandomRewardSize(raw string) bool {
	switch RandomRewardSize(strings.ToLower(strings.TrimSpace(raw))) {
	case RandomRewardSizeSmall, RandomRewardSizeMedium, RandomRewardSizeLarge:
		return true
	default:
		return false
	}
}
