package ayaka

import (
	"context"
	"sync"
	"testing"
	"time"

	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/stretchr/testify/assert"
)

type testLogger struct {
	levels   []string
	messages []string
	infos    []map[string]any
	mutex    *sync.Mutex
}

type testLoggerData struct {
	levels   []string
	messages []string
	infos    []map[string]any
}

func (t *testLogger) Debug(_ context.Context, message string, info map[string]any) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.levels = append(t.levels, "debug")
	t.messages = append(t.messages, message)
	t.infos = append(t.infos, info)
}

func (t *testLogger) Info(_ context.Context, message string, info map[string]any) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.levels = append(t.levels, "info")
	t.messages = append(t.messages, message)
	t.infos = append(t.infos, info)
}

func (t *testLogger) Warn(_ context.Context, message string, info map[string]any) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.levels = append(t.levels, "warn")
	t.messages = append(t.messages, message)
	t.infos = append(t.infos, info)
}

func (t *testLogger) Error(_ context.Context, message string, info map[string]any) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.levels = append(t.levels, "error")
	t.messages = append(t.messages, message)
	t.infos = append(t.infos, info)
}

func (t *testLogger) getData() testLoggerData {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	lvlCopy := make([]string, len(t.levels))
	copy(lvlCopy, t.levels)
	messagesCopy := make([]string, len(t.messages))
	copy(messagesCopy, t.messages)
	infoCopy := make([]map[string]any, len(t.infos))
	copy(infoCopy, t.infos)
	return testLoggerData{
		levels:   lvlCopy,
		messages: messagesCopy,
		infos:    infoCopy,
	}
}

func newTestLogger() *testLogger {
	return &testLogger{
		levels:   make([]string, 0, 10),
		messages: make([]string, 0, 10),
		infos:    make([]map[string]any, 0, 10),
		mutex:    &sync.Mutex{},
	}
}

func TestLogger(t *testing.T) {
	logger := newTestLogger()

	ayaka.NewApp[*Container](&ayaka.Options[*Container]{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
		Container:   &Container{},
		Logger:      logger,
	}).WithConfig(&ayaka.Config{
		StartTimeout:    time.Second * 2,
		GracefulTimeout: time.Second * 3,
		Info: map[string]any{
			"test": "kek",
		},
	})

	data := logger.getData()

	assert.Equal(t, []string{"info"}, data.levels)
	assert.Equal(t, []string{"add new config"}, data.messages)
	assert.Equal(t, []map[string]any{
		{
			"config": &ayaka.Config{
				StartTimeout:    time.Second * 2,
				GracefulTimeout: time.Second * 3,
				Info: map[string]any{
					"test": "kek",
				},
			},
		},
	}, data.infos)
}
