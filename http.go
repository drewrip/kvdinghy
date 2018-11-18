package main

import (
	"encoding/json"
	"net"
	"net/http"
	"time"
	"fmt"
	"net/url"
	"github.com/hashicorp/raft"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type httpServer struct {
	address net.Addr
	node    *node
	logger  *zerolog.Logger
}

func (server *httpServer) Start() {
	server.logger.Info().Str("address", server.address.String()).Msg("Starting server")
	c := alice.New()
	c = c.Append(hlog.NewHandler(*server.logger))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("req.method", r.Method).
			Str("req.url", r.URL.String()).
			Int("req.status", status).
			Int("req.size", size).
			Dur("req.duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RemoteAddrHandler("req.ip"))
	c = c.Append(hlog.UserAgentHandler("req.useragent"))
	c = c.Append(hlog.RefererHandler("req.referer"))
	c = c.Append(hlog.RequestIDHandler("req.id", "Request-Id"))
	handler := c.Then(server)

	if err := http.ListenAndServe(server.address.String(), handler); err != nil {
		server.logger.Fatal().Err(err).Msg("Error running HTTP server")
	}
}

func (server *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path{


	// GET
	case "/get":
		if r.Method != http.MethodGet{
			w.WriteHeader(http.StatusMethodNotAllowed)
			break
		}
		
			w.WriteHeader(http.StatusOK)
			getKey := struct {
				Key string `json:"key"`
			}{}

			if err := json.NewDecoder(r.Body).Decode(&getKey); err != nil {
				server.logger.Error().Err(err).Msg("Bad request")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			response := struct {
				Value int `json:"value"`
			}{
				Value: server.node.fsm.KV[getKey.Key],
			}

			responseBytes, err := json.Marshal(response)
			if err != nil {
				server.logger.Error().Err(err).Msg("")
			}

			w.Write(responseBytes)






	// GET
	case "/dump":
		if r.Method != http.MethodGet{
			w.WriteHeader(http.StatusMethodNotAllowed)
			break
		}
		
		w.WriteHeader(http.StatusOK)

		allentries := []keyval{}
		for k,v := range server.node.fsm.KV{
			allentries = append(allentries, keyval{Key: k, Value: v})
		}

		responseBytes, err := json.Marshal(allentries)
		if err != nil {
			server.logger.Error().Err(err).Msg("")
		}

		w.Write(responseBytes)






	// POST
	case "/set":
		if r.Method != http.MethodPost{
			w.WriteHeader(http.StatusMethodNotAllowed)
			break
		}
		if server.node.raftNode.State() != raft.Leader{
			server.logger.Info().Msg("Forwarding Message To Leader")
			leaderAddr := string(server.node.raftNode.Leader())
			leaderParsed,err := net.ResolveTCPAddr("tcp", leaderAddr)
			if err != nil {
				server.logger.Error().Err(err).Msg("")
			}
			url := url.URL{
				Scheme: "http",
				Host:   leaderParsed.IP.String() + ":" + fmt.Sprintf("%d", leaderParsed.Port + 1000),
				Path:   "/set",
			}
			req, err := http.NewRequest(http.MethodPost, url.String(), nil)
			if err != nil {
				server.logger.Error().Err(err).Msg("")
			}
			req.Body = r.Body
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				server.logger.Error().Err(err).Msg("")
			}

			if resp.StatusCode != http.StatusOK {
					server.logger.Error().Err(err).Msg(fmt.Sprintf("non 200 status code: %d", resp.StatusCode))
			}
			break
		}
		request := keyval{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			server.logger.Error().Err(err).Msg("Bad request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body.Close()
		event := &event{
			Key: request.Key,
			Value: request.Value,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			server.logger.Error().Err(err).Msg("")
		}

		applyFuture := server.node.raftNode.Apply(eventBytes, 5*time.Second)
		if err := applyFuture.Error(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)






	// POST (Made in main.go)
	case "/join":
		if r.Method != http.MethodPost{
			w.WriteHeader(http.StatusMethodNotAllowed)
			break
		}
		peerAddress := r.Header.Get("Peer-Address")
		if peerAddress == "" {
			server.logger.Error().Msg("Peer-Address not set on request")
			w.WriteHeader(http.StatusBadRequest)
		}

		addPeerFuture := server.node.raftNode.AddVoter(
			raft.ServerID(peerAddress), raft.ServerAddress(peerAddress), 0, 0)
		if err := addPeerFuture.Error(); err != nil {
			server.logger.Error().
				Err(err).
				Str("peer.remoteaddr", peerAddress).
				Msg("Error joining peer to Raft")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		server.logger.Info().Str("peer.remoteaddr", peerAddress).Msg("Peer joined Raft")
		w.WriteHeader(http.StatusOK)






	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}