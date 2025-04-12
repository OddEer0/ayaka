package grpc_job

import (
	"context"
	"encoding/json"
	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/OddEer0/ayaka/ecosystem"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"testing"
)

type testOut struct {
	out []byte
}

func (t *testOut) Write(p []byte) (n int, err error) {
	t.out = p
	return len(p), nil
}

func TestAppLoggerFromSlog(t *testing.T) {
	out := &testOut{}
	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	log := ecosystem.NewAppLoggerWithSlog(logger)

	testAppLogger(t, log, out, "msg")
}

func testAppLogger(t *testing.T, log ayaka.Logger, out *testOut, msgKey string) {
	ctx := context.Background()
	log.Debug(ctx, "kekw", map[string]interface{}{
		"foo": "bar",
		"baz": "qux",
	})

	var data map[string]interface{}
	err := json.Unmarshal(out.out, &data)
	assert.NoError(t, err)

	assert.Equal(t, "bar", data["foo"])
	assert.Equal(t, "qux", data["baz"])
	assert.Equal(t, "kekw", data[msgKey])
	assert.Equal(t, "debug", strings.ToLower(data["level"].(string)))

	log.Info(ctx, "kekw", map[string]interface{}{
		"foo": "bar",
		"baz": "qux",
		"hoho": map[string]interface{}{
			"hehe": "lol",
			"kek":  "xd",
		},
	})
	data = map[string]interface{}{}
	err = json.Unmarshal(out.out, &data)
	assert.NoError(t, err)

	assert.Equal(t, "bar", data["foo"])
	assert.Equal(t, "qux", data["baz"])
	assert.Equal(t, "kekw", data[msgKey])
	assert.Equal(t, map[string]interface{}{"hehe": "lol", "kek": "xd"}, data["hoho"])
	assert.Equal(t, "info", strings.ToLower(data["level"].(string)))

	log.Warn(ctx, "kekw", map[string]interface{}{
		"foo": "bar",
		"baz": "qux",
		"hoho": map[string]interface{}{
			"hehe": "lol",
			"kek":  "xd",
		},
	})
	data = map[string]interface{}{}
	err = json.Unmarshal(out.out, &data)
	assert.NoError(t, err)

	assert.Equal(t, "bar", data["foo"])
	assert.Equal(t, "qux", data["baz"])
	assert.Equal(t, "kekw", data[msgKey])
	assert.Equal(t, map[string]interface{}{"hehe": "lol", "kek": "xd"}, data["hoho"])
	assert.Equal(t, "warn", strings.ToLower(data["level"].(string)))

	log.Error(ctx, "kekw", map[string]interface{}{
		"foo": "bar",
		"baz": "qux",
		"hoho": map[string]interface{}{
			"hehe": "lol",
			"kek":  "xd",
		},
	})
	data = map[string]interface{}{}
	err = json.Unmarshal(out.out, &data)
	assert.NoError(t, err)

	assert.Equal(t, "bar", data["foo"])
	assert.Equal(t, "qux", data["baz"])
	assert.Equal(t, "kekw", data[msgKey])
	assert.Equal(t, map[string]interface{}{"hehe": "lol", "kek": "xd"}, data["hoho"])
	assert.Equal(t, "error", strings.ToLower(data["level"].(string)))

	log.Info(ctx, "kekw2", nil)
	data = map[string]interface{}{}
	err = json.Unmarshal(out.out, &data)
	assert.NoError(t, err)

	assert.Equal(t, nil, data["foo"])
	assert.Equal(t, nil, data["baz"])
	assert.Equal(t, "kekw2", data[msgKey])
	assert.Equal(t, nil, data["hoho"])
	assert.Equal(t, "info", strings.ToLower(data["level"].(string)))
}
