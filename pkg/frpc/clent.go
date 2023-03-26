package frpc

import (
  "github.com/fatedier/frp/client"
  "github.com/fatedier/frp/pkg/config"
  "github.com/fatedier/frp/pkg/util/log"
)

func startService(
  cfg config.ClientCommonConf,
  pxyCfgs map[string]config.ProxyConf,
  visitorCfgs map[string]config.VisitorConf,
  cfgFile string,
) (err error) {
  log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel,
    cfg.LogMaxDays, cfg.DisableLogColor)

  if cfgFile != "" {
    log.Trace("start frpc service for config file [%s]", cfgFile)
    defer log.Trace("frpc service for config file [%s] stopped", cfgFile)
  }
  svr, errRet := client.NewService(cfg, pxyCfgs, visitorCfgs, cfgFile)
  if errRet != nil {
    err = errRet
    return
  }

  closedDoneCh := make(chan struct{})
  shouldGracefulClose := cfg.Protocol == "kcp" || cfg.Protocol == "quic"
  // Capture the exit signal if we use kcp or quic.
  if shouldGracefulClose {
    go handleSignal(svr, closedDoneCh)
  }

  err = svr.Run()
  if err == nil && shouldGracefulClose {
    <-closedDoneCh
  }
  return
}
