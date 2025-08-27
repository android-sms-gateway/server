package handlers

type Config struct {
	// PublicHost is host[:port] without scheme. Empty → use request Host.
	PublicHost string
	// PublicPath is API base path; normalized to start with "/" and have no trailing "/".
	PublicPath string

	UpstreamEnabled bool
	OpenAPIEnabled  bool
}
