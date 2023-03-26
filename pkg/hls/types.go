package hls

type Manager interface {
	Start() error
	Stop()
	Cleanup()
	SetRunPath(string) error

	OnStart(event func())
	OnCmdLog(event func(message string))
	OnStop(event func(err error))
}
