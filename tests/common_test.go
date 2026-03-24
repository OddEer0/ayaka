package ayaka

import "github.com/OddEer0/ayaka"

func Filter[T any](data []T, fn func(T) bool) []T {
	result := make([]T, 0, len(data))
	for _, v := range data {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

var globalLogMsg = map[string]struct{}{
	ayaka.LogInfoInitAllJobsStarted:             {},
	ayaka.LogInfoInitAllJobsFinished:            {},
	ayaka.LogWarnInitJobsGracefulShutdownFailed: {},
	ayaka.LogInfoRunAllJobsStarted:              {},
	ayaka.LogInfoRunAllJobsFinished:             {},
	ayaka.LogWarnRunJobsGracefulShutdownFailed:  {},
}

func filterPickGlobalLogsMessages(msg string) bool {
	_, ok := globalLogMsg[msg]
	return ok
}
