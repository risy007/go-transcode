package api

import (
  "fmt"
  "github.com/m1k1o/go-transcode/pkg/hls"
  "github.com/rs/zerolog/log"
  "os"
  "os/exec"
  "path"

  "github.com/m1k1o/go-transcode/internal/config"
)

var hlsManagers map[string]hls.Manager = make(map[string]hls.Manager)

type HlsManagerCtx struct {
  config *config.HtsServer
}

func New(config *config.HtsServer) *HlsManagerCtx {
  return &HlsManagerCtx{
    config: config,
  }
}

func (manager *HlsManagerCtx) Start() {
  /*	lp := manager.config
  	rdisk := manager.config.RamDisk
  	for _, s := range manager.config.Streams {
  		for _, p := range lp {
  			pp, err := manager.ProfilePath("hls", p)
  			if err != nil {
  				log.Warn().Err(err).Msg("转码模板脚本不存在!")
  				return
  			}
  			ID := fmt.Sprintf("%s_%s", pp, s)
  			runPath := fmt.Sprintf("%s/%s/%s", rdisk, "/", ID)

  			hlsm, ok := hlsManagers[ID]
  			if !ok {
  				// create new manager
  				hlsm = hls.New(func() *exec.Cmd {
  					// get transcode cmd
  					cmd, err := manager.transcodeStart(pp, s)
  					if err != nil {
  						log.Error().Err(err).Msg("启动转码进程失败")
  					}
  					return cmd
  				})
  				hlsm.SetRunPath(runPath)
  				err := hlsm.Start()
  				if err != nil {
  					log.Warn().Err(err).Str("profilePath", pp).Str("url", s).Msg("转码命令启动失败！")
  					return
  				}
  				hlsManagers[ID] = hlsm
  			}
  		}
  	}*/
}

func (manager *HlsManagerCtx) Shutdown() error {
  // stop all hls managers
  for _, hlsm := range hlsManagers {
    hlsm.Stop()
  }
  return nil
}

func (manager *HlsManagerCtx) ProfilePath(folder string, profile string) (string, error) {

  profilePath := path.Join(manager.config.Profiles, folder, fmt.Sprintf("%s.sh", profile))
  if _, err := os.Stat(profilePath); os.IsNotExist(err) {
    log.Info().Str("profilepath", profilePath).Msg("脚本文件不存在！")
    return "", err
  }
  return profilePath, nil
}

// Call ProfilePath before
func (a *HlsManagerCtx) transcodeStart(profilePath string, input string) (*exec.Cmd, error) {

  log.Info().Str("profilePath", profilePath).Str("url", input).Msg("转码开始!")

  return exec.Command(profilePath, input), nil
}
