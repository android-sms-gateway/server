package devices_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	deviceshandler "github.com/android-sms-gateway/server/internal/sms-gateway/handlers/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
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
		push_token_type text,
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

func newTestApp(t *testing.T, svc *devices.Service) *fiber.App {
	t.Helper()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		permissions.SetScopes(c, []string{permissions.ScopeAll})
		userauth.SetUserID(c, "user-1")

		return c.Next()
	})

	controller := deviceshandler.NewThirdPartyController(svc, zap.NewNop(), validator.New())
	controller.Register(app.Group("/3rdparty/v1/devices"))

	return app
}

func patchDevice(t *testing.T, app *fiber.App, id, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPatch, "/3rdparty/v1/devices/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	require.NoError(t, err)

	return res
}

func TestPatchDeviceStateRejectsMissingValue(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	device, err := svc.Insert(context.Background(), "user-1", devices.DeviceInfo{})
	require.NoError(t, err)

	app := newTestApp(t, svc)

	res := patchDevice(t, app, device.ID, `{}`)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestPatchDeviceStateRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	device, err := svc.Insert(context.Background(), "user-1", devices.DeviceInfo{})
	require.NoError(t, err)

	app := newTestApp(t, svc)

	res := patchDevice(t, app, device.ID, `{"state":"Unknown"}`)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestPatchDeviceStateRejectsForeignDevice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	svc := newTestService(t)
	device, err := svc.Insert(ctx, "user-2", devices.DeviceInfo{})
	require.NoError(t, err)

	app := newTestApp(t, svc)

	res := patchDevice(t, app, device.ID, `{"state":"Disabled"}`)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestPatchDeviceStateUpdatesState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	svc := newTestService(t)
	device, err := svc.Insert(ctx, "user-1", devices.DeviceInfo{})
	require.NoError(t, err)

	app := newTestApp(t, svc)

	res := patchDevice(t, app, device.ID, `{"state":"Disabled"}`)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	got, err := svc.Get(ctx, "user-1", devices.WithID(device.ID))
	require.NoError(t, err)
	require.NotNil(t, got.State)
	require.Equal(t, devices.DeviceStateDisabled, *got.State)

	res = patchDevice(t, app, device.ID, `{"state":"Active"}`)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	got, err = svc.Get(ctx, "user-1", devices.WithID(device.ID))
	require.NoError(t, err)
	require.NotNil(t, got.State)
	require.Equal(t, devices.DeviceStateActive, *got.State)
}
