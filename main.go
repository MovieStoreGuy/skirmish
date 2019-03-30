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

var planPath string

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
	orc := orchestra.NewRunner(log)
	signal.GlobalHandler().Register(func() {
		if err := orc.Shutdown(); err != nil {
			log.Error("Issue with shutting down orchestrator", zap.Error(err))
		}
	})
	plan, err := types.LoadPlan(planPath)
	if err != nil {
		log.Panic("Invalid plan path defined", zap.Error(err))
	}
	if err = orc.Execute(plan); err != nil {
		log.Error("Issue executing plan", zap.Error(err))
	}
	signal.GlobalHandler().Done()
}
