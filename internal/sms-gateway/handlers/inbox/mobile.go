package inbox

import (
	"errors"
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type MobileController struct {
	base.Handler

	inboxSvc *inbox.Service
}

func NewMobileController(
	inboxSvc *inbox.Service,
	logger *zap.Logger,
	validator *validator.Validate,
) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},
		inboxSvc: inboxSvc,
	}
}

func (h *MobileController) Register(router fiber.Router) {
	router.Use(errorHandler)
	router.Post("", deviceauth.WithDevice(h.post))
}

//	@Summary		Upload inbox messages
//	@Description	Stores a batch of encrypted inbox messages from the device
//	@Security		MobileToken
//	@Tags			Device, Inbox
//	@Accept			json
//	@Param			request	body	smsgateway.MobilePostInboxRequest	true	"Batch of inbox messages"
//	@Success		201		"Created"
//	@Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/mobile/v1/inbox [post]
//
// Upload inbox messages.
func (h *MobileController) post(device devices.Device, c *fiber.Ctx) error {
	req := make(smsgateway.MobilePostInboxRequest, 0)
	if err := h.BodyParserValidator(c, &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	msgs := make([]inbox.MessageInput, 0, len(req))
	for _, r := range req {
		atts := make([]inbox.AttachmentInput, 0, len(r.Attachments))
		for _, a := range r.Attachments {
			atts = append(atts, inbox.AttachmentInput{
				PartID:      a.PartID,
				ContentType: a.ContentType,
				Name:        a.Name,
				Size:        a.Size,
				Data:        a.Data,
				IsEncrypted: true,
			})
		}

		msgs = append(msgs, inbox.MessageInput{
			ID:          r.ID,
			Type:        inbox.MessageType(r.Type),
			Sender:      r.Sender,
			Recipient:   r.Recipient,
			SimNumber:   r.SimNumber,
			Content:     r.Content,
			IsEncrypted: r.IsEncrypted,
			CreatedAt:   r.CreatedAt,
			Attachments: atts,
		})
	}

	if err := h.inboxSvc.InsertBatch(c.Context(), device.ID, msgs); err != nil {
		if errors.Is(err, inbox.ErrNotEncrypted) ||
			errors.Is(err, inbox.ErrEmptyBatch) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fmt.Errorf("failed to insert inbox messages: %w", err)
	}

	return c.SendStatus(fiber.StatusCreated)
}
