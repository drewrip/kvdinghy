package main

import (
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
	"sync"
	"github.com/hashicorp/raft"
	"github.com/rs/zerolog"
)


type TestNodeConf struct {
	RaftPort int
	HTTPPort int
	JoinAddress string
	Bootstrap bool
}

func (tnc *TestNodeConf)StartNode(){
	
	logger := zerolog.New(os.Stdout)

	nodeLogger := logger.With().Str("component", "node").Logger()
	node, err := NewTestNode(tnc, &nodeLogger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error configuring node: %s", err)
		os.Exit(1)
	}

	if tnc.JoinAddress != "" {
		retryJoin := func() error {
			url := url.URL{
				Scheme: "http",
				Host:   tnc.JoinAddress,
				Path:   "join",
			}

			req, err := http.NewRequest(http.MethodPost, url.String(), nil)
			if err != nil {
				return err
			}
			req.Header.Add("Peer-Address", "localhost:" + fmt.Sprintf("%d", tnc.RaftPort))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("non 200 status code: %d", resp.StatusCode)
			}

			return nil
		}

		for {
			if err := retryJoin(); err != nil {
				logger.Error().Err(err).Str("component", "join").Msg("Error joining cluster")
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
	}

	httpAddr := &net.TCPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: tnc.HTTPPort,
	}

	httpLogger := logger.With().Str("component", "http").Logger()
	httpServer := &httpServer{
		node:    node,
		address: httpAddr,
		logger:  &httpLogger,
	}

	httpServer.Start()
}

func NewTestNode(tnc *TestNodeConf, log *zerolog.Logger) (*node, error) {
	fsm := &fsm{
		KV: make(map[string]int),
	}

	TCPHTTPADDR := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: tnc.HTTPPort}
	TCPRAFTADDR := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: tnc.RaftPort}

	raftConfig := raft.DefaultConfig()
	
	raftConfig.Logger = stdlog.New(log, "", 0)

	raftConfig.LocalID = raft.ServerID("127.0.0.1:" + fmt.Sprintf("%d", tnc.RaftPort))
	transportLogger := log.With().Str("component", "raft-transport").Logger()
	transport, err := raftTransport(TCPRAFTADDR, transportLogger)
	snapshotStore := raft.NewInmemSnapshotStore()

	logStore := raft.NewInmemStore()

	stableStore := raft.NewInmemStore()

	raftNode, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}
	if tnc.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		raftNode.BootstrapCluster(configuration)
	}

	return &node{
		config:   &Config{HTTPAddress: TCPHTTPADDR, Bootstrap: tnc.Bootstrap, RaftAddress: TCPRAFTADDR, JoinAddress: tnc.JoinAddress},
		raftNode: raftNode,
		log:      log,
		fsm:      fsm,
	}, nil
}

func InitCluster(NumberOfNodes int, StartPort int){
	var wg sync.WaitGroup
	startnode := TestNodeConf{RaftPort: StartPort, HTTPPort: StartPort+1000, Bootstrap: true, JoinAddress: ""}
	go startnode.StartNode()
	// Make sure the leader is started
	time.Sleep(5 * time.Second)
	for i:=1; i<NumberOfNodes; i++{
		wg.Add(1)
		nodetemp := TestNodeConf{RaftPort: StartPort+i , HTTPPort: StartPort+1000+i, JoinAddress: "127.0.0.1:"+fmt.Sprintf("%d", StartPort+1000),}
		go nodetemp.StartNode()
	}
	wg.Wait()
}