package version

// Version is injected at build time via -ldflags.
// Defaults to "dev" for local builds.
var Version = "dev"
