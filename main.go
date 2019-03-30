package main

import (
	"github.com/MovieStoreGuy/skirmish/pkg/orchestra"
	"go.uber.org/zap"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	orc := orchestra.NewRunner(log)
	defer orc.Shutdown()
}
