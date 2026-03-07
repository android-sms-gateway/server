package version

import "strconv"

const notSet string = "not set"

// This information will be collected when build, by `-ldflags "-X main.appVersion=0.1"`.
//
//nolint:gochecknoglobals // build-time constant
var (
	AppVersion = notSet
	AppRelease = notSet
)

func AppReleaseID() int {
	id, _ := strconv.Atoi(AppRelease)

	return id
}
