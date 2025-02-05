package logs

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type thirdPartyControllerParams struct {
	fx.In

	Validator *validator.Validate
	Logger    *zap.Logger
}

type ThirdPartyController struct {
	base.Handler
}

//	@Summary		Get logs
//	@Description	Retrieve a list of log entries within a specified time range.
//	@Security		ApiAuth
//	@Tags			System, Logs
//	@Produce		json
//	@Param			from	query		string						false	"The start of the time range for the logs to retrieve. Logs created after this timestamp will be included."	Format(date-time)
//	@Param			to		query		string						false	"The end of the time range for the logs to retrieve. Logs created before this timestamp will be included."	Format(date-time)
//	@Success		200		{object}	smsgateway.GetLogsResponse	"Log entries"
//	@Failure		401		{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Failure		501		{object}	smsgateway.ErrorResponse	"Not implemented"
//	@Router			/3rdparty/v1/logs [get]
//
// List webhooks
func (h *ThirdPartyController) get(user models.User, c *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusNotImplemented, "For privacy reasons, device's logs are not accessible through Cloud server")
}

func (h *ThirdPartyController) Register(router fiber.Router) {
	router.Get("", userauth.WithUser(h.get))
}

func NewThirdPartyController(params thirdPartyControllerParams) *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    params.Logger.Named("logs"),
			Validator: params.Validator,
		},
	}
}
