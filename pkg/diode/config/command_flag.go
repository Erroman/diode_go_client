// Diode Network Client
// Copyright 2019 IoT Blockchain Technology Corporation LLC (IBTC)
// Licensed under the Diode License, Version 1.0
package config

import (
	"flag"
	// "fmt"
	"os"
	// "reflect"
	// "sort"
	// "strings"
)

type CommandFlag struct {
	Name        string
	HelpText    string
	ExampleText string
	Flag        flag.FlagSet
}

func command(name string, commandFlags *[]CommandFlag) *CommandFlag {
	for _, commandFlag := range *commandFlags {
		if commandFlag.Name == name {
			return &commandFlag
		}
	}
	return nil
}

// Parse the args (flag.Args()[1:]) with the given command flag
func (commandFlag *CommandFlag) Parse(args []string) {
	err := commandFlag.Flag.Parse(args)
	if !commandFlag.Flag.Parsed() {
		commandFlag.Flag.Usage()
		os.Exit(2)
	} else if err != nil {
		// always exit
		if err == flag.ErrHelp {
			os.Exit(2)
		} else {
			commandFlag.Flag.Usage()
			os.Exit(2)
		}
	}
}

// func makeCommandFlags(cfg *Config) *[]CommandFlag {
// 	commandFlags := make([]CommandFlag, 0, 10)

// 	// diode publish
// 	publishCommandFlag := CommandFlag{
// 		Name:        "publish",
// 		HelpText:    `  Publish ports of the local device to the Diode Network.`,
// 		ExampleText: `  diode publish -public 80:80 -public 8080:8080 -protected 3000:3000 -protected 3001:3001 -private 22:22,0x......,0x...... -private 33:33,0x......,0x......`,
// 	}
// 	publishCommandFlag.Flag.Var(&cfg.PublicPublishedPorts, "public", "expose ports to public users, so that user could connect to")
// 	publishCommandFlag.Flag.Var(&cfg.ProtectedPublishedPorts, "protected", "expose ports to protected users (in fleet contract), so that user could connect to")
// 	publishCommandFlag.Flag.Var(&cfg.PrivatePublishedPorts, "private", "expose ports to private users, so that user could connect to")
// 	publishCommandFlag.Flag.StringVar(&cfg.SocksServerHost, "proxy_host", "127.0.0.1", "host of socksd proxy server")
// 	publishCommandFlag.Flag.IntVar(&cfg.SocksServerPort, "proxy_port", 1080, "port of socksd proxy server")
// 	publishCommandFlag.Flag.BoolVar(&cfg.EnableSocksServer, "socksd", false, "enable socksd proxy server")
// 	commandFlags = append(commandFlags, publishCommandFlag)

// 	// diode config
// 	configCommandFlag := CommandFlag{
// 		Name:        "config",
// 		HelpText:    `  Manage variables in the local config store.`,
// 		ExampleText: `  diode config -delete lvbn2 -delete lvbn`,
// 	}
// 	configCommandFlag.Flag.Var(&cfg.ConfigDelete, "delete", "deletes the given variable from the config")
// 	configCommandFlag.Flag.BoolVar(&cfg.ConfigList, "list", false, "list all stored config keys")
// 	configCommandFlag.Flag.BoolVar(&cfg.ConfigUnsafe, "unsafe", false, "display private keys (disabled by default)")
// 	configCommandFlag.Flag.Var(&cfg.ConfigSet, "set", "sets the given variable in the config")
// 	commandFlags = append(commandFlags, configCommandFlag)

// 	// diode socksd
// 	socksdCommandFlag := CommandFlag{
// 		Name:        "socksd",
// 		HelpText:    `  Enable a socks proxy for use with browsers and other apps.`,
// 		ExampleText: `  diode socksd -socksd_port 8082 -socksd_host 127.0.0.1`,
// 	}
// 	socksdCommandFlag.Flag.StringVar(&cfg.SocksServerHost, "socksd_host", "127.0.0.1", "host of socks server listening to")
// 	socksdCommandFlag.Flag.IntVar(&cfg.SocksServerPort, "socksd_port", 1080, "port of socks server listening to")
// 	socksdCommandFlag.Flag.StringVar(&cfg.SocksFallback, "fallback", "localhost", "how to resolve web2 addresses")
// 	commandFlags = append(commandFlags, socksdCommandFlag)

// 	// diode httpd
// 	httpdCommandFlag := CommandFlag{
// 		Name:        "httpd",
// 		HelpText:    `  Enable a public http server as is used by the "diode.link" website`,
// 		ExampleText: `  diode httpd -httpd_port 8080 -httpsd_port 443 -secure -certpath ./cert.pem -privpath ./priv.pem`,
// 	}
// 	httpdCommandFlag.Flag.StringVar(&cfg.SocksServerHost, "proxy_host", "127.0.0.1", "host of socksd proxy server")
// 	httpdCommandFlag.Flag.IntVar(&cfg.SocksServerPort, "proxy_port", 1080, "port of socksd proxy server")
// 	httpdCommandFlag.Flag.BoolVar(&cfg.EnableSocksServer, "socksd", false, "enable socksd proxy server")
// 	httpdCommandFlag.Flag.StringVar(&cfg.ProxyServerHost, "httpd_host", "127.0.0.1", "host of httpd server listening to")
// 	httpdCommandFlag.Flag.IntVar(&cfg.ProxyServerPort, "httpd_port", 80, "port of httpd server listening to")
// 	httpdCommandFlag.Flag.StringVar(&cfg.SProxyServerHost, "httpsd_host", "127.0.0.1", "host of httpsd server listening to")
// 	httpdCommandFlag.Flag.IntVar(&cfg.SProxyServerPort, "httpsd_port", 443, "port of httpsd server listening to")
// 	httpdCommandFlag.Flag.StringVar(&cfg.SProxyServerCertPath, "certpath", "./priv/cert.pem", "Pem format of certificate file path of httpsd secure server")
// 	httpdCommandFlag.Flag.StringVar(&cfg.SProxyServerPrivPath, "privpath", "./priv/priv.pem", "Pem format of private key file path of httpsd secure server")
// 	httpdCommandFlag.Flag.BoolVar(&cfg.EnableSProxyServer, "secure", false, "enable httpsd server")
// 	httpdCommandFlag.Flag.BoolVar(&cfg.AllowRedirectToSProxy, "allow_redirect", false, "allow redirect all http transmission to httpsd")
// 	commandFlags = append(commandFlags, httpdCommandFlag)

// 	// diode reset
// 	initCommandFlag := CommandFlag{
// 		Name:        "reset",
// 		HelpText:    `  Initialize a new account and a new fleet contract in the network. WARNING deletes current credentials!`,
// 		ExampleText: `  diode reset`,
// 	}
// 	initCommandFlag.Flag.BoolVar(&cfg.Experimental, "experimental", false, "send transactions of fleet deployment and device allowlist at seme time")
// 	commandFlags = append(commandFlags, initCommandFlag)

// 	// diode bns
// 	bnsCommandFlag := CommandFlag{
// 		Name:        "bns",
// 		HelpText:    `  Register/Update name service on diode blockchain.`,
// 		ExampleText: `  diode bns -register hello-world=0x......`,
// 	}
// 	bnsCommandFlag.Flag.StringVar(&cfg.BNSRegister, "register", "", "Register a new BNS name with <name>=<address>.")
// 	bnsCommandFlag.Flag.StringVar(&cfg.BNSLookup, "lookup", "", "Lookup a given BNS name.")
// 	commandFlags = append(commandFlags, bnsCommandFlag)

// 	// diode time
// 	timeCommandFlag := CommandFlag{
// 		Name:        "time",
// 		HelpText:    `  Lookup the current time from the blockchain consensus.`,
// 		ExampleText: `  diode time`,
// 	}
// 	commandFlags = append(commandFlags, timeCommandFlag)

// 	// Finishing up, and sorting.
// 	for i, flags := range commandFlags {
// 		name := flags.Name
// 		commandFlags[i].Flag.Usage = func() {
// 			printUsage(*command(name, &commandFlags))
// 		}
// 	}
// 	sort.Slice(commandFlags, func(i, j int) bool { return commandFlags[i].Name < commandFlags[j].Name })
// 	return &commandFlags
// }

// // isZeroValue determines whether the string represents the zero
// // value for a flag.
// func isZeroValue(f *flag.Flag, value string) bool {
// 	// Build a zero value of the flag's Value type, and see if the
// 	// result of calling its String method equals the value passed in.
// 	// This works unless the Value type is itself an interface type.
// 	typ := reflect.TypeOf(f.Value)
// 	var z reflect.Value
// 	if typ.Kind() == reflect.Ptr {
// 		z = reflect.New(typ.Elem())
// 	} else {
// 		z = reflect.Zero(typ)
// 	}
// 	return value == z.Interface().(flag.Value).String()
// }

// func isStringValue(f *flag.Flag) bool {
// 	typ := reflect.TypeOf(f.Value)
// 	if typ.Kind() != reflect.Ptr {
// 		return false

// 	}
// 	return typ.Elem().String() == "flag.stringValue"
// }

// func printUsage(command CommandFlag) {
// 	fmt.Printf("Name\n  diode %s -%s\n\n", command.Name, command.HelpText)
// 	fmt.Printf("SYNOPSYS\n  diode %s <args>\n\n", command.Name)
// 	printCommandDefaults(&command, 0)
// }

// func printCommandDefaults(commandFlag *CommandFlag, indent int) {
// 	s := fmt.Sprintf("%*sARGS\n", indent, "")
// 	commandFlag.Flag.VisitAll(func(f *flag.Flag) {
// 		s += fmt.Sprintf("%*s-%s", indent+2, "", f.Name) // Two spaces before -; see next two comments.
// 		name, usage := flag.UnquoteUsage(f)
// 		if len(name) > 0 {
// 			s += " " + name
// 		}
// 		// Boolean flags of one ASCII letter are so common we
// 		// treat them specially, putting their usage on the same line.
// 		if len(s) <= 4 { // space, space, '-', 'x'.
// 			s += "\t"
// 		} else {
// 			// Four spaces before the tab triggers good alignment
// 			// for both 4- and 8-space tab stops.
// 			s += "\n    \t"
// 		}
// 		s += strings.ReplaceAll(usage, "\n", "\n    \t")

// 		if !isZeroValue(f, f.DefValue) {
// 			if ok := isStringValue(f); ok {
// 				// put quotes on the value
// 				s += fmt.Sprintf(" (default %q)", f.DefValue)
// 			} else {
// 				s += fmt.Sprintf(" (default %v)", f.DefValue)
// 			}
// 		}
// 		s += "\n"
// 	})
// 	s += fmt.Sprintf("%*sEXAMPLE\n%*s%s\n", indent, "", indent, "", commandFlag.ExampleText)
// 	fmt.Fprint(commandFlag.Flag.Output(), s)
// }
