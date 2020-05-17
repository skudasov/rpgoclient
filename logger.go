package rpgoclient

import (
	"encoding/json"
	"go.uber.org/zap"
)

func NewLogger(level string) *zap.SugaredLogger {
	rawJSON := []byte(`{
	  "level": "` + level + `",
	  "encoding": "console",
	  "outputPaths": ["stdout", "/tmp/logs"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.Sugar()
}
