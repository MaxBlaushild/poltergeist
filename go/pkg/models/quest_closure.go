package models

import (
	"strings"
	"time"
)

type QuestClosurePolicy string

const (
	QuestClosurePolicyAuto     QuestClosurePolicy = "auto"
	QuestClosurePolicyRemote   QuestClosurePolicy = "remote"
	QuestClosurePolicyInPerson QuestClosurePolicy = "in_person"
)

type QuestDebriefPolicy string

const (
	QuestDebriefPolicyNone                QuestDebriefPolicy = "none"
	QuestDebriefPolicyOptional            QuestDebriefPolicy = "optional"
	QuestDebriefPolicyRequiredForFollowup QuestDebriefPolicy = "required_for_followup"
)

type QuestClosureMethod string

const (
	QuestClosureMethodAuto     QuestClosureMethod = "auto"
	QuestClosureMethodRemote   QuestClosureMethod = "remote"
	QuestClosureMethodInPerson QuestClosureMethod = "in_person"
)

func DefaultQuestClosurePolicy(category string) QuestClosurePolicy {
	if IsMainStoryQuestCategory(category) {
		return QuestClosurePolicyInPerson
	}
	return QuestClosurePolicyRemote
}

func NormalizeQuestClosurePolicy(raw string, category string) QuestClosurePolicy {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestClosurePolicyAuto):
		return QuestClosurePolicyAuto
	case string(QuestClosurePolicyRemote):
		return QuestClosurePolicyRemote
	case string(QuestClosurePolicyInPerson):
		return QuestClosurePolicyInPerson
	default:
		return DefaultQuestClosurePolicy(category)
	}
}

func DefaultQuestDebriefPolicy(category string) QuestDebriefPolicy {
	if IsMainStoryQuestCategory(category) {
		return QuestDebriefPolicyRequiredForFollowup
	}
	return QuestDebriefPolicyOptional
}

func NormalizeQuestDebriefPolicy(raw string, category string) QuestDebriefPolicy {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestDebriefPolicyNone):
		return QuestDebriefPolicyNone
	case string(QuestDebriefPolicyRequiredForFollowup):
		return QuestDebriefPolicyRequiredForFollowup
	case string(QuestDebriefPolicyOptional):
		return QuestDebriefPolicyOptional
	default:
		return DefaultQuestDebriefPolicy(category)
	}
}

func NormalizeQuestClosureMethod(raw string) QuestClosureMethod {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestClosureMethodAuto):
		return QuestClosureMethodAuto
	case string(QuestClosureMethodRemote):
		return QuestClosureMethodRemote
	default:
		return QuestClosureMethodInPerson
	}
}

func (q *Quest) ClosurePolicyNormalized() QuestClosurePolicy {
	if q == nil {
		return DefaultQuestClosurePolicy(QuestCategorySide)
	}
	return NormalizeQuestClosurePolicy(string(q.ClosurePolicy), q.Category)
}

func (q *Quest) DebriefPolicyNormalized() QuestDebriefPolicy {
	if q == nil {
		return DefaultQuestDebriefPolicy(QuestCategorySide)
	}
	return NormalizeQuestDebriefPolicy(string(q.DebriefPolicy), q.Category)
}

func (q *QuestArchetype) ClosurePolicyNormalized() QuestClosurePolicy {
	if q == nil {
		return DefaultQuestClosurePolicy(QuestCategorySide)
	}
	return NormalizeQuestClosurePolicy(string(q.ClosurePolicy), q.Category)
}

func (q *QuestArchetype) DebriefPolicyNormalized() QuestDebriefPolicy {
	if q == nil {
		return DefaultQuestDebriefPolicy(QuestCategorySide)
	}
	return NormalizeQuestDebriefPolicy(string(q.DebriefPolicy), q.Category)
}

func (q *QuestAcceptanceV2) EffectiveClosedAt() *time.Time {
	if q == nil {
		return nil
	}
	if q.ClosedAt != nil {
		return q.ClosedAt
	}
	return q.TurnedInAt
}

func (q *QuestAcceptanceV2) EffectiveDebriefedAt() *time.Time {
	if q == nil {
		return nil
	}
	if q.DebriefedAt != nil {
		return q.DebriefedAt
	}
	if q.TurnedInAt != nil {
		return q.TurnedInAt
	}
	return nil
}

func (q *QuestAcceptanceV2) IsTurnedIn() bool {
	return q.EffectiveDebriefedAt() != nil
}

func (q *QuestAcceptanceV2) IsClosed() bool {
	return q.EffectiveClosedAt() != nil
}

func (q *QuestAcceptanceV2) IsDebriefPending() bool {
	if q == nil || !q.DebriefPending {
		return false
	}
	return q.EffectiveClosedAt() != nil && q.EffectiveDebriefedAt() == nil
}
