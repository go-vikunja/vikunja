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

package models

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"xorm.io/xorm"
)

func SyncExternalTeamsForUser(s *xorm.Session, u *user.User, teams []*Team, issuer, teamNameSuffix string) (err error) {

	if len(teams) == 0 {
		return removeUserFromAllTeamsForThisIssuer(s, u, issuer)
	}

	// Find old teams for user through LDAP
	oldLdapTeams, err := findAllExternalTeamIDsForUser(s, u.ID)
	if err != nil {
		return
	}

	// Assign or create teams for the user
	externalTeamIDs, err := assignOrCreateUserToTeams(s, u, teams, issuer, teamNameSuffix)
	if err != nil {
		return
	}

	// Remove user from teams they're no longer a member of
	teamIDsToLeave := utils.NotIn(oldLdapTeams, externalTeamIDs)
	err = removeUserFromTeamsByIDs(s, u, teamIDsToLeave)

	return
}

// GetTeamByExternalIDAndIssuer returns a team matching the given external_id
// For oidc team creation oidcID and Name need to be set
func GetTeamByExternalIDAndIssuer(s *xorm.Session, oidcID string, issuer string) (*Team, error) {
	team := &Team{}
	has, err := s.
		Table("teams").
		Where("external_id = ? AND issuer = ?", oidcID, issuer).
		Get(team)
	if !has || err != nil {
		return nil, ErrExternalTeamDoesNotExist{issuer, oidcID}
	}
	return team, nil
}

func findAllExternalTeamIDsForUser(s *xorm.Session, userID int64) (ts []int64, err error) {
	err = s.
		Table("team_members").
		Where("user_id = ? ", userID).
		Join("RIGHT", "teams", "teams.id = team_members.team_id").
		Where("teams.external_id != ? AND teams.external_id IS NOT NULL", "").
		Cols("teams.id").
		Find(&ts)
	return
}

func assignOrCreateUserToTeams(s *xorm.Session, u *user.User, teamData []*Team, issuer, teamNameSuffix string) (syncedTeamIDs []int64, err error) {
	if len(teamData) == 0 {
		return
	}

	// Check if we have seen these teams before.
	// Find or create Teams and assign user as teammember.
	teams, err := getOrCreateTeamsByIssuer(s, teamData, u, issuer, teamNameSuffix)
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		tm := &TeamMember{
			TeamID:   team.ID,
			UserID:   u.ID,
			Username: u.Username,
		}
		exists, _ := tm.MembershipExists(s)
		if !exists {
			err = tm.Create(s, u)
			if err != nil {
				log.Errorf("Could not assign user %s to team %s: %v", u.Username, team.Name, err)
			}
		}
		syncedTeamIDs = append(syncedTeamIDs, team.ID)
	}

	return syncedTeamIDs, err
}

func removeUserFromTeamsByIDs(s *xorm.Session, u *user.User, teamIDs []int64) (err error) {
	if len(teamIDs) < 1 {
		return nil
	}

	log.Debugf("Removing team_member with user_id %v from team_ids %v", u.ID, teamIDs)
	_, err = s.
		In("team_id", teamIDs).
		And("user_id = ?", u.ID).
		Delete(&TeamMember{})
	return err
}

func removeUserFromAllTeamsForThisIssuer(s *xorm.Session, u *user.User, issuer string) (err error) {
	teamIDs := []int64{}
	err = s.
		Table("teams").
		Where("issuer = ?", issuer).
		Cols("id").
		Find(&teamIDs)
	if err != nil {
		return
	}

	_, err = s.
		In("team_id", teamIDs).
		And("user_id = ?", u.ID).
		Delete(&TeamMember{})
	return err
}

// getOrCreateTeamsByIssuer returns a slice of teams which were generated from the external provider data.
// If a team did not exist previously it is automatically created.
func getOrCreateTeamsByIssuer(s *xorm.Session, teamData []*Team, u *user.User, issuer, teamNameSuffix string) (teams []*Team, err error) {
	teams = []*Team{}

	for _, externalTeam := range teamData {
		t, err := GetTeamByExternalIDAndIssuer(s, externalTeam.ExternalID, issuer)
		if err != nil && !IsErrExternalTeamDoesNotExist(err) {
			return nil, err
		}

		if err != nil && IsErrExternalTeamDoesNotExist(err) {
			log.Debugf("Team with external ID %s and name %s for issuer %s does not exist. Creating team...", externalTeam.ExternalID, externalTeam.Name, externalTeam.Issuer)
			newTeam, err := createExternalTeam(s, externalTeam, u, issuer, teamNameSuffix)
			if err != nil {
				return teams, err
			}
			teams = append(teams, newTeam)
			continue
		}

		// Compare the name and update if it changed
		if t.Name != getExternalTeamName(externalTeam.Name, teamNameSuffix) {
			t.Name = getExternalTeamName(externalTeam.Name, teamNameSuffix)
		}

		// Compare the description and update if it changed
		if t.Description != externalTeam.Description {
			t.Description = externalTeam.Description
		}

		err = t.Update(s, u)
		if err != nil {
			return nil, err
		}

		log.Debugf("Team with external id %s and name %s for issuer %s already exists.", externalTeam.ExternalID, t.Name, externalTeam.Issuer)
		teams = append(teams, t)
	}

	return teams, err
}

func createExternalTeam(s *xorm.Session, teamData *Team, u *user.User, issuer, teamNameSuffix string) (team *Team, err error) {
	team = &Team{
		Name:        getExternalTeamName(teamData.Name, teamNameSuffix),
		Description: teamData.Description,
		ExternalID:  teamData.ExternalID,
		Issuer:      issuer,
	}
	err = team.CreateNewTeam(s, u, false)
	return team, err
}

func getExternalTeamName(name, suffix string) string {
	return name + " (" + suffix + ")"
}
