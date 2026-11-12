// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(schedulerCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "worker",
		Short: "Run durable workers without HTTP or shared scheduling",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runBackground(cmd, false) },
	})
}

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Runs scheduled background work without the HTTP server",
	Args:  cobra.NoArgs,
	RunE:  func(cmd *cobra.Command, _ []string) error { return runBackground(cmd, true) },
}

func runBackground(cmd *cobra.Command, scheduler bool) error {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	initialize.FullInitWithoutAsync()
	config.ServiceEnableScheduler.Set(scheduler)
	initialize.StartAsync()
	log.Infof("Background workers started (scheduler=%t)", scheduler)
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ServiceShutdownTimeout.GetDuration())
	defer cancel()
	return initialize.Shutdown(shutdownCtx)
}
