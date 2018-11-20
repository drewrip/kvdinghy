package main

import (
	"fmt"
	"net"

	multierror "github.com/hashicorp/go-multierror"
	template "github.com/hashicorp/go-sockaddr/template"
	flag "github.com/ogier/pflag"
)

type RawConfig struct {
	BindAddress string
	JoinAddress string
	RaftPort    int
	HTTPPort    int
	Bootstrap   bool
}

type Config struct {
	RaftAddress net.Addr
	HTTPAddress net.Addr
	JoinAddress string
	Bootstrap   bool
}

type ConfigError struct {
	ConfigurationPoint string
	Err                error
}

func (err *ConfigError) Error() string {
	return fmt.Sprintf("%s: %s", err.ConfigurationPoint, err.Err.Error())
}

func resolveConfig(rawConfig *RawConfig) (*Config, error) {
	var errors *multierror.Error

	// Bind address
	var bindAddr net.IP
	resolvedBindAddr, err := template.Parse(rawConfig.BindAddress)
	if err != nil {
		configErr := &ConfigError{
			ConfigurationPoint: "bind-address",
			Err:                err,
		}
		errors = multierror.Append(errors, configErr)
	} else {
		bindAddr = net.ParseIP(resolvedBindAddr)
		if bindAddr == nil {
			err := fmt.Errorf("cannot parse IP address: %s", resolvedBindAddr)
			configErr := &ConfigError{
				ConfigurationPoint: "bind-address",
				Err:                err,
			}
			errors = multierror.Append(errors, configErr)
		}
	}

	// Raft port
	if rawConfig.RaftPort < 1 || rawConfig.RaftPort > 65536 {
		configErr := &ConfigError{
			ConfigurationPoint: "raft-port",
			Err:                fmt.Errorf("port numbers must be 1 < port < 65536"),
		}
		errors = multierror.Append(errors, configErr)
	}

	// Construct Raft Address
	raftAddr := &net.TCPAddr{
		IP:   bindAddr,
		Port: rawConfig.RaftPort,
	}

	// HTTP port
	if rawConfig.HTTPPort < 1 || rawConfig.HTTPPort > 65536 {
		configErr := &ConfigError{
			ConfigurationPoint: "http-port",
			Err:                fmt.Errorf("port numbers must be 1 < port < 65536"),
		}
		errors = multierror.Append(errors, configErr)
	}

	// Construct HTTP Address
	httpAddr := &net.TCPAddr{
		IP:   bindAddr,
		Port: rawConfig.HTTPPort,
	}

	if err := errors.ErrorOrNil(); err != nil {
		return nil, err
	}

	return &Config{
		JoinAddress: rawConfig.JoinAddress,
		RaftAddress: raftAddr,
		HTTPAddress: httpAddr,
		Bootstrap:   rawConfig.Bootstrap,
	}, nil
}

func readRawConfig() (*RawConfig, bool, int, int) {
	var config RawConfig
	var test bool
	var num int
	var startport int
	flag.StringVarP(&config.BindAddress, "bind-address", "a",
		"127.0.0.1", "IP Address on which to bind")

	flag.IntVarP(&config.RaftPort, "raft-port", "r",
		7000, "Port on which to bind Raft")

	flag.IntVarP(&config.HTTPPort, "http-port", "h",
		8000, "Port on which to bind HTTP")

	flag.StringVar(&config.JoinAddress, "join",
		"", "Address of another node to join")

	flag.BoolVar(&config.Bootstrap, "bootstrap",
		false, "Bootstrap the cluster with this node")

	flag.BoolVar(&test, "test",
		false, "Put Dinghy in test mode")

	flag.IntVarP(&num, "size", "s",
		7, "Number of nodes in the test Raft cluster")

	flag.IntVarP(&startport, "start-port", "p",
		7000, "Starting port for test Raft cluster")

	flag.Parse()
	return &config, test, num, startport
}
