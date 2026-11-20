package push

func NewEventWrapper(token, tokenType string) *eventWrapper {
	return &eventWrapper{Token: token, TokenType: tokenType}
}

func EventWrapperKey(wrapper *eventWrapper) string {
	return wrapper.key()
}
