package bootstrap

import "go.uber.org/zap"

func log() {
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
}
