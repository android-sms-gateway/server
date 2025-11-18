package main

import (
	"os"

	smsgateway "github.com/android-sms-gateway/server/internal/sms-gateway"
	"github.com/android-sms-gateway/server/internal/worker"
)

const (
	cmdWorker = "worker"
)

//	@securitydefinitions.basic	ApiAuth
//	@description				User authentication

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

//	@host		localhost:3000/api
//	@host		api.sms-gate.app
//	@schemes	https
//
// SMSGate Backend.
func main() {
	args := os.Args[1:]
	cmd := "start"
	if len(args) > 0 && args[0] == cmdWorker {
		cmd = cmdWorker
	}

	if cmd == cmdWorker {
		worker.Run()
	} else {
		smsgateway.Run()
	}
}
