package gcloudtracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestTracer(t *testing.T) {
	t.Run("tracer=success", func(t *testing.T) {
		tracer, err := NewTracer(
			context.Background(),
			WithProject("test_project"),
			WithLogger(&defaultLogger{}),
			WithAPIKey("test_api_key"),
		)
		assert.NoError(t, err)
		assert.NotNil(t, tracer)
	})

	t.Run("tracer=failed", func(t *testing.T) {
		tracer, err := NewTracer(
			context.Background(),
		)
		assert.Error(t, err)
		assert.Nil(t, tracer)
	})
}
