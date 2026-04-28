package smpp

type Config struct {
	BindAddress    string
	TLSBindAddress string
	TLSCert        string
	TLSKey         string
	APIBaseURL     string
	WebhookBaseURL string
	SourceTON      uint8
	SourceNPI      uint8
	DestTON        uint8
	DestNPI        uint8
}

func DefaultConfig() Config {
	return Config{
		BindAddress:    "0.0.0.0:2775",
		TLSBindAddress: "0.0.0.0:2776",
		APIBaseURL:     "http://localhost:8080",
		WebhookBaseURL: "http://localhost:2777",
		SourceTON:      1,
		SourceNPI:      1,
		DestTON:        1,
		DestNPI:        1,
	}
}
