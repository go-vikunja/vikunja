// Copyright 2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"xorm.io/xorm"
)

var (
	userFlagUsername              string
	userFlagEmail                 string
	userFlagPassword              string
	userFlagAvatar                = "default"
	userFlagResetPasswordDirectly bool
	userFlagEnableUser            bool
	userFlagDisableUser           bool
)

func init() {
	// User create flags
	userCreateCmd.Flags().StringVarP(&userFlagUsername, "username", "u", "", "The username of the new user.")
	_ = userCreateCmd.MarkFlagRequired("username")
	userCreateCmd.Flags().StringVarP(&userFlagEmail, "email", "e", "", "The email address of the new user.")
	_ = userCreateCmd.MarkFlagRequired("email")
	userCreateCmd.Flags().StringVarP(&userFlagPassword, "password", "p", "", "The password of the new user. You will be asked to enter it if not provided through the flag.")
	userCreateCmd.Flags().StringVarP(&userFlagAvatar, "avatar-provider", "a", "", "The avatar provider of the new user. Optional.")

	// User update flags
	userUpdateCmd.Flags().StringVarP(&userFlagUsername, "username", "u", "", "The new username of the user.")
	userUpdateCmd.Flags().StringVarP(&userFlagEmail, "email", "e", "", "The new email address of the user.")
	userUpdateCmd.Flags().StringVarP(&userFlagAvatar, "avatar-provider", "a", "", "The new avatar provider of the new user.")

	// Reset PW flags
	userResetPasswordCmd.Flags().BoolVarP(&userFlagResetPasswordDirectly, "direct", "d", false, "If provided, reset the password directly instead of sending the user a reset mail.")
	userResetPasswordCmd.Flags().StringVarP(&userFlagPassword, "password", "p", "", "The new password of the user. Only used in combination with --direct. You will be asked to enter it if not provided through the flag.")

	// Change status flags
	userChangeEnabledCmd.Flags().BoolVarP(&userFlagDisableUser, "disable", "d", false, "Disable the user.")
	userChangeEnabledCmd.Flags().BoolVarP(&userFlagEnableUser, "enable", "e", false, "Enable the user.")

	userCmd.AddCommand(userListCmd, userCreateCmd, userUpdateCmd, userResetPasswordCmd, userChangeEnabledCmd)
	rootCmd.AddCommand(userCmd)
}

func getPasswordFromFlagOrInput() (pw string) {
	pw = userFlagPassword
	if userFlagPassword == "" {
		fmt.Print("Enter Password: ")
		bytePW, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Error reading password: %s", err)
		}
		fmt.Printf("\nConfirm Password: ")
		byteConfirmPW, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Error reading password: %s", err)
		}
		if string(bytePW) != string(byteConfirmPW) {
			log.Critical("Passwords don't match!")
		}
		fmt.Printf("\n")
		pw = strings.TrimSpace(string(bytePW))
	}
	return
}

func getUserFromArg(s *xorm.Session, arg string) *user.User {
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		log.Fatalf("Invalid user id: %s", err)
	}

	u, err := user.GetUserByID(s, id)
	if err != nil {
		log.Fatalf("Could not get user: %s", err)
	}
	return u
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users locally through the cli.",
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all users.",
	PreRun: func(cmd *cobra.Command, args []string) {
		initialize.FullInit()
	},
	Run: func(cmd *cobra.Command, args []string) {
		s := db.NewSession()
		defer s.Close()

		users, err := user.ListUsers(s, "")
		if err != nil {
			_ = s.Rollback()
			log.Fatalf("Error getting users: %s", err)
		}

		if err := s.Commit(); err != nil {
			log.Fatalf("Error getting users: %s", err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"ID",
			"Username",
			"Email",
			"Active",
			"Created",
			"Updated",
		})

		for _, u := range users {
			table.Append([]string{
				strconv.FormatInt(u.ID, 10),
				u.Username,
				u.Email,
				strconv.FormatBool(u.IsActive),
				u.Created.Format(time.RFC3339),
				u.Updated.Format(time.RFC3339),
			})
		}

		table.Render()
	},
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user.",
	PreRun: func(cmd *cobra.Command, args []string) {
		initialize.FullInit()
	},
	Run: func(cmd *cobra.Command, args []string) {
		s := db.NewSession()
		defer s.Close()

		u := &user.User{
			Username: userFlagUsername,
			Email:    userFlagEmail,
			Password: getPasswordFromFlagOrInput(),
		}
		newUser, err := user.CreateUser(s, u)
		if err != nil {
			_ = s.Rollback()
			log.Fatalf("Error creating new user: %s", err)
		}

		err = models.CreateNewNamespaceForUser(s, newUser)
		if err != nil {
			_ = s.Rollback()
			log.Fatalf("Error creating new namespace for user: %s", err)
		}

		if err := s.Commit(); err != nil {
			log.Fatalf("Error saving everything: %s", err)
		}

		fmt.Printf("\nUser was created successfully.\n")
	},
}

var userUpdateCmd = &cobra.Command{
	Use:   "update [user id]",
	Short: "Update an existing user.",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initialize.FullInit()
	},
	Run: func(cmd *cobra.Command, args []string) {
		s := db.NewSession()
		defer s.Close()

		u := getUserFromArg(s, args[0])

		if userFlagUsername != "" {
			u.Username = userFlagUsername
		}
		if userFlagEmail != "" {
			u.Email = userFlagEmail
		}
		if userFlagAvatar != "default" {
			u.AvatarProvider = userFlagAvatar
		}

		_, err := user.UpdateUser(s, u)
		if err != nil {
			_ = s.Rollback()
			log.Fatalf("Error updating the user: %s", err)
		}

		if err := s.Commit(); err != nil {
			log.Fatalf("Error saving everything: %s", err)
		}

		fmt.Println("User updated successfully.")
	},
}

var userResetPasswordCmd = &cobra.Command{
	Use:   "reset-password [user id]",
	Short: "Reset a users password, either through mailing them a reset link or directly.",
	PreRun: func(cmd *cobra.Command, args []string) {
		initialize.FullInit()
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := db.NewSession()
		defer s.Close()

		u := getUserFromArg(s, args[0])

		// By default we reset as usual, only with specific flag directly.
		if userFlagResetPasswordDirectly {
			err := user.UpdateUserPassword(s, u, getPasswordFromFlagOrInput())
			if err != nil {
				_ = s.Rollback()
				log.Fatalf("Could not update user password: %s", err)
			}
			fmt.Println("Password updated successfully.")
		} else {
			err := user.RequestUserPasswordResetToken(s, u)
			if err != nil {
				_ = s.Rollback()
				log.Fatalf("Could not send password reset email: %s", err)
			}
			fmt.Println("Password reset email sent successfully.")
		}

		if err := s.Commit(); err != nil {
			log.Fatalf("Could not send password reset email: %s", err)
		}
	},
}

var userChangeEnabledCmd = &cobra.Command{
	Use:   "change-status [user id]",
	Short: "Enable or disable a user. Will toggle the current status if no flag (--enable or --disable) is provided.",
	PreRun: func(cmd *cobra.Command, args []string) {
		initialize.FullInit()
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := db.NewSession()
		defer s.Close()

		u := getUserFromArg(s, args[0])

		if userFlagEnableUser {
			u.IsActive = true
		} else if userFlagDisableUser {
			u.IsActive = false
		} else {
			u.IsActive = !u.IsActive
		}
		_, err := user.UpdateUser(s, u)
		if err != nil {
			_ = s.Rollback()
			log.Fatalf("Could not enable the user")
		}

		if err := s.Commit(); err != nil {
			log.Fatalf("Error saving everything: %s", err)
		}

		fmt.Printf("User status successfully changed, user is now active: %t.\n", u.IsActive)
	},
}
