package sshtunnel

import (
	"context"
	"errors"
	"testing"
)

func TestIsExpectedShutdown(t *testing.T) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name string
		ctx  context.Context
		err  error
		want bool
	}{
		{
			name: "context canceled error",
			ctx:  context.Background(),
			err:  context.Canceled,
			want: true,
		},
		{
			name: "canceled context with docker kill exit",
			ctx:  cancelCtx,
			err:  errors.New("exit status 137"),
			want: true,
		},
		{
			name: "signal error",
			ctx:  context.Background(),
			err:  errors.New("signal: killed"),
			want: true,
		},
		{
			name: "unexpected error",
			ctx:  context.Background(),
			err:  errors.New("exit status 137"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExpectedShutdown(tt.ctx, tt.err); got != tt.want {
				t.Fatalf("isExpectedShutdown() = %v, want %v", got, tt.want)
			}
		})
	}
}
