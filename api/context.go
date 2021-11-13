package api

import (
	"context"
	"time"
)

const RequestTimeLimit = 500 * time.Millisecond

func Context() (context.Context, func()) {
	return context.WithTimeout(context.Background(), RequestTimeLimit)
}
