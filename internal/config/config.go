package config

import (
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

type Root struct {
	Debug   bool
	PProf   bool
	CfgFile string
}

func (Root) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().BoolP("debug", "d", false, "enable debug mode")
	if err := viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pprof", false, "enable pprof endpoint available at /debug/pprof")
	if err := viper.BindPFlag("pprof", cmd.PersistentFlags().Lookup("pprof")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("config", "", "configuration file path")
	if err := viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")); err != nil {
		return err
	}

	return nil
}

func (s *Root) Set() {
	s.Debug = viper.GetBool("debug")
	s.PProf = viper.GetBool("pprof")
	s.CfgFile = viper.GetString("config")
}

type HtsServer struct {
	BaseDir   string `yaml:"basedir,omitempty"`
	RamDisk   string `yaml:"ramdisk,omitempty"`
	Profiles  string `yaml:"profiles,omitempty"`
	SN        string `yaml:"sn,omitempty"`
	MID       string `yaml:"mid,omitempty"`
	HcsServer HCS
}

func (HtsServer) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("basedir", "", "base directory for assets and profiles")
	if err := viper.BindPFlag("basedir", cmd.PersistentFlags().Lookup("basedir")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("profiles", "", "hardware encoding profiles to load for ffmpeg (default, nvidia)")
	if err := viper.BindPFlag("profiles", cmd.PersistentFlags().Lookup("profiles")); err != nil {
		return err
	}

	return nil
}

func (s *HtsServer) Set() {

	s.BaseDir = viper.GetString("basedir")
	if s.BaseDir == "" {
		if _, err := os.Stat("/etc/transcode"); os.IsNotExist(err) {
			cwd, _ := os.Getwd()
			s.BaseDir = cwd
		} else {
			s.BaseDir = "/etc/transcode"
		}
	}

	s.RamDisk = viper.GetString("ramdisk")
	if s.RamDisk == "" {
		if _, err := os.Stat("/mnt/ram"); os.IsNotExist(err) {
			s.RamDisk = "/tmp"
		}
	}

	s.Profiles = viper.GetString("profiles")
	if s.Profiles == "" {
		// TODO: issue #5
		s.Profiles = fmt.Sprintf("%s/profiles", s.BaseDir)
	}
	s.SN = viper.GetString("sn")
	if s.SN == "" {
		panic("sn is empty，please check the configs！")
	}

	s.MID = viper.GetString("mid")

	cMID, _ := machineid.ProtectedID(s.SN)

	if s.MID == "" {
		s.MID = cMID
		viper.Set("MID", s.MID)
		if err := viper.SafeWriteConfig(); err != nil {
			panic(err)
		}
	}

	if s.MID != cMID {
		panic("machine code is changed, please contact admin!")
	}
	if err := viper.UnmarshalKey("hcs", &s.HcsServer); err != nil {
		panic(err)
	}
}

func (s *HtsServer) AbsPath(elem ...string) string {
	// prepend base path
	elem = append([]string{s.BaseDir}, elem...)
	return path.Join(elem...)
}

type HCS struct {
	ServerAddr string `mapstructure:"server_addr"`
	ServerPort string `mapstructure:"server_port"`

	Token string `mapstructure:"token"`
}
