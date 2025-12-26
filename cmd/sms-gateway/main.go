package main

import (
	"os"

	"github.com/android-sms-gateway/server/internal/health"
	smsgateway "github.com/android-sms-gateway/server/internal/sms-gateway"
	"github.com/android-sms-gateway/server/internal/worker"
)

const (
	cmdWorker = "worker"
	cmdHealth = "health"
)

//	@securitydefinitions.basic	ApiAuth
//	@description				User authentication

//	@securitydefinitions.apikey	JWTAuth
//	@in							header
//	@name						Authorization
//	@description				JWT authentication

//	@securitydefinitions.apikey	UserCode
//	@in							header
//	@name						Authorization
//	@description				User one-time code authentication

//	@securitydefinitions.apikey	MobileToken
//	@in							header
//	@name						Authorization
//	@description				Mobile device token

//	@securitydefinitions.apikey	ServerKey
//	@in							header
//	@name						Authorization
//	@description				Private server authentication

//	@title			SMSGate API
//	@version		{APP_VERSION}
//	@description	This API provides programmatic access to sending SMS messages on Android devices. Features include sending SMS, checking message status, device management, webhook configuration, and system health checks.

//	@contact.name	SMSGate Support
//	@contact.email	support@sms-gate.app
//	@contact.url	https://docs.sms-gate.app/

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0

//	@host		localhost:3000/api
//	@host		api.sms-gate.app
//	@schemes	https
//
// SMSGate Backend.
func main() {
	args := os.Args[1:]
	cmd := "start"
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case cmdHealth:
		health.Run()
		return
	case cmdWorker:
		worker.Run()
		return
	}

	smsgateway.Run()
}
