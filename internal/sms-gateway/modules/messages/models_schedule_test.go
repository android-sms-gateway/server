//nolint:testpackage // exercises messageModel.toStateDomain directly.
package messages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMessageModelToStateDomainKeepsScheduleAt pins GAP5 at the server domain
// layer: the schedule_at column is read back into MessageState.ScheduleAt so
// the 3rdparty wire DTO can emit it (AC-GAP5-3). This test runs against the
// pinned client-go because it only touches server-owned types.
func TestMessageModelToStateDomainKeepsScheduleAt(t *testing.T) {
	scheduleAt := time.Date(2026, 12, 1, 8, 30, 45, 123456789, time.UTC)

	model := messageModel{
		ExtID:      "scheduled-1",
		DeviceID:   "device-1",
		State:      ProcessingStatePending,
		ScheduleAt: &scheduleAt,
	}

	state, err := model.toStateDomain()
	require.NoError(t, err)
	require.NotNil(t, state.ScheduleAt)
	require.True(t, state.ScheduleAt.Equal(scheduleAt))
}
