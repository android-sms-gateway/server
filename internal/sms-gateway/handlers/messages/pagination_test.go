//nolint:testpackage // ThirdPartyController holds a concrete *messages.Service; in-package test required.
package messages

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	messages "github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	paginationTestUserID   = "user-pagination"
	paginationTestDeviceID = "device-12345678901234"
)

const paginationTestNegativeOffsetBody = `{"message":"failed to validate: Key: 'thirdPartyGetQueryParams.PaginationOptions.Offset' ` +
	`Error:Field validation for 'Offset' failed on the 'min' tag"}`

func newPaginationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// messages.Migrate cannot run under sqlite: the auto-migrated devices table
	// uses MySQL-only DDL defaults. Create the tables the repository actually
	// queries with sqlite-compatible DDL instead.
	for _, ddl := range []string{
		`CREATE TABLE messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			device_id TEXT NOT NULL,
			ext_id TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'Text',
			content TEXT NOT NULL,
			state TEXT NOT NULL DEFAULT 'Pending',
			valid_until DATETIME,
			schedule_at DATETIME,
			sim_number INTEGER,
			with_delivery_report INTEGER NOT NULL DEFAULT 0,
			priority INTEGER NOT NULL DEFAULT 0,
			is_hashed INTEGER NOT NULL DEFAULT 0,
			is_encrypted INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME
		)`,
		`CREATE TABLE message_recipients (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id INTEGER NOT NULL,
			phone_number TEXT NOT NULL,
			state TEXT NOT NULL DEFAULT 'Pending',
			error TEXT
		)`,
		`CREATE TABLE message_states (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id INTEGER NOT NULL,
			state TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	return db
}

func seedPaginationMessage(t *testing.T, db *gorm.DB, extID string, createdAt time.Time, state string) {
	t.Helper()

	res := db.Exec(
		"INSERT INTO messages (user_id, device_id, ext_id, type, content, state, with_delivery_report, priority, is_hashed, is_encrypted, created_at, updated_at) VALUES (?, ?, ?, 'Text', ?, ?, 0, 0, 0, 0, ?, ?)",
		paginationTestUserID,
		paginationTestDeviceID,
		extID,
		`{"text":"`+extID+`"}`,
		state,
		createdAt,
		createdAt,
	)
	require.NoError(t, res.Error)
}

func newPaginationTestController(t *testing.T, db *gorm.DB) *ThirdPartyController {
	t.Helper()

	svc := messages.NewService(
		messages.Config{},
		nil,
		messages.NewRepository(db),
		nil,
		nil,
		nil,
		nil,
		zap.NewNop(),
		nil,
	)

	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    zap.NewNop(),
			Validator: validator.New(),
		},
		messagesSvc: svc,
	}
}

func newPaginationTestApp(t *testing.T, ctrl *ThirdPartyController) *fiber.App {
	t.Helper()

	app := fiber.New(fiber.Config{
		// Mirrors go-infra-fx/http errorHandler ({"message": err.Error()}).
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
			}
			return c.Status(code).JSON(&fiber.Map{"message": err.Error()})
		},
	})
	app.Get("/messages", func(c *fiber.Ctx) error {
		return ctrl.list(paginationTestUserID, c)
	})

	return app
}

func listPaginationIDs(t *testing.T, app *fiber.App, query string) ([]string, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/messages"+query, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var msgs []smsgateway.MessageState
	require.NoError(t, json.Unmarshal(body, &msgs))

	ids := make([]string, 0, len(msgs))
	for _, m := range msgs {
		ids = append(ids, m.ID)
	}

	return ids, resp.Header.Get("X-Total-Count")
}

func getPaginationResponse(t *testing.T, app *fiber.App, query string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/messages"+query, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(body)
}

func seedPaginationWindow(t *testing.T, db *gorm.DB) {
	t.Helper()

	baseTime := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	extIDs := []string{"ext-1", "ext-2", "ext-3", "ext-4", "ext-5"}
	for i, extID := range extIDs {
		seedPaginationMessage(t, db, extID, baseTime.Add(time.Duration(i)*time.Minute), "Pending")
	}
}

func TestThirdPartyControllerListOffsetShiftsPaginationWindow(t *testing.T) {
	db := newPaginationTestDB(t)
	seedPaginationWindow(t, db)

	ctrl := newPaginationTestController(t, db)
	app := newPaginationTestApp(t, ctrl)

	ids, total := listPaginationIDs(t, app, "?limit=2&offset=2")
	require.Equal(t, []string{"ext-3", "ext-2"}, ids)
	require.Equal(t, "5", total)
}

func TestThirdPartyControllerListOffsetZeroEqualsOmitted(t *testing.T) {
	db := newPaginationTestDB(t)
	seedPaginationWindow(t, db)

	ctrl := newPaginationTestController(t, db)
	app := newPaginationTestApp(t, ctrl)

	omitted, _ := listPaginationIDs(t, app, "?limit=2")
	zero, _ := listPaginationIDs(t, app, "?limit=2&offset=0")
	require.Equal(t, []string{"ext-5", "ext-4"}, omitted)
	require.Equal(t, omitted, zero)
}

func TestThirdPartyControllerListOffsetBeyondTotal(t *testing.T) {
	db := newPaginationTestDB(t)
	seedPaginationWindow(t, db)

	ctrl := newPaginationTestController(t, db)
	app := newPaginationTestApp(t, ctrl)

	ids, total := listPaginationIDs(t, app, "?limit=2&offset=10")
	require.Empty(t, ids)
	require.Equal(t, "5", total)
}

func TestThirdPartyControllerListNegativeOffsetRejected(t *testing.T) {
	db := newPaginationTestDB(t)

	ctrl := newPaginationTestController(t, db)
	app := newPaginationTestApp(t, ctrl)

	status, body := getPaginationResponse(t, app, "?limit=2&offset=-1")
	require.Equal(t, fiber.StatusBadRequest, status)
	require.JSONEq(t, paginationTestNegativeOffsetBody, body)
}

func TestThirdPartyControllerListOffsetWithFilters(t *testing.T) {
	db := newPaginationTestDB(t)
	baseTime := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	rows := []struct {
		extID string
		min   int
		state string
	}{
		{extID: "ext-1", min: 1, state: "Pending"},
		{extID: "ext-2", min: 2, state: "Processed"},
		{extID: "ext-3", min: 3, state: "Processed"},
		{extID: "ext-4", min: 4, state: "Pending"},
		{extID: "ext-5", min: 5, state: "Processed"},
		{extID: "ext-6", min: 6, state: "Processed"},
	}
	for _, row := range rows {
		seedPaginationMessage(t, db, row.extID, baseTime.Add(time.Duration(row.min)*time.Minute), row.state)
	}

	ctrl := newPaginationTestController(t, db)
	app := newPaginationTestApp(t, ctrl)

	q := url.Values{}
	q.Set("state", "Processed")
	q.Set("from", baseTime.Add(time.Minute).UTC().Format(time.RFC3339))
	q.Set("to", baseTime.Add(7*time.Minute).UTC().Format(time.RFC3339))
	q.Set("limit", "2")
	q.Set("offset", "2")

	ids, total := listPaginationIDs(t, app, "?"+q.Encode())
	require.Equal(t, []string{"ext-3", "ext-2"}, ids)
	require.Equal(t, "4", total)
}

// seedScheduledPaginationMessage inserts a message with a non-nil schedule_at
// so the 3rdparty GET path can read it back (AC-GAP5-3).
func seedScheduledPaginationMessage(t *testing.T, db *gorm.DB, extID string, scheduleAt time.Time) {
	t.Helper()

	res := db.Exec(
		"INSERT INTO messages (user_id, device_id, ext_id, type, content, state, schedule_at, with_delivery_report, priority, is_hashed, is_encrypted, created_at, updated_at) VALUES (?, ?, ?, 'Text', ?, 'Pending', ?, 0, 0, 0, 0, ?, ?)",
		paginationTestUserID,
		paginationTestDeviceID,
		extID,
		`{"text":"`+extID+`"}`,
		scheduleAt,
		time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	)
	require.NoError(t, res.Error)
}

// TestThirdPartyControllerListReturnsScheduleAt pins GAP5: GET messages for a
// message whose domain row carries schedule_at emits scheduleAt in the wire
// body (AC-GAP5-3).
func TestThirdPartyControllerListReturnsScheduleAt(t *testing.T) {
	db := newPaginationTestDB(t)
	scheduleAt := time.Date(2026, 12, 1, 8, 30, 0, 123456789, time.UTC)
	seedScheduledPaginationMessage(t, db, "scheduled-1", scheduleAt)

	ctrl := newPaginationTestController(t, db)
	app := newPaginationTestApp(t, ctrl)

	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var msgs []map[string]any
	require.NoError(t, json.Unmarshal(body, &msgs))
	require.Len(t, msgs, 1)

	schedStr, ok := msgs[0]["scheduleAt"].(string)
	require.Truef(t, ok, "scheduleAt absent in response body %s", body)

	got, err := time.Parse(time.RFC3339Nano, schedStr)
	require.NoError(t, err)
	require.True(t, got.Equal(scheduleAt), "scheduleAt = %s, want %s", schedStr, scheduleAt)
}
