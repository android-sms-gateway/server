package devices_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) *devices.Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.Exec(`CREATE TABLE devices (
		created_at datetime,
		updated_at datetime,
		deleted_at datetime,
		id text PRIMARY KEY,
		name text,
		auth_token text NOT NULL,
		push_token text,
		last_seen datetime NOT NULL,
		user_id text NOT NULL,
		sim_cards text,
		public_key text,
		key_version integer,
		state text NOT NULL DEFAULT 'Active'
	)`).Error)

	var seq int
	idGen := func() string {
		seq++
		return fmt.Sprintf("device-%02d", seq)
	}

	return devices.NewService(devices.Config{}, devices.NewRepository(db), idGen, zap.NewNop())
}

func newTestDevices(t *testing.T, svc *devices.Service, userID string) (devices.Device, devices.Device) {
	t.Helper()

	ctx := context.Background()

	enabledDevice, err := svc.Insert(ctx, userID, devices.DeviceInfo{})
	require.NoError(t, err)

	disabledDevice, err := svc.Insert(ctx, userID, devices.DeviceInfo{})
	require.NoError(t, err)

	require.NoError(t, svc.Update(ctx, disabledDevice.ID, devices.DeviceUpdate{
		State: lo.ToPtr(devices.DeviceStateDisabled),
	}))

	return *enabledDevice, *disabledDevice
}

func TestGetAnySkipsDisabledDevices(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const userID = "user-1"

	svc := newTestService(t)
	enabled, disabled := newTestDevices(t, svc, userID)

	t.Run("random selection never picks a disabled device", func(t *testing.T) {
		t.Parallel()

		for range 10 {
			got, getErr := svc.GetAny(ctx, userID, "", 0)
			require.NoError(t, getErr)
			require.Equal(t, enabled.ID, got.ID)
		}
	})

	t.Run("explicit selection of a disabled device fails", func(t *testing.T) {
		t.Parallel()

		_, getErr := svc.GetAny(ctx, userID, disabled.ID, 0)
		require.ErrorIs(t, getErr, devices.ErrNotFound)
	})

	t.Run("explicit selection of an enabled device succeeds", func(t *testing.T) {
		t.Parallel()

		got, getErr := svc.GetAny(ctx, userID, enabled.ID, 0)
		require.NoError(t, getErr)
		require.Equal(t, enabled.ID, got.ID)
	})

	t.Run("device list still includes the disabled device", func(t *testing.T) {
		t.Parallel()

		all, selectErr := svc.Select(ctx, userID)
		require.NoError(t, selectErr)
		require.Len(t, all, 2)
	})
}

func TestReEnablingRestoresExplicitSelection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const userID = "user-2"

	svc := newTestService(t)
	_, disabled := newTestDevices(t, svc, userID)

	require.NoError(t, svc.Update(ctx, disabled.ID, devices.DeviceUpdate{
		State: lo.ToPtr(devices.DeviceStateActive),
	}))

	got, err := svc.GetAny(ctx, userID, disabled.ID, 0)
	require.NoError(t, err)
	require.Equal(t, disabled.ID, got.ID)
}

func TestUpdateWithNoStateKeepsState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const userID = "user-3"

	svc := newTestService(t)
	device, err := svc.Insert(ctx, userID, devices.DeviceInfo{})
	require.NoError(t, err)

	require.NoError(t, svc.Update(ctx, device.ID, devices.DeviceUpdate{
		State: lo.ToPtr(devices.DeviceStateDisabled),
	}))
	require.NoError(t, svc.Update(ctx, device.ID, devices.DeviceUpdate{}))

	got, err := svc.Get(ctx, userID, devices.WithID(device.ID))
	require.NoError(t, err)
	require.NotNil(t, got.State)
	require.Equal(t, devices.DeviceStateDisabled, *got.State)
}

func TestInsertedDeviceIsActiveByDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const userID = "user-4"

	svc := newTestService(t)
	device, err := svc.Insert(ctx, userID, devices.DeviceInfo{})
	require.NoError(t, err)

	got, err := svc.GetAny(ctx, userID, device.ID, 0)
	require.NoError(t, err)
	require.NotNil(t, got.State)
	require.Equal(t, devices.DeviceStateActive, *got.State)
}
