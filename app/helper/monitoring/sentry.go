package monitoring

import (
	"os"

	"github.com/getsentry/sentry-go"
)

func InitSentry() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Environment: os.Getenv("SENTRY_ENV"),
	}); err != nil {
		panic("Sentry initialization failed")
	}
}
