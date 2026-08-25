// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package cmd

import (
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(webPushCmd)
	webPushCmd.AddCommand(webPushGenerateKeysCmd)
}

var webPushCmd = &cobra.Command{
	Use:   "webpush",
	Short: "Manage Web Push configuration",
}

var webPushGenerateKeysCmd = &cobra.Command{
	Use:   "generate-keys",
	Short: "Generate a VAPID key pair for Web Push",
	Args:  cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "webpush:\n  enabled: true\n  publickey: %q\n  privatekey: %q\n", publicKey, privateKey)
		return err
	},
}
