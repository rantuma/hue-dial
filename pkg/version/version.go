package version

var (
	Version   = "dev" //nolint:gochecknoglobals // build-time injected via ldflags
	BuildTime = ""    //nolint:gochecknoglobals // build-time injected via ldflags
)
