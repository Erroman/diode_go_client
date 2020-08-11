// Diode Network Client
// Copyright 2019 IoT Blockchain Technology Corporation LLC (IBTC)
// Licensed under the Diode License, Version 1.0
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/kierdavis/ansi"
	"github.com/spf13/cobra"
	"github.com/diodechain/go-update"
	"github.com/diodechain/go-update/progress"
	"github.com/diodechain/go-update/stores/github"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update diode client to the latest version.",
		Long:  "Update diode client to the latest version.",
		RunE:  updateHandler,
	}
	ErrFailedToUpdateClient = fmt.Errorf("failed to update diode client")
)

func updateHandler(cmd *cobra.Command, args []string) (err error) {
	ret := doUpdate()
	if ret != 0 {
		err = ErrFailedToUpdateClient
	}
	return
}

func doUpdate() int {
	m := &update.Manager{
		Command: "diode",
		Store: &github.Store{
			Owner:   "diodechain",
			Repo:    "diode_go_client",
			Version: version,
		},
	}

	if runtime.GOOS == "windows" {
		m.Command += ".exe"
	}

	tarball, ok := download(m)
	if !ok {
		return 0
	}

	// searching for binary in path
	bin, err := exec.LookPath(m.Command)
	if err != nil {
		// just update local file
		bin = os.Args[0]
	}

	dir := filepath.Dir(bin)
	if err := m.InstallTo(tarball, dir); err != nil {
		printError("Error installing", err)
		return 129
	}

	cmd := path.Join(dir, m.Command)
	fmt.Printf("Updated, restarting %s...\n", cmd)

	update.Restart(cmd)
	return 0
}

func download(m *update.Manager) (string, bool) {
	ansi.HideCursor()
	defer ansi.ShowCursor()

	printInfo("Checking for updates...")

	// fetch the new releases
	releases, err := m.LatestReleases()
	if err != nil {
		printInfo(fmt.Sprintf("Error fetching releases: %s", err))
		return "", false
	}

	// no updates
	if len(releases) == 0 {
		printInfo("No updates")
		return "", false
	}

	// latest release
	latest := releases[0]
	printInfo(fmt.Sprintf("Found version %s > %s\n", latest.Version, version))

	a := latest.FindZip(runtime.GOOS, runtime.GOARCH)
	if a == nil {
		printInfo(fmt.Sprintf("No binary for your system (%s_%s)", runtime.GOOS, runtime.GOARCH))
		return "", false
	}

	// whitespace
	fmt.Println()

	// download tarball to a tmp dir
	tarball, err := a.DownloadProxy(progress.Reader)
	if err != nil {
		printError("Error downloading", err)
	}

	return tarball, true
}
