// Diode Network Client
// Copyright 2019 IoT Blockchain Technology Corporation LLC (IBTC)
// Licensed under the Diode License, Version 1.0
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/diodechain/diode_go_client/config"
	"github.com/diodechain/diode_go_client/crypto"
	"github.com/diodechain/diode_go_client/db"
	"github.com/diodechain/diode_go_client/rpc"
	"github.com/diodechain/diode_go_client/util"
	"github.com/diodechain/log15"
)

var (
	version string = "development"
)

func init() {
	config.ParseFlag()
}

func main() {
	var socksServer *rpc.Server
	var proxyServer *rpc.ProxyServer
	var err error
	var pool *rpc.DataPool

	if version != "development" {
		doUpdate()
	}

	cfg := config.AppConfig
	if len(cfg.PublishedPorts) > 0 {
		pool = rpc.NewPoolWithPublishedPorts(cfg.PublishedPorts)
	} else {
		pool = rpc.NewPool()
	}

	printLabel("Diode Client version", version)

	// Initialize db
	clidb, err := db.OpenFile(cfg.DBPath)
	if err != nil {
		printError("Couldn't open database", err, 129)
	}
	db.DB = clidb

	if cfg.Command == "config" {
		activity := false
		if len(cfg.ConfigDelete) > 0 {
			activity = true
			for _, deleteKey := range cfg.ConfigDelete {
				db.DB.Del(deleteKey)
				printLabel("Deleted:", deleteKey)
			}
		}
		if len(cfg.ConfigSet) > 0 {
			activity = true
			for _, configSet := range cfg.ConfigSet {
				list := strings.Split(configSet, "=")
				if len(list) == 2 {
					var err error
					value := []byte(list[1])
					if util.IsHex(value) {
						value, err = util.DecodeString(list[1])
						if err != nil {
							printError("Couldn't decode hex string", err, 1)
						}
					}
					db.DB.Put(list[0], value)
					printLabel("Set:", list[0])
				} else {
					printError("Couldn't set value", fmt.Errorf("Expected -set name=value format"), 1)
				}
			}
		}
		if cfg.ConfigList || activity == false {
			printLabel("<KEY>", "<VALUE>")
			for _, name := range db.DB.List() {
				label := "<************************>"
				value, err := db.DB.Get(name)
				if err == nil && name != "private" {
					label = util.EncodeToString(value)
				}
				printLabel(name, label)
			}
		}

		os.Exit(0)
	}

	{
		lvbn, lvbh := rpc.LastValid()
		printLabel("Last valid block", fmt.Sprintf("%v %v", lvbn, util.EncodeToString(lvbh[:])))

		addr := crypto.PubkeyToAddress(rpc.LoadClientPubKey())
		printLabel("Client address", util.EncodeToString(addr[:]))

		fleetAddr, err := db.DB.Get("fleet")
		if err != nil {
			// Migration if existing
			fleetAddr, err = db.DB.Get("fleet_id")
			if err == nil {
				cfg.FleetAddr, err = util.DecodeAddress(string(fleetAddr))
				if err == nil {
					db.DB.Put("fleet", cfg.FleetAddr[:])
					db.DB.Del("fleet_id")
				}
			}
		} else {
			copy(cfg.FleetAddr[:], fleetAddr)
		}
		printLabel("Fleet address", util.EncodeToString(cfg.FleetAddr[:]))
	}

	// Connect to first server to respond
	wg := &sync.WaitGroup{}
	rpcAddrLen := len(cfg.RemoteRPCAddrs)
	c := make(chan *rpc.RPCClient, rpcAddrLen)
	wg.Add(rpcAddrLen)
	for _, RemoteRPCAddr := range cfg.RemoteRPCAddrs {
		go connect(c, RemoteRPCAddr, cfg, wg, pool)
	}

	// var client *rpc.RPCClient
	var client *rpc.RPCClient
	go func() {
		for rpcClient := range c {
			if client == nil && rpcClient != nil {
				cfg.Logger.Info(fmt.Sprintf("Connected to host: %s, validating...", rpcClient.Host()), "module", "main")
				isValid, err := rpcClient.ValidateNetwork()
				if isValid {
					client = rpcClient
				} else {
					if err != nil {
						cfg.Logger.Error(fmt.Sprintf("Network is not valid (err: %s), trying next...", err.Error()), "module", "main")
					} else {
						cfg.Logger.Error("Network is not valid for unknown reasons", "module", "main")
					}
					rpcClient.Close()
				}
			} else if rpcClient != nil {
				rpcClient.Close()
			}
			wg.Done()
		}
	}()
	wg.Wait()
	close(c)

	if client == nil {
		printError("Couldn't connect to any server", fmt.Errorf("server are not validated"), 129)
	}
	lvbn, _ := rpc.LastValid()
	cfg.Logger.Info(fmt.Sprintf("Network is validated, last valid block number: %d", lvbn), "module", "main")

	// check device access to fleet contract and registry
	clientAddr, err := client.GetClientAddress()
	if err != nil {
		cfg.Logger.Error(err.Error())
		return
	}

	// check device whitelist
	isDeviceWhitelisted, err := client.IsDeviceWhitelisted(clientAddr)
	if !isDeviceWhitelisted {
		cfg.Logger.Error(fmt.Sprintf("Device was not whitelisted: <%v>", err), "module", "main")
		return
	}

	// send ticket
	err = client.Greet()
	if err != nil {
		cfg.Logger.Error(err.Error(), "module", "main")
		return
	}

	// listen to signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGINT:
			closeDiode(client, socksServer, proxyServer, cfg)
		}
	}()

	socksConfig := &rpc.Config{
		Addr:            cfg.SocksServerAddr,
		FleetAddr:       cfg.FleetAddr,
		Blacklists:      cfg.Blacklists,
		Whitelists:      cfg.Whitelists,
		EnableProxy:     cfg.EnableProxyServer,
		ProxyServerAddr: cfg.ProxyServerAddr,
	}
	socksServer = client.NewSocksServer(socksConfig, pool)

	if cfg.EnableSocksServer {
		// start socks server
		if err := socksServer.Start(); err != nil {
			cfg.Logger.Error(err.Error(), "module", "main")
			return
		}
	}
	if cfg.EnableProxyServer {
		proxyConfig := rpc.ProxyConfig{
			EnableProxy:      cfg.EnableProxyServer,
			EnableSProxy:     cfg.EnableSProxyServer,
			ProxyServerAddr:  cfg.ProxyServerAddr,
			SProxyServerAddr: cfg.SProxyServerAddr,
			CertPath:         cfg.SProxyServerCertPath,
			PrivPath:         cfg.SProxyServerPrivPath,
			AllowRedirect:    cfg.AllowRedirectToSProxy,
		}
		// Start proxy server
		if proxyServer, err = rpc.NewProxyServer(socksServer, proxyConfig); err != nil {
			cfg.Logger.Error(err.Error(), "module", "main")
			return
		}
		if err := proxyServer.Start(); err != nil {
			cfg.Logger.Error(err.Error(), "module", "main")
			return
		}
	}

	for _, bind := range cfg.Binds {
		err = socksServer.StartBind(bind)
		if err != nil {
			cfg.Logger.Error(err.Error(), "module", "main")
			return
		}
	}

	// start
	client.Wait()
	closeDiode(client, socksServer, proxyServer, cfg)
}

func printLabel(label string, value string) {
	msg := fmt.Sprintf("%-20s : %-80s", label, value)
	config.AppConfig.Logger.Info(msg, "module", "main")
}

func printError(msg string, err error, status int) {
	config.AppConfig.Logger.Error(msg, "module", "main", "error", err)
	os.Exit(status)
}

func printInfo(msg string) {
	config.AppConfig.Logger.Info(msg, "module", "main")
}

func connect(c chan *rpc.RPCClient, host string, cfg *config.Config, wg *sync.WaitGroup, pool *rpc.DataPool) {
	client, err := rpc.DoConnect(host, cfg, pool)
	if err != nil {
		cfg.Logger.Error(fmt.Sprintf("Connection to host: %s failed: %+v", host, err), "module", "main")
		wg.Done()
	} else {
		c <- client
	}
}

func closeDiode(client *rpc.RPCClient, socksServer *rpc.Server, proxyServer *rpc.ProxyServer, cfg *config.Config) {
	if client.Started() {
		client.Close()
	}
	if socksServer.Started() {
		socksServer.Close()
	}
	if proxyServer != nil && proxyServer.Started() {
		proxyServer.Close()
	}
	handler := cfg.Logger.GetHandler()
	if closingHandler, ok := handler.(log15.ClosingHandler); ok {
		closingHandler.WriteCloser.Close()
	}
	os.Exit(0)
}
