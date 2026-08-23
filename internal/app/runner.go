package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lacsar712/cooltower/internal/config"
)

func RunCLI() int {
	once := flag.Bool("once", false, "run one control cycle then exit")
	flag.Parse()

	cfg := config.LoadFromEnv()
	application, err := New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cooltower: %v\n", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *once {
		if err := application.RunOnce(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "cooltower: %v\n", err)
			return 1
		}
		fmt.Println(application.StatusLine())
		return 0
	}

	ticker := time.NewTicker(cfg.ProcessTick())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return 0
		case <-ticker.C:
			if err := application.RunOnce(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "cycle error: %v\n", err)
			}
		}
	}
}
