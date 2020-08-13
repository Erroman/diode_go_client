// Diode Network Client
// Copyright 2019 IoT Blockchain Technology Corporation LLC (IBTC)
// Licensed under the Diode License, Version 1.0
package rpc

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/diodechain/diode_go_client/pkg/diode/config"
)

func unzip(in []byte) string {
	out := bytes.Buffer{}
	gz := bytes.NewBuffer(in)

	zr, err := gzip.NewReader(gz)
	if err != nil {
		config.AppConfig.Logger.Error(fmt.Sprintf("failed to unzip: %s", err.Error()))
		os.Exit(129)
	}
	if _, err := io.Copy(&out, zr); err != nil {
		config.AppConfig.Logger.Error(fmt.Sprintf("failed to unzip: %s", err.Error()))
		os.Exit(129)
	}
	return out.String()

}

func unzip64(in []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(unzip(in)))
}

func image(code int) (string, string) {
	switch {
	case code >= 200 && code < 300:
		return unicornHappy, "Welcome to the your Web3 proxy. To learn how to create your own Web3 content visit us at <a href='https://diode.io'>DIODE</a>"
	case code >= 400 && code < 500:
		return unicornPee, "Um, looks like you don't have access to this resource, or it may be down. To learn how to create your own Web3 content visit us at <a href='https://diode.io'>DIODE</a>"
	case code >= 500 && code < 600:
		return unicornSad, "Uh you found a bug, please copy the url and above error message and <a href='https://github.com/diodechain/diode_go_client/issues/new'>submit them to us here</a> with a short description of what happened."
	default:
		return unicornDrinking, "We have no idea what happened here."
	}
}

func Page(title string, code int, codeMessage string, more string) string {
	image, hint := image(code)
	return fmt.Sprintf(template, title, strconv.FormatInt(int64(code), 10), codeMessage, more, hint, image)
}
