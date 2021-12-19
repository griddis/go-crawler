package crawler

import (
	"context"
	"net/http"
	"testing"

	"github.com/griddis/atlant_test/tools/logging"
)

func TestCrawler_Close(t *testing.T) {
	logger := logging.NewLogger("debug", "2006-01-02T15:04:05.999999999Z07:00")
	ctx := logging.WithContext(context.Background(), logger)
	type fields struct {
		ctx         context.Context
		concurrency int
		client      *http.Client
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"successClose",
			fields{
				ctx,
				0,
				&http.Client{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCrawler(tt.fields.ctx, tt.fields.concurrency, tt.fields.client)
			c.Close()
		})
	}
}

func TestNewCrawler(t *testing.T) {
	logger := logging.NewLogger("debug", "2006-01-02T15:04:05.999999999Z07:00")
	ctx := logging.WithContext(context.Background(), logger)
	type args struct {
		ctx         context.Context
		concurrency int
		client      *http.Client
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			"new",
			args{
				ctx,
				10,
				&http.Client{},
			},
			10,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCrawler(tt.args.ctx, tt.args.concurrency, tt.args.client); (len(got.Workers) != tt.want) != tt.wantErr {
				t.Errorf("NewCrawler() concurrency = %d, want %d", len(got.Workers), tt.want)
			}
		})
	}
}
