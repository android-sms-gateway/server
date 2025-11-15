package handlers

import (
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/capcom6/go-helpers/anys"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type upstreamHandler struct {
	base.Handler

	config  Config
	pushSvc *push.Service
}

type upstreamHandlerParams struct {
	fx.In

	Config  Config
	PushSvc *push.Service

	Logger    *zap.Logger
	Validator *validator.Validate
}

func newUpstreamHandler(params upstreamHandlerParams) *upstreamHandler {
	return &upstreamHandler{
		Handler: base.Handler{Logger: params.Logger, Validator: params.Validator},
		config:  params.Config,
		pushSvc: params.PushSvc,
	}
}

//	@Summary		Send push notifications
//	@Description	Enqueues notifications for sending to devices
//	@Tags			Upstream
//	@Accept			json
//	@Produce		json
//	@Param			request	body	smsgateway.UpstreamPushRequest	true	"Push request"
//	@Success		202		"Notification enqueued"
//	@Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		429		{object}	smsgateway.ErrorResponse	"Too many requests"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/upstream/v1/push [post]
//
// Send push notifications.
func (h *upstreamHandler) postPush(c *fiber.Ctx) error {
	req := smsgateway.UpstreamPushRequest{}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if len(req) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Empty request")
	}

	for _, v := range req {
		if err := h.ValidateStruct(v); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		event := push.Event{
			Type: anys.ZeroDefault(v.Event, smsgateway.PushMessageEnqueued),
			Data: v.Data,
		}

		if err := h.pushSvc.Enqueue(v.Token, event); err != nil {
			h.Logger.Error("failed to push message", zap.Error(err))
		}
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// Register registers upstream handlers with the given router.
//
// If upstream is disabled in the configuration, this function does nothing.
//
// It registers a single handler for sending push notifications to
// /upstream/v1/push, with a rate limiter of 5 requests per minute.
func (h *upstreamHandler) Register(router fiber.Router) {
	if !h.config.UpstreamEnabled {
		return
	}

	router = router.Group("/upstream/v1")

	const (
		rateLimit = 5
		rateTime  = 60 * time.Second
	)
	router.Post("/push", limiter.New(limiter.Config{
		Max:               rateLimit,
		Expiration:        rateTime,
		LimiterMiddleware: limiter.SlidingWindow{},
	}), h.postPush)
}
