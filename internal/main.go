package transcode

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/m1k1o/go-transcode/internal/api"
	"github.com/m1k1o/go-transcode/internal/config"
)

var Service *Main

func init() {
	Service = &Main{
		RootConfig:   &config.Root{},
		ServerConfig: &config.Server{},
	}
}

type Main struct {
	RootConfig   *config.Root
	ServerConfig *config.Server

	logger     zerolog.Logger
	hlsManager *api.HlsManagerCtx
}

func (main *Main) Preflight() {
	main.logger = log.With().Str("service", "main").Logger()
}

func (main *Main) Start() {
	config := main.ServerConfig
	main.hlsManager = api.New(config)
	main.hlsManager.Start()

}

func (main *Main) Shutdown() {
	var err error
	err = main.hlsManager.Shutdown()
	main.logger.Err(err).Msg("api manager shutdown")
}

func (main *Main) ServeCommand(cmd *cobra.Command, args []string) {
	main.logger.Info().Msg("starting main server")
	main.Start()
	main.logger.Info().Msg("main ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	main.logger.Warn().Msgf("received %s, attempting graceful shutdown", sig)
	main.Shutdown()
	main.logger.Info().Msg("shutdown complete")
}

func (main *Main) ConfigReload() {
	main.RootConfig.Set()
	main.ServerConfig.Set()
}
