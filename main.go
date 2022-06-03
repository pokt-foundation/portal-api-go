package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/portal-api-go/relay"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/pokt-foundation/portal-api-go/session"
	"github.com/pokt-foundation/portal-api-go/sticky"
	"github.com/pokt-foundation/portal-api-go/web"
)

const (
	webServerPort = 8090
	loggingLevel  = "info"
)

type settings struct {
	RpcUrls    []string
	PrivateKey string
	Port       int
	LogLevel   logger.Level
}

func gatherSettings(args []string) (settings, error) {
	var (
		urls  string
		level string
		s     settings
	)
	fs := flag.NewFlagSet("PortalAPI", flag.ContinueOnError)
	fs.StringVar(&urls, "rpcUrls", "", "Comma-separated list of RPC URLs")
	fs.StringVar(&s.PrivateKey, "privateKey", "", "Private key used for signing relays")
	fs.IntVar(&s.Port, "port", webServerPort, "Port to listen on")
	fs.StringVar(&level, "logLevel", loggingLevel, "Logging level: accepted values are warn, info, and debug")

	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		return settings{}, err
	}

	logLevel, err := logger.ParseLevel(level)
	if err != nil {
		fmt.Printf("Invalid logging level: %q, set to info.", level)
		s.LogLevel = logger.InfoLevel
	} else {
		s.LogLevel = logLevel
	}

	rpcUrls := strings.Split(urls, ",")
	if len(rpcUrls) < 1 {
		return settings{}, fmt.Errorf("Invalid list of RPC URLs: %q\n", urls)
	}
	s.RpcUrls = rpcUrls
	return s, nil
}

func main() {
	log := logger.New()
	settings, err := gatherSettings(os.Args)
	if err != nil {
		log.WithFields(logger.Fields{"error": err}).Warn("Error gathering settings")
		os.Exit(1)
	}
	log.SetLevel(settings.LogLevel)

	repo, err := repository.NewRepository("/tmp", log)
	if err != nil {
		fmt.Printf("Error setting up repository: %v\n", err)
	}

	sessionManager := session.NewSessionManager(settings.RpcUrls)

	relayerSettings := relay.FreemiumSettings()
	relayerSettings.DefaultStickyOptions = repository.StickyOptions{
		Duration:       30,
		UseRPCID:       true,
		RpcIDThreshold: 2,
	}
	relayerSettings.DefaultClientStickyOptions = sticky.StickyClient{}

	relayer, err := relay.NewRelayServer(
		settings.RpcUrls,
		settings.PrivateKey,
		relayerSettings,
		repo,
		sessionManager,
		log,
	)
	if err != nil {
		log.WithFields(logger.Fields{"error": err}).Warn("Error creating relayer")
		return
	}

	log.Info("Starting http server")
	http.HandleFunc("/", web.GetHttpServer(relayer, log))
	http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil)
}
