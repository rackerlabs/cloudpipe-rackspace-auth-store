package main

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	lruks "github.com/rgbkrk/lru-key-store"
)

// Context provides shared state among route handlers.
type Context struct {
	Settings

	KeyStore *lruks.KeyStore
}

// Settings contains configuration options loaded from the environment.
type Settings struct {
	InternalPort   int
	ExternalPort   int
	LogLevel       string
	LogColors      bool
	InternalCACert string
	InternalCert   string
	InternalKey    string
	ExternalCert   string
	ExternalKey    string
	CacheSize      int
}

// Load reads configuration settings from the environment and validates them.
func (c *Context) Load() error {
	if err := envconfig.Process("AUTH", &c.Settings); err != nil {
		return err
	}

	if c.InternalPort == 0 {
		c.InternalPort = 9001
	}

	if c.ExternalPort == 0 {
		c.ExternalPort = 9000
	}

	if c.CacheSize == 0 {
		c.CacheSize = 10000
	}

	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

	if c.InternalCACert == "" {
		c.InternalCACert = "/certificates/ca.pem"
	}

	if c.InternalCert == "" {
		c.InternalCert = "/certificates/auth-store-cert.pem"
	}

	if c.InternalKey == "" {
		c.InternalKey = "/certificates/auth-store-key.pem"
	}

	if c.ExternalCert == "" {
		c.ExternalCert = "/certificates/external-cert.pem"
	}

	if c.ExternalKey == "" {
		c.ExternalKey = "/certificates/external-key.pem"
	}

	if _, err := log.ParseLevel(c.LogLevel); err != nil {
		return err
	}

	return nil
}

// NewContext loads configuration from the environment and applies immediate, global settings.
func NewContext() (*Context, error) {
	c := &Context{}

	if err := c.Load(); err != nil {
		return c, err
	}

	// Configure the logging level and formatter.

	level, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		return c, err
	}
	log.SetLevel(level)

	log.SetFormatter(&log.TextFormatter{
		ForceColors: c.LogColors,
	})

	// Summarize the loaded settings.

	log.WithFields(log.Fields{
		"internal port":    c.InternalPort,
		"external port":    c.ExternalPort,
		"logging level":    c.LogLevel,
		"log with color":   c.LogColors,
		"internal CA cert": c.InternalCACert,
		"internal cert":    c.InternalCert,
		"internal key":     c.InternalKey,
		"external cert":    c.ExternalCert,
		"external key":     c.ExternalKey,
		"cache size":       c.CacheSize,
	}).Info("Initializing with loaded settings.")

	c.KeyStore, err = lruks.New(c.CacheSize)

	if err != nil {
		return c, err
	}

	return c, nil
}

// InternalListenAddr generates an address to bind the private net/http server to.
func (c *Context) InternalListenAddr() string {
	return fmt.Sprintf(":%d", c.InternalPort)
}

// ExternalListenAddr generates an address to bind the public net/http server to.
func (c *Context) ExternalListenAddr() string {
	return fmt.Sprintf(":%d", c.ExternalPort)
}

// BindContext returns an http.HandlerFunc that binds a ContextHandler to a specific Context.
func BindContext(c *Context, handler ContextHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { handler(c, w, r) }
}

// ContextHandler is an HTTP HandlerFunc that accepts an additional parameter containing the
// server context.
type ContextHandler func(c *Context, w http.ResponseWriter, r *http.Request)
