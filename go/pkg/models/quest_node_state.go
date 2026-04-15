package models

import "strings"

type QuestNodeFailurePolicy string

const (
	QuestNodeFailurePolicyRetry      QuestNodeFailurePolicy = "retry"
	QuestNodeFailurePolicyTransition QuestNodeFailurePolicy = "transition"
)

func NormalizeQuestNodeFailurePolicy(raw string) QuestNodeFailurePolicy {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestNodeFailurePolicyTransition):
		return QuestNodeFailurePolicyTransition
	default:
		return QuestNodeFailurePolicyRetry
	}
}

func (p QuestNodeFailurePolicy) IsValid() bool {
	switch NormalizeQuestNodeFailurePolicy(string(p)) {
	case QuestNodeFailurePolicyRetry, QuestNodeFailurePolicyTransition:
		return true
	default:
		return false
	}
}

type QuestNodeProgressStatus string

const (
	QuestNodeProgressStatusActive    QuestNodeProgressStatus = "active"
	QuestNodeProgressStatusCompleted QuestNodeProgressStatus = "completed"
	QuestNodeProgressStatusFailed    QuestNodeProgressStatus = "failed"
)

func NormalizeQuestNodeProgressStatus(raw string) QuestNodeProgressStatus {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestNodeProgressStatusCompleted):
		return QuestNodeProgressStatusCompleted
	case string(QuestNodeProgressStatusFailed):
		return QuestNodeProgressStatusFailed
	default:
		return QuestNodeProgressStatusActive
	}
}

type QuestNodeTransitionOutcome string

const (
	QuestNodeTransitionOutcomeSuccess QuestNodeTransitionOutcome = "success"
	QuestNodeTransitionOutcomeFailure QuestNodeTransitionOutcome = "failure"
)

func NormalizeQuestNodeTransitionOutcome(raw string) QuestNodeTransitionOutcome {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestNodeTransitionOutcomeFailure):
		return QuestNodeTransitionOutcomeFailure
	default:
		return QuestNodeTransitionOutcomeSuccess
	}
}
