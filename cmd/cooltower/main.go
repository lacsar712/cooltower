package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lacsar712/cooltower/internal/app"
	"github.com/lacsar712/cooltower/internal/config"
	"github.com/lacsar712/cooltower/internal/web"
	"github.com/lacsar712/cooltower/internal/web/api"
)

func main() {
	once := flag.Bool("once", false, "run one control cycle then exit")
	webAddr := flag.String("web", "", "listen address for HMI (overrides config)")
	flag.Parse()

	cfg := config.LoadFromEnv()
	if *webAddr != "" {
		cfg.WebListenAddr = *webAddr
	}

	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cooltower: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.WebListenAddr != "" {
		go serveWeb(ctx, application, cfg.WebListenAddr)
	}

	if *once {
		if err := application.RunOnce(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "cooltower: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(application.StatusLine())
		return
	}

	ticker := time.NewTicker(cfg.ProcessTick())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := application.RunOnce(ctx); err != nil {
				log.Printf("cycle error: %v", err)
			}
		}
	}
}

func serveWeb(ctx context.Context, application *app.App, addr string) {
	srv := api.NewServer(application)
	handler, err := web.Handler(srv)
	if err != nil {
		log.Printf("web setup: %v", err)
		return
	}
	httpSrv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdown)
	}()
	log.Printf("cooltower HMI listening on %s", addr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("web: %v", err)
	}
}
