package e2e

import (
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	PublicURL  = "http://localhost:3000/api"
	PrivateURL = "http://localhost:3001/api"
)

var (
	publicMobileClient = resty.New().
				SetBaseURL(PublicURL + "/mobile/v1").
				SetTimeout(300 * time.Millisecond)
	privateMobileClient = resty.New().
				SetBaseURL(PrivateURL + "/mobile/v1").
				SetTimeout(300 * time.Millisecond)

	publicUserClient = resty.New().
				SetBaseURL(PublicURL + "/3rdparty/v1").
				SetTimeout(300 * time.Millisecond)
	privateUserClient = resty.New().
				SetBaseURL(PrivateURL + "/3rdparty/v1").
				SetTimeout(300 * time.Millisecond)
)
