package config_test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/ole-larsen/binance-subscriber/internal/server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultAddress = "localhost:8080"
)

func Test_NewConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "#1 server config test. check only one config was created",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetConfig()

			configPtr := fmt.Sprintf("%p", cfg)

			assert.NotEqual(t, "localhost", cfg.Host)
			assert.NotEqual(t, "", cfg.Port)

			for i := 0; i < 10; i++ {
				c := config.GetConfig()
				cPtr := fmt.Sprintf("%p", c)
				assert.Equal(t, configPtr, cPtr)
			}

			cfg = &config.Config{}

			assert.Empty(t, cfg)

			cfg = config.InitConfig()

			assert.Empty(t, cfg)
		})
	}
}

func Test_InitConfig(t *testing.T) {
	type args struct {
		opts []func(*config.Config)
	}

	address := defaultAddress

	tests := []struct {
		name string
		args args
	}{
		{
			name: "test init config functional options with env variables",
			args: args{
				opts: []func(*config.Config){
					config.WithAddress(address, nil),
				},
			},
		},
		{
			name: "test init config functional options with parsed flags",
			args: args{
				opts: []func(*config.Config){
					config.WithAddress("", &address),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.InitConfig(tt.args.opts...)
			assert.Equal(t, "", cfg.Host)
			assert.Equal(t, 8080, cfg.Port)
		})
	}
}

func Test_parseFlags(t *testing.T) {
	address := defaultAddress

	tests := []struct {
		want config.Opts
		name string
	}{
		{
			name: "test parseFlags",
			want: config.Opts{
				APtr: &address,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, flag.Lookup("a"))
			assert.NotEmpty(t, flag.Lookup("i"))

			// check default values
			aFlag, ok := flag.Lookup("a").Value.(flag.Getter).Get().(string)
			require.True(t, ok)
			assert.Equal(t, *tt.want.APtr, aFlag)
		})
	}
}

func Test_withAddress(t *testing.T) {
	type args struct {
		aPtr *string
		a    string
	}

	address := defaultAddress
	invalidAddressFormat := "localhost"       // Missing port
	invalidAddressPort := "localhost:notPort" // Invalid port

	tests := []struct {
		name      string
		args      args
		wantHost  string
		wantPort  int
		wantPanic bool
	}{
		{
			name: "valid address with environment variable",
			args: args{
				a:    address,
				aPtr: nil,
			},
			wantHost:  "",
			wantPort:  8080,
			wantPanic: false,
		},
		{
			name: "valid address with command line argument",
			args: args{
				a:    "",
				aPtr: &address,
			},
			wantHost:  "",
			wantPort:  8080,
			wantPanic: false,
		},
		{
			name: "invalid address format",
			args: args{
				a:    invalidAddressFormat,
				aPtr: nil,
			},
			wantPanic: true,
		},
		{
			name: "invalid address format with command line argument",
			args: args{
				a:    "",
				aPtr: &invalidAddressFormat,
			},
			wantPanic: true,
		},
		{
			name: "invalid port number",
			args: args{
				a:    invalidAddressPort,
				aPtr: nil,
			},
			wantPanic: true,
		},
		{
			name: "invalid port number with command line argument",
			args: args{
				a:    "",
				aPtr: &invalidAddressPort,
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					config.InitConfig(config.WithAddress(tt.args.a, tt.args.aPtr))
				})
			} else {
				cfg := config.InitConfig(config.WithAddress(tt.args.a, tt.args.aPtr))
				assert.Equal(t, tt.wantHost, cfg.Host)
				assert.Equal(t, tt.wantPort, cfg.Port)
			}
		})
	}
}
