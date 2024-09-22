package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultAddr                = "127.0.0.1:8080"
	defaultDir                 = "/pub"
	defaultHealthcheckEndpoint = "/_health"

	defaultIdleTimeoutSec       = 120
	defaultReadTimeoutSec       = 60
	defaultWriteTimeoutSec      = 1800
	defaultReadHeaderTimeoutSec = 30
)

var (
	CommitHash   string
	BuildVersion string
)

type Conf struct {
	Address             string
	Directory           string
	TlsCertFile         string
	TlsKeyFile          string
	HealthcheckEndpoint string

	IdleTimeoutSec       int
	ReadTimeoutSec       int
	WriteTimeoutSec      int
	ReadHeaderTimeoutSec int
}

func (c *Conf) Validate() error {
	_, err := net.ResolveTCPAddr("tcp", c.Address)
	if err != nil {
		return fmt.Errorf("invalid listen address provided: %w", err)
	}

	_, err = os.Stat(c.Directory)
	if err != nil {
		return fmt.Errorf("directory %q does not exist", c.Directory)
	}

	if len(c.HealthcheckEndpoint) > 0 && !strings.HasPrefix(c.HealthcheckEndpoint, "/") {
		return fmt.Errorf("health check endpoint must start with \"/\"")
	}

	if len(c.TlsCertFile) > 0 && len(c.TlsCertFile) == 0 || len(c.TlsCertFile) == 0 && len(c.TlsCertFile) > 0 {
		return errors.New("either both tls cert and key must be provided or none at all")
	}

	if c.WriteTimeoutSec <= 0 || c.IdleTimeoutSec <= 0 || c.ReadTimeoutSec <= 0 || c.ReadHeaderTimeoutSec <= 0 {
		return errors.New("timeout must be > 0")
	}

	return nil
}

func (c *Conf) UseTls() bool {
	return len(c.TlsCertFile) > 0 && len(c.TlsKeyFile) > 0
}

func (c *Conf) getTlsConf() (*tls.Config, error) {
	if !c.UseTls() {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion:     tls.VersionTLS13,
		GetCertificate: c.loadCert,
	}

	// don't wait for lazy loading the tls keypair when the first request hits the server to
	// verify whether the files exist and are readable or not.
	_, err := c.loadCert(nil)
	if err != nil {
		return nil, err
	}

	return tlsConfig, nil
}

func (c *Conf) loadCert(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certificate, err := tls.LoadX509KeyPair(c.TlsCertFile, c.TlsKeyFile)
	if err != nil {
		slog.Error("user-defined client certificates could not be loaded", "error", err)
	}
	return &certificate, err
}

func envOrDefaultInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if len(val) == 0 {
		return defaultVal
	}

	converted, err := strconv.Atoi(val)
	if err != nil {
		slog.Warn("could not convert to string", "var", key, "val", val)
		return defaultVal
	}

	return converted
}

func envOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		val = defaultVal
	}
	return val
}

func getConf() Conf {
	conf := Conf{
		Address:              envOrDefault("APLOS_ADDR", defaultAddr),
		Directory:            envOrDefault("APLOS_DIRECTORY", defaultDir),
		TlsCertFile:          envOrDefault("APLOS_TLS_CRT_FILE", ""),
		TlsKeyFile:           envOrDefault("APLOS_TLS_KEY_FILE", ""),
		HealthcheckEndpoint:  envOrDefault("APLOS_HEALTHCHECK_ENDPOINT", defaultHealthcheckEndpoint),
		IdleTimeoutSec:       envOrDefaultInt("APLOS_TIMEOUT_IDLE", defaultIdleTimeoutSec),
		ReadHeaderTimeoutSec: envOrDefaultInt("APLOS_TIMEOUT_READ_HEADER", defaultReadHeaderTimeoutSec),
		ReadTimeoutSec:       envOrDefaultInt("APLOS_TIMEOUT_READ", defaultReadTimeoutSec),
		WriteTimeoutSec:      envOrDefaultInt("APLOS_TIMEOUT_WRITE", defaultWriteTimeoutSec),
	}

	flag.StringVar(&conf.Address, "a", conf.Address, "The address to run the server on")
	flag.StringVar(&conf.Directory, "d", conf.Directory, "The directory to serve")
	flag.StringVar(&conf.TlsCertFile, "c", conf.TlsCertFile, "File that contains the TLS certificate")
	flag.StringVar(&conf.TlsKeyFile, "k", conf.TlsCertFile, "File that contains the TLS private key")
	flag.StringVar(&conf.HealthcheckEndpoint, "p", conf.HealthcheckEndpoint, "Endpoint where to expose the healthcheck handler. Set to \"\" to disable the health check handler.")
	flag.IntVar(&conf.IdleTimeoutSec, "idle-timeout", conf.IdleTimeoutSec, "Set the idle timeout in seconds")
	flag.IntVar(&conf.ReadHeaderTimeoutSec, "read-header-timeout", conf.ReadHeaderTimeoutSec, "Set the read-header timeout in seconds")
	flag.IntVar(&conf.ReadTimeoutSec, "read-timeout", conf.ReadTimeoutSec, "Set the read timeout in seconds")
	flag.IntVar(&conf.WriteTimeoutSec, "write-timeout", conf.WriteTimeoutSec, "Set the write timeout in seconds")
	flag.Parse()
	return conf
}

func main() {
	slog.Info("Starting aplos", "version", BuildVersion, "commit", CommitHash)
	conf := getConf()
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(conf.Directory)))
	if len(conf.HealthcheckEndpoint) > 0 {
		mux.HandleFunc(conf.HealthcheckEndpoint, func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("OK"))
		})
	}

	tlsConfig, err := conf.getTlsConf()
	if err != nil {
		log.Fatalf("invalid tls config: %v", err)
	}

	server := http.Server{
		Addr:              conf.Address,
		Handler:           mux,
		TLSConfig:         tlsConfig,
		IdleTimeout:       time.Duration(conf.IdleTimeoutSec) * time.Second,
		ReadTimeout:       time.Duration(conf.ReadTimeoutSec) * time.Second,
		WriteTimeout:      time.Duration(conf.WriteTimeoutSec) * time.Second,
		ReadHeaderTimeout: time.Duration(conf.ReadHeaderTimeoutSec) * time.Second,
	}

	go func() {
		var err error
		if conf.UseTls() {
			slog.Info("Starting TLS server", "directory", conf.Directory, "addr", conf.Address)
			err = server.ListenAndServeTLS("", "")
		} else {
			slog.Info("Starting server", "directory", conf.Directory, "addr", conf.Address)
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("can not start server: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	slog.Info("Caught signal, shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
