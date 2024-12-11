package log

import "go.uber.org/zap"

var Logger *zap.Logger

func init() {
	var err error
	Logger, err = NewLogger()
	if err != nil {
		panic(err)
	}
}

func NewLogger() (*zap.Logger, error) {
	pro := zap.NewProductionConfig()
	pro.OutputPaths = append(pro.OutputPaths,
		"./accountHandler.log")
	return pro.Build()
}
