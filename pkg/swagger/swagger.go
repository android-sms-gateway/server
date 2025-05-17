package swagger

import "embed"

//go:generate swag init --parseDependency --tags=User,System --outputTypes json,yaml -d ../../ -g ./cmd/sms-gateway/main.go -o ../../pkg/swagger/docs

//go:embed docs/*.png docs/*.js docs/*.css docs/swagger.* docs/*.html
var Docs embed.FS
