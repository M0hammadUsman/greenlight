package data

import (
	"context"
	"time"
)

func newQueryContext(timeoutInSec int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(timeoutInSec)*time.Second)
}
