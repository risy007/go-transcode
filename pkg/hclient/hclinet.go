package hclient

import (
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/m1k1o/go-transcode/internal/config"
	"github.com/m1k1o/go-transcode/pkg/channel"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type ClientCtx struct {
	logger      zerolog.Logger
	mu          sync.Mutex
	server_addr string
	server_port string
	token       string
	sn          string
	mid         string

	events struct {
		onStart  func()
		onCmdLog func(message string)
		onStop   func(err error)
	}

	shutdown chan interface{}
}

func New(c *config.HtsServer) *ClientCtx {
	return &ClientCtx{
		logger:      log.With().Str("module", "hclient").Logger(),
		shutdown:    make(chan interface{}),
		server_addr: c.HcsServer.ServerAddr,
		server_port: c.HcsServer.ServerPort,
		token:       c.HcsServer.Token,
		sn:          c.SN,
		mid:         c.MID,
	}
}

func (c *ClientCtx) init() error {
	if c.token == "" {
		return c.Login()
	}
	return nil
}

type TssLoginReq struct {
	SN  string `json:"sn"`
	MID string `json:"mid"`
}

func (c *ClientCtx) Login() error {
	var result interface{}
	var errmsg interface{}
	client := req.C().
		SetUserAgent("HTS-CLIENT"). // Chainable client settings.
		SetTimeout(5 * time.Second)
	url := fmt.Sprintf("http://%s:%s/base/tsslogin", c.server_addr, c.server_port)

	resp, err := client.R().
		SetHeader("Accept", "application/vnd.github.v3+json").
		SetBody(&TssLoginReq{SN: c.sn, MID: c.mid}).
		SetSuccessResult(&result).
		SetErrorResult(&errmsg).
		EnableDump().
		Post(url)

	if err != nil { // Error handling.
		c.logger.Err(err)
		return err
	}

	if resp.IsErrorState() { // Status code >= 400.
		c.logger.Err(err)
		return err
	}

	if resp.IsSuccessState() { // Status code is between 200 and 299.
		c.logger.Info().Msg("HTS Login to HCS Success！")

	}
	return nil
}

func (c *ClientCtx) GetChannels(url string) (map[string]string, channel.ErrorMessage) {
	var errmsg channel.ErrorMessage

	client := req.C().
		SetUserAgent("HTS-CLIENT"). // Chainable client settings.
		SetTimeout(5 * time.Second)

	var chns map[string]string
	var cs []channel.Channel

	resp, err := client.R().
		SetHeader("Accept", "application/vnd.github.v3+json"). // Chainable request settings.
		SetSuccessResult(&cs).                                 // Unmarshal response body into userInfo automatically if status code is between 200 and 299.
		SetErrorResult(&errmsg).                               // Unmarshal response body into errMsg automatically if status code >= 400.
		EnableDump().                                          // Enable dump at request level, only print dump content if there is an error or some unknown situation occurs to help troubleshoot.
		Get(url)

	if err != nil { // Error handling.
		c.logger.Err(err)
		return nil, errmsg
	}

	if resp.IsErrorState() { // Status code >= 400.
		c.logger.Err(err)
		return nil, errmsg
	}

	if resp.IsSuccessState() { // Status code is between 200 and 299.
		c.logger.Info().Msg("获取频道列表成功！")
		for _, c := range cs {
			chns[c.Code] = c.Url
		}
		return chns, errmsg
	}
	return nil, errmsg
}

func (c *ClientCtx) OnStart(event func()) {
	c.events.onStart = event
}

func (c *ClientCtx) OnCmdLog(event func(message string)) {
	c.events.onCmdLog = event
}

func (c *ClientCtx) OnStop(event func(err error)) {
	c.events.onStop = event
}
