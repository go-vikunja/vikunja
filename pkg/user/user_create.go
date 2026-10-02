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

package user

import (
	"regexp"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm"
)

const (
	IssuerLocal = `local`
	IssuerLDAP  = `ldap`

	// IssuerImport marks a user created by the scheduled user list import that has not logged in
	// yet. The first Entra login replaces it with the real issuer.
	IssuerImport = `import`

	// issuerEntraPrefix is the issuer prefix of Microsoft Entra ID v2 tokens.
	issuerEntraPrefix = `https://login.microsoftonline.com/`
)

// IsEntraIssuer reports whether an OpenID issuer is a Microsoft Entra ID (v2) tenant.
func IsEntraIssuer(issuer string) bool {
	return strings.HasPrefix(issuer, issuerEntraPrefix)
}

// IsImportedEntraIssuer reports whether an issuer belongs to the Entra tenant the scheduled user
// import manages, and whose `oid` claims it may trust. With userimport.tenantid set that is exactly
// that tenant's v2 issuer. Without it, any Entra tenant qualifies, which is only safe with a single
// Entra provider: an object id is unique per tenant, not globally.
func IsImportedEntraIssuer(issuer string) bool {
	tenant := config.UserImportTenantID.GetString()
	if tenant == "" {
		return IsEntraIssuer(issuer)
	}
	return strings.EqualFold(issuer, issuerEntraPrefix+tenant+"/v2.0")
}

type CreateUserOptions struct {
	SkipEmailConfirm bool
}

// CreateUser creates a new user and inserts it into the database
func CreateUser(s *xorm.Session, user *User, options ...CreateUserOptions) (newUser *User, err error) {

	if user.Issuer == "" {
		user.Issuer = IssuerLocal
	}

	// Check if we have all required information
	err = checkIfUserIsValid(user)
	if err != nil {
		return nil, err
	}

	// Check if the user already exists with that username
	err = checkIfUserExists(s, user)
	if err != nil {
		return nil, err
	}

	if user.Issuer == IssuerLocal {
		// Hash the password
		user.Password, err = HashPassword(user.Password)
		if err != nil {
			return nil, err
		}
	}

	user.ID = 0
	user.Status = StatusActive
	user.AvatarProvider = config.DefaultSettingsAvatarProvider.GetString()
	user.AvatarFileID = config.DefaultSettingsAvatarFileID.GetInt64()
	user.EmailRemindersEnabled = config.DefaultSettingsEmailRemindersEnabled.GetBool()
	user.DiscoverableByName = config.DefaultSettingsDiscoverableByName.GetBool()
	user.DiscoverableByEmail = config.DefaultSettingsDiscoverableByEmail.GetBool()
	user.OverdueTasksRemindersEnabled = config.DefaultSettingsOverdueTaskRemindersEnabled.GetBool()
	user.OverdueTasksRemindersTime = config.DefaultSettingsOverdueTaskRemindersTime.GetString()
	user.DefaultProjectID = config.DefaultSettingsDefaultProjectID.GetInt64()
	user.WeekStart = config.DefaultSettingsWeekStart.GetInt()
	user.Timezone = config.DefaultSettingsTimezone.GetString()

	if user.Language == "" {
		user.Language = config.DefaultSettingsLanguage.GetString()
	}

	// Insert it
	_, err = s.Insert(user)
	if err != nil {
		return nil, err
	}

	// Get the  full new User
	newUserOut, err := GetUserByID(s, user.ID)
	if err != nil && !IsErrUserStatusError(err) {
		return nil, err
	}

	events.DispatchOnCommit(s, &CreatedEvent{
		User: newUserOut,
	})

	// Don't send a mail if no mailer is configured
	if !config.MailerEnabled.GetBool() || user.Issuer != IssuerLocal || (len(options) > 0 && options[0].SkipEmailConfirm) {
		return newUserOut, err
	}

	user.Status = StatusEmailConfirmationRequired
	token, err := generateToken(s, user, TokenEmailConfirm)
	if err != nil {
		return nil, err
	}

	confirmationUser := *user
	confirmation := &EmailConfirmNotification{User: &confirmationUser, IsNew: true, ConfirmToken: token.ClearTextToken}
	_, err = s.
		After(func(_ any) {
			// XORM runs this after commit, before the CLI can stop the mail daemon.
			if notifyErr := notifications.Notify(&confirmationUser, confirmation); notifyErr != nil {
				log.Errorf("Failed to queue email confirmation for user %d: %v", confirmationUser.ID, notifyErr)
			}
		}).
		Where("id = ?", user.ID).
		Cols("email", "status").
		Update(user)
	if err != nil {
		return
	}

	// Callers passing a stale status to UpdateUser would silently reactivate the account.
	newUserOut.Status = StatusEmailConfirmationRequired
	return newUserOut, err
}

// CreateBotUser creates a bot user owned by the given owner.
// Bots have no email or password and cannot authenticate interactively.
// It intentionally bypasses checkIfUserIsValid / checkIfUserExists because
// those enforce email+password and would flag duplicate empty emails.
func CreateBotUser(s *xorm.Session, bot *User, owner *User) (*User, error) {
	if owner == nil || owner.ID == 0 {
		return nil, ErrNoUsernamePassword{}
	}
	if owner.IsBot() {
		return nil, &ErrBotNotOwned{UserID: owner.ID}
	}

	// Reuse the same username format rules as regular user creation
	if err := checkUsernameFormat(bot.Username); err != nil {
		return nil, err
	}
	if !hasBotUsernamePrefix(bot.Username) {
		return nil, &ErrBotUsernameMustHavePrefix{Username: bot.Username}
	}

	if _, err := GetUserByUsername(s, bot.Username); err == nil {
		return nil, ErrUsernameExists{Username: bot.Username}
	} else if !IsErrUserDoesNotExist(err) {
		return nil, err
	}

	bot.ID = 0
	bot.BotOwnerID = owner.ID
	bot.Status = StatusActive
	bot.Issuer = IssuerLocal
	bot.Password = ""
	bot.Email = ""

	if _, err := s.Insert(bot); err != nil {
		return nil, err
	}

	newBot, err := GetUserByID(s, bot.ID)
	if err != nil {
		return nil, err
	}

	events.DispatchOnCommit(s, &CreatedEvent{User: newBot})
	return newBot, nil
}

// HashPassword hashes a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), config.ServiceBcryptRounds.GetInt())
	return string(bytes), err
}

// UsernameFromLogin derives a username from an external login name such as a preferred_username
// claim or a userPrincipalName. Spaces become dashes, an email-style value keeps only the part before
// the "@" (a username containing "@" can never be @mentioned because the mention parser stops
// there), and the "#EXT#" marker of Microsoft Entra guest accounts is dropped. An empty result means
// the caller should generate a random username.
func UsernameFromLogin(login string) string {
	name := strings.ReplaceAll(login, " ", "-")
	if at := strings.Index(name, "@"); at >= 0 {
		name = name[:at]
	}
	return strings.TrimSuffix(name, "#EXT#")
}

// botUsernamePrefix is reserved for bot users, which are created via CreateBotUser.
const botUsernamePrefix = "bot-"

var linkSharePattern = regexp.MustCompile(`^link-share-\d+$`)

// isLinkShareUsername reports whether a username looks like the name of a link share.
// The single definition of that rule, used by every username check in this package.
func isLinkShareUsername(username string) bool {
	return linkSharePattern.MatchString(username)
}

// hasBotUsernamePrefix reports whether a username uses the prefix reserved for bots.
func hasBotUsernamePrefix(username string) bool {
	return strings.HasPrefix(username, botUsernamePrefix)
}

// IsReservedUsername reports whether CreateUser refuses a username as reserved: the link share
// pattern and the bot- prefix. It is built from the same predicates CreateUser uses.
func IsReservedUsername(username string) bool {
	return isLinkShareUsername(username) || hasBotUsernamePrefix(username)
}

// checkUsernameFormat validates username format rules shared by regular and bot users.
func checkUsernameFormat(username string) error {
	if username == "" {
		return ErrNoUsernamePassword{}
	}

	if strings.Contains(username, " ") {
		return &ErrUsernameMustNotContainSpaces{
			Username: username,
		}
	}

	// Check if username matches the reserved link-share pattern
	if isLinkShareUsername(username) {
		return ErrUsernameReserved{
			Username: username,
		}
	}

	return nil
}

func checkIfUserIsValid(user *User) error {
	if user.Email == "" ||
		(user.Issuer != IssuerLocal && user.Subject == "") ||
		(user.Issuer == IssuerLocal && (user.Password == "" ||
			user.Username == "")) {
		return ErrNoUsernamePassword{}
	}

	if err := checkUsernameFormat(user.Username); err != nil {
		return err
	}

	// Reserve the bot- prefix for bot users (created via CreateBotUser)
	if hasBotUsernamePrefix(user.Username) {
		return ErrUsernameReserved{
			Username: user.Username,
		}
	}

	return nil
}

func checkIfUserExists(s *xorm.Session, user *User) (err error) {
	exists := true
	_, err = GetUserByUsername(s, user.Username)
	if err != nil {
		switch {
		case IsErrUserDoesNotExist(err):
			exists = false
		case IsErrUserStatusError(err):
			// A disabled or locked user still owns their username. Reporting it as taken lets
			// callers such as CreateUserWithRandomUsername pick another one instead of failing.
		default:
			return err
		}
	}
	if exists {
		return ErrUsernameExists{user.ID, user.Username}
	}

	// Check if the user already existst with that email
	exists = true
	userToCheck := &User{
		Email:   user.Email,
		Issuer:  user.Issuer,
		Subject: user.Subject,
	}

	if user.Issuer != IssuerLocal {
		userToCheck.Email = ""
	}

	_, err = getUser(s, userToCheck, false)
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = false
		} else {
			return err
		}
	}
	if exists && user.Issuer == IssuerLocal {
		return ErrUserEmailExists{user.ID, user.Email}
	}

	return nil
}
