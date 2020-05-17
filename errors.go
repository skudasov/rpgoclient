package rpgoclient

import "errors"

var (
	noLaunchIdErr               = errors.New("launch is not started, no LaunchId")
	logNotAttachableToLaunchErr = errors.New("cannot attach log to launch item, only to test items")
	responseErr                 = errors.New("failed to perform request")
	httpRetriesReachedErr       = errors.New("http max retries reached")
)
