/*
 * SPDX-FileCopyrightText: Bosch Rexroth AG
 *
 * SPDX-License-Identifier: MIT
 */
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"datalayer.client/pkg/client"
	"datalayer.client/pkg/utils"
	"github.com/boschrexroth/ctrlx-datalayer-golang/v2/pkg/datalayer"
)

func main() {

	// Create a ctrlX Data Layer system
	system := datalayer.NewSystem("")
	fmt.Println("starting ctrlX Data Layer system")
	system.Start(false)

	// Create a remote string with the parameters according to your environment (see remote.go in package utils)
	remote := utils.NewRemote().Ip("10.0.2.2").SslPort(8443).Build()

	// Create a ctrlX Data Layer client
	c := system.Factory().CreateClient(remote)

	// We use a client wrapper
	dlc := client.NewDatalayerClient(c)

	// Graceful shutdown
	defer func() {
		dlc.Delete()
		datalayer.DeleteSystem(system)

		fmt.Println("stopped ctrlX Data Layer system")
	}()

	if !c.IsConnected() {
		// client not connected
		fmt.Println("connecting to ", remote, " failed")
		return
	}

	// Run the client
	go func() {
		tick := time.NewTicker(time.Second)
		for range tick.C {
			if !c.IsConnected() {
				return
			}
			dlc.Run()
		}
	}()

	// Handle process exit event (SIGTERM)
	sigterm := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(sigterm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	// Wait for process termination
	<-sigterm
}
