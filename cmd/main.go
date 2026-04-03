package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"runtime"

	"github.com/orimono/ito"
	"github.com/orimono/shutter/internal/collector"
	"github.com/orimono/shutter/internal/collector/subsystem"
	"github.com/orimono/shutter/internal/config"
	"github.com/orimono/shutter/internal/logger"
	"github.com/orimono/shutter/internal/reporter"
	"github.com/orimono/shutter/internal/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()
	logger.Init(cfg.LogLevel)

	nodeID, err := ito.GenerateNodeID("")
	if err != nil {
		slog.Error("failed to generate node ID", "err", err)
		return
	}

	manager := collector.NewManager(nodeID)
	manager.AddCollector(&subsystem.MemoryCollector{})

	client := ws.NewClient(cfg)
	go client.Run(ctx)

	rep := reporter.NewReporter(client)

	go func() {
		hostname, err := os.Hostname()
		if err != nil {
			slog.Error("failed to get hostname", "err", err)
			return
		}

		joinPacket := &ito.JoinPacket{
			NodeID:       nodeID,
			Hostname:     hostname,
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			Tags:         cfg.Tags,
			TaskManifest: manager.Manifest(),
		}

		data, err := json.Marshal(joinPacket)
		if err != nil {
			slog.Error("failed to marshal JoinPacket", "err", err)
			return
		}

		<-client.Ready()
		rep.Send(data)
	}()

	go manager.Start(ctx)

	go func() {
		for t := range manager.Out() {
			data, err := json.Marshal(t)
			if err != nil {
				slog.Error("failed to marshal telemetry", "err", err)
				continue
			}
			if err := client.Send(data); err != nil {
				slog.Warn("failed to send telemetry", "err", err)
			}
		}
	}()

	<-ctx.Done()
}
