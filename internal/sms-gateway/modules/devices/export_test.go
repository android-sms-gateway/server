package devices

// ValidateE2EPair exposes the unexported E2E guard to tests.
func ValidateE2EPair(publicKey *string, keyVersion *int) error {
	return validateE2EPair(publicKey, keyVersion)
}
