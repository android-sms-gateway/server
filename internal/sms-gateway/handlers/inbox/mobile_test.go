package inbox_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/inbox"
	inboxsvc "github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// attachmentEnvelope is the encrypted attachment payload produced by the device.
const attachmentEnvelope = "$rsa-oaep-aes-256-gcm$v=1$k=1$QUFBQQ==$QUFBQQ==$QUFBQQ=="

// parseUpload parses the body the same way the mobile inbox handler does and
// returns the response status, the response body and the mapped domain inputs.
func parseUpload(t *testing.T, body string) (int, string, []inboxsvc.MessageInput) {
	t.Helper()

	handler := base.Handler{
		Logger:    zap.NewNop(),
		Validator: validator.New(),
	}

	parsed := make(chan []inboxsvc.MessageInput, 1)

	app := fiber.New(fiber.Config{
		// Mirrors go-infra-fx/http errorHandler ({"message": err.Error()}).
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
				code = fiberErr.Code
			}
			return c.Status(code).JSON(&fiber.Map{"message": err.Error()})
		},
	})

	app.Post("/mobile/v1/inbox", func(c *fiber.Ctx) error {
		req := make(smsgateway.MobilePostInboxRequest, 0)
		if err := handler.BodyParserValidator(c, &req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		parsed <- inbox.ToMessageInputs(req)

		return c.SendStatus(fiber.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/mobile/v1/inbox", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var msgs []inboxsvc.MessageInput
	select {
	case msgs = <-parsed:
	default:
	}

	return resp.StatusCode, string(raw), msgs
}

// TestMobileUpload_AttachmentEnvelopeIsAccepted pins that an upload with an
// attachment is accepted and that the encrypted payload is stored verbatim.
func TestMobileUpload_AttachmentEnvelopeIsAccepted(t *testing.T) {
	t.Parallel()

	body := `[
  {
    "id": "mms-001",
    "type": "MMS_DOWNLOADED",
    "sender": "+79990001234",
    "recipient": "+79990005678",
    "simNumber": 1,
    "content": "` + attachmentEnvelope + `",
    "isEncrypted": true,
    "createdAt": "2026-10-21T10:30:00Z",
    "attachments": [
      {
        "partId": 1,
        "contentType": "image/jpeg",
        "name": "$aes-256-cbc/pbkdf2-sha1$i=75000$c2FsdA==$Y2lwaGVy",
        "size": 102400,
        "data": "` + attachmentEnvelope + `"
      }
    ]
  }
]`

	status, raw, msgs := parseUpload(t, body)

	require.Equal(t, fiber.StatusCreated, status, "response body: %s", raw)
	require.Len(t, msgs, 1)
	require.Len(t, msgs[0].Attachments, 1)

	att := msgs[0].Attachments[0]
	require.Equal(t, int64(1), att.PartID)
	require.Equal(t, "image/jpeg", att.ContentType)
	require.Equal(t, "$aes-256-cbc/pbkdf2-sha1$i=75000$c2FsdA==$Y2lwaGVy", att.Name)
	require.NotNil(t, att.Size)
	require.Equal(t, int64(102400), *att.Size)
	require.Equal(t, []byte(attachmentEnvelope), att.Data)
}

// TestMobileUpload_AttachmentPayloadIsNotDecoded pins that the attachment
// payload is opaque: a base64-looking value is kept as is and never decoded.
func TestMobileUpload_AttachmentPayloadIsNotDecoded(t *testing.T) {
	t.Parallel()

	const payload = "SGVsbG8gV29ybGQh"

	body := `[{"id":"mms-001","type":"MMS_DOWNLOADED","sender":"+79990001234","content":"encrypted",` +
		`"isEncrypted":true,"createdAt":"2026-10-21T10:30:00Z","attachments":[{"partId":1,` +
		`"contentType":"image/jpeg","name":"part001","data":"` + payload + `"}]}]`

	status, raw, msgs := parseUpload(t, body)

	require.Equal(t, fiber.StatusCreated, status, "response body: %s", raw)
	require.Len(t, msgs, 1)
	require.Len(t, msgs[0].Attachments, 1)
	require.Equal(t, []byte(payload), msgs[0].Attachments[0].Data)
}

// TestMobileUpload_MissingAttachmentPayloadIsRejected pins that an attachment
// without a payload is rejected.
func TestMobileUpload_MissingAttachmentPayloadIsRejected(t *testing.T) {
	t.Parallel()

	body := `[{"id":"mms-001","type":"MMS_DOWNLOADED","sender":"+79990001234","content":"encrypted",` +
		`"isEncrypted":true,"createdAt":"2026-10-21T10:30:00Z","attachments":[{"partId":1,` +
		`"contentType":"image/jpeg","name":"part001"}]}]`

	status, raw, _ := parseUpload(t, body)

	require.Equal(t, fiber.StatusBadRequest, status)
	require.JSONEq(
		t,
		`{"message":"failed to validate: Key: '[0].Attachments[0].Data' Error:Field validation for 'Data' failed on the 'required' tag"}`,
		raw,
	)
}
