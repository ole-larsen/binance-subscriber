// Package config to configure server. The part of server package.
// Copyright 2024 The Oleg Nazarov. All rights reserved.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	Host        string
	Instruments []string
	Port        int
}

type Opts struct {
	APtr *string
	IPtr *string
}

var (
	config = &Config{}
	once   sync.Once
)

// GetConfig rewrites using singleton pattern.
func GetConfig() *Config {
	once.Do(func() {
		f := parseFlags()
		config = InitConfig(
			WithAddress(os.Getenv("ADDRESS"), f.APtr),
			WithInstruments(os.Getenv("INSTRUMENTS"), f.IPtr),
		)
	})

	return config
}

func (c *Config) Reload(opts ...func(*Config)) {
	for _, opt := range opts {
		opt(c)
	}
}

func InitConfig(opts ...func(*Config)) *Config {
	c := &Config{}
	c.Reload(opts...)

	return c
}

func parseFlags() Opts {
	flags := Opts{
		APtr: flag.String("a", "localhost:8080", "HTTP-server endpoint (default localhost:8080)"),
		IPtr: flag.String("i", "btcusdt@depth", "streams (default btcusdt@depth)"),
	}

	flag.Parse()

	return flags
}

func WithAddress(a string, aPtr *string) func(*Config) {
	return func(c *Config) {
		if a == "" && aPtr != nil {
			a = *aPtr
		}

		addr := strings.Split(a, ":")

		const reqLen = 2

		if len(addr) != reqLen {
			panic(fmt.Errorf("wrong a parameters"))
		}

		if addr[0] != "localhost" {
			c.Host = addr[0]
		}

		port, err := strconv.Atoi(addr[1])
		if err != nil {
			panic(fmt.Errorf("wrong a parameters"))
		}

		c.Port = port
	}
}

func WithInstruments(i string, iPtr *string) func(*Config) {
	return func(c *Config) {
		if i == "" && iPtr != nil {
			i = *iPtr
		}

		split := strings.Split(i, ",")
		instruments := make([]string, len(split))

		for i := range split {
			instruments[i] = strings.TrimSpace(split[i])
		}

		c.Instruments = instruments
	}
}
