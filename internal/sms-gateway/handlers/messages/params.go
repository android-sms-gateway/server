package messages

type postQueryParams struct {
	SkipPhoneValidation bool `query:"skipPhoneValidation"`
	DeviceActiveWithin  uint `query:"deviceActiveWithin"`
}
