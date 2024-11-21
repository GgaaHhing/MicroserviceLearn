package internal

import "go.uber.org/zap"

func init() {
	development, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(development)
}
