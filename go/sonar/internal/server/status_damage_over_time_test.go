package server

import (
	"testing"
	"time"
)

func TestBattleStatusTickReady(t *testing.T) {
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	t.Run("newly applied status does not tick immediately", func(t *testing.T) {
		if battleStatusTickReady(now, nil, now) {
			t.Fatal("expected newly applied status to wait until the next turn")
		}
	})

	t.Run("status with last tick set to now does not tick again immediately", func(t *testing.T) {
		lastTickAt := now
		if battleStatusTickReady(now.Add(-time.Minute), &lastTickAt, now) {
			t.Fatal("expected status with current-turn last tick to wait until the next turn")
		}
	})

	t.Run("older status can tick", func(t *testing.T) {
		if !battleStatusTickReady(now.Add(-time.Minute), nil, now) {
			t.Fatal("expected older status to tick on a later turn")
		}
	})
}
