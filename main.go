package main

import (
	"context"
	"flag"
	"syscall"

	"github.com/MovieStoreGuy/skirmish/pkg/orchestra"
	"github.com/MovieStoreGuy/skirmish/pkg/signal"
	"github.com/MovieStoreGuy/skirmish/pkg/types"

	"go.uber.org/zap"
)

var (
	planPath string
)

func init() {
	flag.StringVar(&planPath, "plan-path", "", "the path to the plan to run")
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go signal.GlobalHandler().Await(ctx, cancel, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGINT)

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	defer signal.GlobalHandler().Finalise()
	defer cancel()
	orc, err := orchestra.NewRunner(ctx, cancel, log)
	if err != nil {
		log.Panic("Failed to create new orchestra runner", zap.Error(err))
	}
	log.Info("Running info", zap.Any("orchestrator", orc))
	signal.GlobalHandler().Register(func() {
		if err := orc.Shutdown(); err != nil {
			log.Error("Issue with shutting down orchestrator", zap.Error(err))
		}
	})
	plan, err := types.LoadPlan(planPath)
	if err != nil {
		log.Error("Invalid plan path defined", zap.Error(err))
		return
	}
	if err = orc.Execute(plan); err != nil {
		log.Error("Issue executing plan", zap.Error(err))
	}
	log.Info("finished execute")
}
