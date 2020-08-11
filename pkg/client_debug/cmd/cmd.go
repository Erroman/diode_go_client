// Diode Network Client
// Copyright 2019 IoT Blockchain Technology Corporation LLC (IBTC)
// Licensed under the Diode License, Version 1.0
package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/diodechain/diode_go_client/config"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	clientDebugCmd = &cobra.Command{
		Use:   "client_debug",
		Short: "A debug program for testing diode network.",
		Long:  `This is a debug program for connecting the target through diode network concurrently.`,
		RunE:  clientDebugHandler,
	}
	cfg                      = &Config{}
	ErrUnsupportTransport    = fmt.Errorf("unsupported transport, make sure you use these options (proxy, sproxy, socks5)")
	ErrFailedSetRlimitNofile = fmt.Errorf("cannot set rlimit nofile")
)

func clientDebugHandler(cmd *cobra.Command, args []string) (err error) {
	var proxyTransport *http.Transport = &http.Transport{
		Proxy: http.ProxyURL(&url.URL{
			Scheme: "socks5:",
			Host:   "localhost:33",
		}),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	var tlsTransport *http.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	var wg sync.WaitGroup
	var transport *http.Transport
	if cfg.EnableTransport {
		var prox *url.URL
		if cfg.EnableSocks5Transport {
			prox, _ = url.Parse(fmt.Sprintf("socks5://%s", cfg.SocksServerAddr()))
		} else if cfg.EnableProxyTransport {
			prox, _ = url.Parse(fmt.Sprintf("http://%s", cfg.ProxyServerAddr()))
		} else if cfg.EnableSProxyTransport {
			prox, _ = url.Parse(fmt.Sprintf("https://%s", cfg.SProxyServerAddr()))
			proxyTransport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		} else {
			err = ErrUnsupportTransport
			return
		}
		proxyTransport.Proxy = http.ProxyURL(prox)
		transport = proxyTransport
	} else {
		transport = tlsTransport
	}
	if cfg.RlimitNofile > 0 {
		if err = config.SetRlimitNofile(cfg.RlimitNofile); err != nil {
			return
		}
	}
	log.Printf("Start to connect %d times", cfg.Conn)
	wg.Add(cfg.Conn)
	for i := 0; i < cfg.Conn; i++ {
		go func(j int) {
			client := &http.Client{}
			client.Transport = transport
			resp, err := client.Get(cfg.Target)
			if err != nil {
				log.Printf("Failed to get target #%d: %s\n", j, err.Error())
				wg.Done()
				return
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read from body #%d: %s\n", j, err.Error())
				resp.Body.Close()
				wg.Done()
				return
			}
			if cfg.Verbose {
				log.Printf("Content #%d:  %s\n", j, string(body))
			} else {
				log.Printf("Fetch #%d successfully\n", j)
			}
			resp.Body.Close()
			wg.Done()
		}(i + 1)
	}
	wg.Wait()
	return
}

func init() {
	clientDebugCmd.PersistentFlags().StringVarP(&cfg.Target, "target", "a", "http://pi-taipei.diode", "test target")
	clientDebugCmd.PersistentFlags().BoolVarP(&cfg.EnableTransport, "transport", "b", true, "enable http transport")
	clientDebugCmd.PersistentFlags().IntVarP(&cfg.Conn, "conn", "c", 100, "total connection concurrently")
	clientDebugCmd.PersistentFlags().BoolVarP(&cfg.EnableSocks5Transport, "socks5", "d", true, "enable socks5 transport")
	clientDebugCmd.PersistentFlags().StringVarP(&cfg.SocksServerHost, "socksd_host", "e", "127.0.0.1", "host of socks server")
	clientDebugCmd.PersistentFlags().IntVarP(&cfg.SocksServerPort, "socksd_port", "f", 1080, "port of socks server")
	clientDebugCmd.PersistentFlags().BoolVarP(&cfg.EnableProxyTransport, "proxy", "g", false, "enable proxy transport")
	clientDebugCmd.PersistentFlags().StringVarP(&cfg.ProxyServerHost, "proxy_host", "i", "127.0.0.1", "host of proxy server")
	clientDebugCmd.PersistentFlags().IntVarP(&cfg.ProxyServerPort, "proxy_port", "j", 80, "port of proxy server")
	clientDebugCmd.PersistentFlags().BoolVarP(&cfg.EnableSProxyTransport, "sproxy", "k", false, "enable secure proxy transport")
	clientDebugCmd.PersistentFlags().StringVarP(&cfg.SProxyServerHost, "sproxy_host", "l", "127.0.0.1", "host of secure proxy server")
	clientDebugCmd.PersistentFlags().IntVarP(&cfg.SProxyServerPort, "sproxy_port", "m", 443, "port of secure proxy server")
	clientDebugCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "enable verbose to show the response body")
	clientDebugCmd.PersistentFlags().IntVarP(&cfg.RlimitNofile, "rlimit_nofile", "r", 0, "specify the file descriptor numbers that can be opened by this process")
}

// Execute the client_debug command
func Execute() error {
	return clientDebugCmd.Execute()
}
