package models

import "strings"

func NormalizeTagList(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}

	return normalized
}

func MergeTagLists(lists ...[]string) []string {
	total := 0
	for _, list := range lists {
		total += len(list)
	}
	if total == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, total)
	merged := make([]string, 0, total)
	for _, list := range lists {
		for _, raw := range list {
			tag := strings.ToLower(strings.TrimSpace(raw))
			if tag == "" {
				continue
			}
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			merged = append(merged, tag)
		}
	}

	return merged
}
