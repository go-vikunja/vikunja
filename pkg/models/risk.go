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
	"strings"
	"time"
	"unicode/utf8"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// The statuses of a risk. A risk can move between any two of them at any time: closing marks it
// resolved, and it can be reopened later.
const (
	RiskStatusOpen       = "open"
	RiskStatusMitigating = "mitigating"
	RiskStatusAccepted   = "accepted"
	RiskStatusClosed     = "closed"
)

var riskStatuses = []string{RiskStatusOpen, RiskStatusMitigating, RiskStatusAccepted, RiskStatusClosed}

// IsValidRiskStatus reports whether the status is one of the known ones.
func IsValidRiskStatus(status string) bool {
	for _, s := range riskStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// The rating of a risk, derived from its score (probability times impact).
const (
	RiskRatingLow      = "low"
	RiskRatingMedium   = "medium"
	RiskRatingHigh     = "high"
	RiskRatingCritical = "critical"
)

// riskRatingBands are the inclusive score ranges of the ratings. Keep them in sync with
// frontend/src/helpers/riskRating.ts, both have a test that pins the same table.
var riskRatingBands = []struct {
	Rating string
	Min    int
	Max    int
}{
	{RiskRatingLow, 1, 4},
	{RiskRatingMedium, 5, 9},
	{RiskRatingHigh, 10, 16},
	{RiskRatingCritical, 17, 25},
}

// RiskScore is probability times impact.
func RiskScore(probability, impact int) int {
	return probability * impact
}

// RiskRatingForScore maps a score to its rating. A score outside 1-25 has none.
func RiskRatingForScore(score int) string {
	for _, band := range riskRatingBands {
		if score >= band.Min && score <= band.Max {
			return band.Rating
		}
	}
	return ""
}

func riskRatingRange(rating string) (lo, hi int, ok bool) {
	for _, band := range riskRatingBands {
		if band.Rating == rating {
			return band.Min, band.Max, true
		}
	}
	return 0, 0, false
}

const (
	maxRiskTitleLength    = 250
	maxRiskCategoryLength = 100
	maxRiskTextLength     = 20000
	minRiskLevel          = 1
	maxRiskLevel          = 5
	defaultRiskLevel      = 3
)

// Risk is an entry of the risk register of a project.
//
// v2-only: doc: tags are the schema's source of truth. It implements CRUDable and Permissions because
// the shared handler.Do* pipeline needs them. Permissions follow the project: read access to the
// project reads its risks, write access creates, edits, changes the status of and deletes them.
// The status is not editable through Update, only through ChangeRiskStatus, so the history of
// status changes cannot be bypassed.
type Risk struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"risk" readOnly:"true" doc:"The unique, numeric id of this risk."`

	ProjectID int64 `xorm:"bigint not null index" json:"project_id" param:"project" readOnly:"true" doc:"The project this risk belongs to. Set from the URL when it is created and fixed afterwards."`

	Title       string `xorm:"varchar(250) not null" json:"title" minLength:"1" maxLength:"250" doc:"The title of the risk."`
	Description string `xorm:"text null" json:"description" doc:"What the risk is and how it could affect the project."`
	Category    string `xorm:"varchar(100) null" json:"category" maxLength:"100" doc:"A free text category, for example Schedule, Budget or Technical."`

	Probability int `xorm:"int not null default 3" json:"probability" minimum:"1" maximum:"5" doc:"How likely the risk is, from 1 (rare) to 5 (almost certain). Defaults to 3 when a risk is created."`
	Impact      int `xorm:"int not null default 3" json:"impact" minimum:"1" maximum:"5" doc:"How bad it would be, from 1 (negligible) to 5 (severe). Defaults to 3 when a risk is created."`

	Score  int    `xorm:"-" json:"score" readOnly:"true" doc:"Probability times impact, from 1 to 25. Computed by the server."`
	Rating string `xorm:"-" json:"rating" readOnly:"true" enum:"low,medium,high,critical" doc:"The rating of the score: low 1-4, medium 5-9, high 10-16, critical 17-25. Computed by the server."`

	OwnerID int64      `xorm:"bigint null index" json:"owner_id" minimum:"0" doc:"The id of the user who is responsible for the risk, 0 for none. The owner must be an active user who can read the project."`
	Owner   *user.User `xorm:"-" json:"owner" readOnly:"true" doc:"The owner of the risk, if there is one."`

	Mitigation  string `xorm:"text null" json:"mitigation" doc:"What is done to lower the probability or the impact."`
	Contingency string `xorm:"text null" json:"contingency" doc:"What is done if the risk happens."`

	Status string `xorm:"varchar(20) not null default 'open' index" json:"status" readOnly:"true" enum:"open,mitigating,accepted,closed" doc:"The status of the risk. Changed only through the status operation, which keeps a history. Ignored when sent while creating or editing."`

	IdentifiedDate *time.Time `xorm:"DATETIME null" json:"identified_date" nullable:"true" doc:"When the risk was identified."`
	DueDate        *time.Time `xorm:"DATETIME null index" json:"due_date" nullable:"true" doc:"When the risk has to be reviewed or resolved. A risk that is not closed after this date counts as overdue."`

	ClosedAt   *time.Time `xorm:"DATETIME null" json:"closed_at" nullable:"true" readOnly:"true" doc:"When the risk was closed. Empty while it is not closed, cleared again when it is reopened."`
	ClosedByID int64      `xorm:"bigint null" json:"-"`
	ClosedBy   *user.User `xorm:"-" json:"closed_by" readOnly:"true" doc:"Who closed the risk."`
	Resolution string     `xorm:"text null" json:"resolution" readOnly:"true" doc:"The note given when the risk was closed. Cleared when it is reopened, the history keeps it."`

	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"Who created the risk."`

	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this risk was created. You cannot change this value."`
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this risk was last updated. You cannot change this value."`

	// Filter-only fields (not persisted): set by the v2 list routes, read by ReadAll.
	ProjectIDs      []int64  `xorm:"-" json:"-"`
	Statuses        []string `xorm:"-" json:"-"`
	Ratings         []string `xorm:"-" json:"-"`
	OwnerIDs        []int64  `xorm:"-" json:"-"`
	Categories      []string `xorm:"-" json:"-"`
	Overdue         bool     `xorm:"-" json:"-"`
	IncludeArchived bool     `xorm:"-" json:"-"`
	SortBy          []string `xorm:"-" json:"-"`
	OrderBy         []string `xorm:"-" json:"-"`

	// Filled by the permission checks so that the operation does not read the same rows again.
	loaded  *Risk
	project *Project

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName returns the table name of Risk.
func (*Risk) TableName() string {
	return "risks"
}

// RiskStatusHistory is one change of the status of a risk. The first row of a risk is its creation,
// with an empty from_status.
type RiskStatusHistory struct {
	ID          int64      `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true" doc:"The unique id of this history entry."`
	RiskID      int64      `xorm:"bigint not null index" json:"risk_id" readOnly:"true" doc:"The risk this entry belongs to."`
	FromStatus  string     `xorm:"varchar(20) null" json:"from_status" readOnly:"true" doc:"The status before the change, empty for the creation of the risk."`
	ToStatus    string     `xorm:"varchar(20) not null" json:"to_status" readOnly:"true" doc:"The status after the change."`
	Note        string     `xorm:"text null" json:"note" readOnly:"true" doc:"The note given with the change, for example why the risk was closed."`
	ChangedByID int64      `xorm:"bigint not null" json:"-"`
	ChangedBy   *user.User `xorm:"-" json:"changed_by" readOnly:"true" doc:"Who made the change. Empty if the user does not exist any more."`
	Created     time.Time  `xorm:"created not null" json:"created" readOnly:"true" doc:"When the change was made."`
}

// TableName returns the table name of RiskStatusHistory.
func (*RiskStatusHistory) TableName() string {
	return "risk_status_history"
}

func getRiskByID(s *xorm.Session, id int64) (*Risk, error) {
	if id < 1 {
		return nil, ErrRiskDoesNotExist{ID: id}
	}
	risk := &Risk{}
	has, err := s.Where("id = ?", id).Get(risk)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrRiskDoesNotExist{ID: id}
	}
	return risk, nil
}

// fillRisks adds the computed score and rating and the users of the risks, with one query for all of
// them. A user that does not exist any more is left empty.
func fillRisks(s *xorm.Session, risks []*Risk) error {
	ids := make([]int64, 0, len(risks)*3)
	seen := make(map[int64]bool)
	add := func(id int64) {
		if id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, r := range risks {
		r.Score = RiskScore(r.Probability, r.Impact)
		r.Rating = RiskRatingForScore(r.Score)
		add(r.OwnerID)
		add(r.CreatedByID)
		add(r.ClosedByID)
	}
	if len(ids) == 0 {
		return nil
	}

	users, err := getUsersOrLinkSharesFromIDs(s, ids)
	if err != nil {
		return err
	}
	for _, r := range risks {
		r.Owner = users[r.OwnerID]
		r.CreatedBy = users[r.CreatedByID]
		r.ClosedBy = users[r.ClosedByID]
	}
	return nil
}

func trimRiskText(r *Risk) {
	r.Title = strings.TrimSpace(r.Title)
	r.Category = strings.TrimSpace(r.Category)
}

// validate checks the fields a user can set.
func (r *Risk) validate() error {
	if r.Title == "" {
		return ErrInvalidData{Message: "The title of a risk must not be empty."}
	}
	if utf8.RuneCountInString(r.Title) > maxRiskTitleLength {
		return ErrInvalidData{Message: "The title of a risk must not be longer than 250 characters."}
	}
	if utf8.RuneCountInString(r.Category) > maxRiskCategoryLength {
		return ErrInvalidData{Message: "The category of a risk must not be longer than 100 characters."}
	}
	for _, text := range []string{r.Description, r.Mitigation, r.Contingency} {
		if utf8.RuneCountInString(text) > maxRiskTextLength {
			return ErrInvalidData{Message: "A text field of a risk must not be longer than 20000 characters."}
		}
	}
	if r.Probability < minRiskLevel || r.Probability > maxRiskLevel {
		return ErrInvalidData{Message: "The probability must be between 1 and 5."}
	}
	if r.Impact < minRiskLevel || r.Impact > maxRiskLevel {
		return ErrInvalidData{Message: "The impact must be between 1 and 5."}
	}
	if r.OwnerID < 0 {
		return ErrInvalidData{Message: "The owner is not valid."}
	}
	if r.IdentifiedDate != nil && r.DueDate != nil && r.DueDate.Before(*r.IdentifiedDate) {
		return ErrInvalidData{Message: "The due date must not be before the date the risk was identified."}
	}
	return nil
}

// validateRiskOwner checks that the person can own a risk of the project: an active user, not a bot,
// who can read the project. 0 means no owner.
func validateRiskOwner(s *xorm.Session, ownerID int64, project *Project) error {
	if ownerID == 0 {
		return nil
	}

	owner, err := user.GetUserByID(s, ownerID)
	if err != nil {
		if user.IsErrUserDoesNotExist(err) || user.IsErrUserStatusError(err) {
			return ErrRiskOwnerHasNoAccess{UserID: ownerID}
		}
		return err
	}
	if owner.IsBot() || owner.Status != user.StatusActive {
		return ErrRiskOwnerHasNoAccess{UserID: ownerID}
	}

	can, _, err := project.CanRead(s, owner)
	if err != nil {
		return err
	}
	if !can {
		return ErrRiskOwnerHasNoAccess{UserID: ownerID}
	}
	return nil
}

// ---- CRUDable -------------------------------------------------------------------------------------

// Create adds a risk to the project in r.ProjectID. The status starts as open.
func (r *Risk) Create(s *xorm.Session, a web.Auth) error {
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	project := r.project
	if project == nil {
		project, err = GetProjectSimpleByID(s, r.ProjectID)
		if err != nil {
			return err
		}
	}

	trimRiskText(r)
	if r.Probability == 0 {
		r.Probability = defaultRiskLevel
	}
	if r.Impact == 0 {
		r.Impact = defaultRiskLevel
	}
	err = r.validate()
	if err != nil {
		return err
	}
	err = validateRiskOwner(s, r.OwnerID, project)
	if err != nil {
		return err
	}

	// The status and everything that goes with closing are not ours to take from the request.
	r.ID = 0
	r.Status = RiskStatusOpen
	r.ClosedAt = nil
	r.ClosedByID = 0
	r.Resolution = ""
	r.CreatedByID = doer.ID

	_, err = s.Insert(r)
	if err != nil {
		return err
	}
	_, err = s.Insert(&RiskStatusHistory{RiskID: r.ID, ToStatus: RiskStatusOpen, ChangedByID: doer.ID})
	if err != nil {
		return err
	}

	err = fillRisks(s, []*Risk{r})
	if err != nil {
		return err
	}
	events.DispatchOnCommit(s, &RiskCreatedEvent{Risk: r, Doer: doer})
	return nil
}

// ReadOne has nothing left to do: CanRead loaded the risk.
func (r *Risk) ReadOne(s *xorm.Session, _ web.Auth) error {
	return fillRisks(s, []*Risk{r})
}

// Update changes the fields a user can edit. The project, the status and what goes with closing stay
// as they are, whatever was sent.
func (r *Risk) Update(s *xorm.Session, a web.Auth) error {
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	existing := r.loaded
	if existing == nil {
		existing, err = getRiskByID(s, r.ID)
		if err != nil {
			return err
		}
	}
	project := r.project
	if project == nil {
		project, err = GetProjectSimpleByID(s, existing.ProjectID)
		if err != nil {
			return err
		}
	}

	incoming := &Risk{
		Title:          r.Title,
		Description:    r.Description,
		Category:       r.Category,
		Probability:    r.Probability,
		Impact:         r.Impact,
		OwnerID:        r.OwnerID,
		Mitigation:     r.Mitigation,
		Contingency:    r.Contingency,
		IdentifiedDate: r.IdentifiedDate,
		DueDate:        r.DueDate,
	}
	trimRiskText(incoming)
	err = incoming.validate()
	if err != nil {
		return err
	}
	// An owner who lost access since is kept; only a new owner has to qualify.
	if incoming.OwnerID != existing.OwnerID {
		err = validateRiskOwner(s, incoming.OwnerID, project)
		if err != nil {
			return err
		}
	}

	_, err = s.
		ID(existing.ID).
		Cols("title", "description", "category", "probability", "impact", "owner_id", "mitigation", "contingency", "identified_date", "due_date").
		Update(incoming)
	if err != nil {
		return err
	}

	fresh, err := getRiskByID(s, existing.ID)
	if err != nil {
		return err
	}
	err = fillRisks(s, []*Risk{fresh})
	if err != nil {
		return err
	}
	loaded, loadedProject := r.loaded, r.project
	*r = *fresh
	r.loaded, r.project = loaded, loadedProject

	events.DispatchOnCommit(s, &RiskUpdatedEvent{Risk: r, Doer: doer})
	return nil
}

// Delete removes the risk and its history.
func (r *Risk) Delete(s *xorm.Session, a web.Auth) error {
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	_, err = s.Where("risk_id = ?", r.ID).Delete(&RiskStatusHistory{})
	if err != nil {
		return err
	}
	_, err = s.Where("id = ?", r.ID).Delete(&Risk{})
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &RiskDeletedEvent{Risk: r, Doer: doer})
	return nil
}

// ---- listing --------------------------------------------------------------------------------------

// riskSortColumns are the only things a list can be sorted by. They are SQL fragments chosen here,
// request text never reaches the query.
var riskSortColumns = map[string]string{
	"score":       "(risks.probability * risks.impact)",
	"probability": "risks.probability",
	"impact":      "risks.impact",
	"due_date":    "risks.due_date",
	"status":      "risks.status",
	"title":       "risks.title",
	"created":     "risks.created",
	"updated":     "risks.updated",
	"id":          "risks.id",
}

// IsValidRiskSortKey reports whether a list can be sorted by the key.
func IsValidRiskSortKey(key string) bool {
	_, ok := riskSortColumns[key]
	return ok
}

func (r *Risk) orderBy() (string, error) {
	if len(r.SortBy) == 0 {
		// Worst first, the order a register is read in.
		return "(risks.probability * risks.impact) DESC, risks.id ASC", nil
	}

	parts := make([]string, 0, len(r.SortBy)+1)
	hasID := false
	for i, key := range r.SortBy {
		column, ok := riskSortColumns[key]
		if !ok {
			return "", ErrInvalidData{Message: "Risks cannot be sorted by " + key + "."}
		}
		direction := "ASC"
		if i < len(r.OrderBy) {
			switch strings.ToLower(r.OrderBy[i]) {
			case "asc", "":
			case "desc":
				direction = "DESC"
			default:
				return "", ErrInvalidData{Message: "The sort order must be asc or desc."}
			}
		}
		if key == "id" {
			hasID = true
		}
		parts = append(parts, column+" "+direction)
	}
	if !hasID {
		parts = append(parts, "risks.id ASC")
	}
	return strings.Join(parts, ", "), nil
}

// listCond builds the condition of a list. Access always comes first: every filter only narrows it.
func (r *Risk) listCond(s *xorm.Session, a web.Auth) (builder.Cond, error) {
	cond, err := accessibleProjectIDsCond(s, a, "risks.project_id")
	if err != nil {
		return nil, err
	}
	conds := []builder.Cond{cond}

	if r.ProjectID > 0 {
		conds = append(conds, builder.Eq{"risks.project_id": r.ProjectID})
	}
	if len(r.ProjectIDs) > 0 {
		conds = append(conds, builder.In("risks.project_id", r.ProjectIDs))
	}

	// Archived projects are left out of a list over many projects, unless asked for. A list of one
	// project, or of projects that were chosen, shows what was asked for.
	if !r.IncludeArchived && r.ProjectID == 0 && len(r.ProjectIDs) == 0 {
		conds = append(conds, builder.Expr(
			"NOT EXISTS (SELECT 1 FROM project_ancestors pa INNER JOIN projects ap ON ap.id = pa.ancestor_id WHERE pa.project_id = risks.project_id AND ap.is_archived = ?)",
			true,
		))
	}

	if len(r.Statuses) > 0 {
		for _, status := range r.Statuses {
			if !IsValidRiskStatus(status) {
				return nil, ErrInvalidRiskStatus{Status: status}
			}
		}
		conds = append(conds, builder.In("risks.status", r.Statuses))
	}

	if len(r.Ratings) > 0 {
		ratingConds := make([]builder.Cond, 0, len(r.Ratings))
		for _, rating := range r.Ratings {
			lo, hi, ok := riskRatingRange(rating)
			if !ok {
				return nil, ErrInvalidData{Message: "The rating must be one of low, medium, high or critical."}
			}
			ratingConds = append(ratingConds, builder.Expr("(risks.probability * risks.impact) BETWEEN ? AND ?", lo, hi))
		}
		conds = append(conds, builder.Or(ratingConds...))
	}

	if len(r.OwnerIDs) > 0 {
		ownerConds := make([]builder.Cond, 0, len(r.OwnerIDs))
		var ids []int64
		for _, id := range r.OwnerIDs {
			if id == 0 {
				// "No owner" is stored as 0, and as NULL on rows that never had one.
				ownerConds = append(ownerConds, builder.Eq{"risks.owner_id": 0}, builder.IsNull{"risks.owner_id"})
				continue
			}
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			ownerConds = append(ownerConds, builder.In("risks.owner_id", ids))
		}
		conds = append(conds, builder.Or(ownerConds...))
	}

	if len(r.Categories) > 0 {
		categoryConds := make([]builder.Cond, 0, len(r.Categories))
		for _, category := range r.Categories {
			categoryConds = append(categoryConds, builder.Expr("lower(risks.category) = ?", strings.ToLower(strings.TrimSpace(category))))
		}
		conds = append(conds, builder.Or(categoryConds...))
	}

	if r.Overdue {
		conds = append(conds,
			builder.NotNull{"risks.due_date"},
			builder.Lt{"risks.due_date": time.Now().UTC()},
			builder.Neq{"risks.status": RiskStatusClosed},
		)
	}

	return builder.And(conds...), nil
}

func riskSearchCond(search string) builder.Cond {
	// Plain substring matching: the full text search of some databases needs an index this table
	// does not have.
	fields := []string{"risks.title", "risks.description", "risks.category", "risks.mitigation", "risks.contingency"}
	conds := make([]builder.Cond, 0, len(fields))
	for _, field := range fields {
		conds = append(conds, db.ILIKE(field, search))
	}
	return builder.Or(conds...)
}

// ReadAll lists the risks of every project the user can read, or of the project in r.ProjectID,
// narrowed by the filters. Link shares see none.
func (r *Risk) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result any, resultCount int, numberOfTotalItems int64, err error) {
	// DoReadAll skips the permission check, so the guard lives here: a link share is not a person.
	if _, isShare := a.(*LinkSharing); isShare {
		return []*Risk{}, 0, 0, nil
	}

	cond, err := r.listCond(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if strings.TrimSpace(search) != "" {
		cond = builder.And(cond, riskSearchCond(strings.TrimSpace(search)))
	}
	order, err := r.orderBy()
	if err != nil {
		return nil, 0, 0, err
	}

	total, err := s.Where(cond).Count(&Risk{})
	if err != nil {
		return nil, 0, 0, err
	}

	risks := []*Risk{}
	err = s.
		Where(cond).
		OrderBy(order).
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&risks)
	if err != nil {
		return nil, 0, 0, err
	}

	err = fillRisks(s, risks)
	if err != nil {
		return nil, 0, 0, err
	}
	return risks, len(risks), total, nil
}

// ---- status ---------------------------------------------------------------------------------------

// ChangeRiskStatus moves a risk to another status, from and to any, and records it in the history.
// Closing sets when, by whom and with which note; moving away from closed (reopening) clears all
// three. The same status again changes nothing. It needs write access to the project. It does not
// commit; the caller owns the transaction and dispatches the queued events.
//
// The update only applies while the risk still has the status it was read with, so two people
// changing the status at the same moment cannot both write a history entry.
func ChangeRiskStatus(s *xorm.Session, a web.Auth, id int64, status, note string) (*Risk, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, ErrGenericForbidden{}
	}
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}
	if !IsValidRiskStatus(status) {
		return nil, ErrInvalidRiskStatus{Status: status}
	}
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) > maxRiskTextLength {
		return nil, ErrInvalidData{Message: "The note must not be longer than 20000 characters."}
	}

	existing, err := getRiskByID(s, id)
	if err != nil {
		return nil, err
	}
	project, err := GetProjectSimpleByID(s, existing.ProjectID)
	if err != nil {
		return nil, err
	}
	can, err := project.CanWrite(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}

	if existing.Status == status {
		err = fillRisks(s, []*Risk{existing})
		return existing, err
	}

	update := &Risk{Status: status}
	if status == RiskStatusClosed {
		now := time.Now()
		update.ClosedAt = &now
		update.ClosedByID = doer.ID
		update.Resolution = note
	}
	affected, err := s.
		Where("id = ? AND status = ?", id, existing.Status).
		Cols("status", "closed_at", "closed_by_id", "resolution").
		Update(update)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrRiskStatusConflict{ID: id}
	}

	_, err = s.Insert(&RiskStatusHistory{
		RiskID:      id,
		FromStatus:  existing.Status,
		ToStatus:    status,
		Note:        note,
		ChangedByID: doer.ID,
	})
	if err != nil {
		return nil, err
	}

	fresh, err := getRiskByID(s, id)
	if err != nil {
		return nil, err
	}
	err = fillRisks(s, []*Risk{fresh})
	if err != nil {
		return nil, err
	}
	events.DispatchOnCommit(s, &RiskStatusChangedEvent{Risk: fresh, Doer: doer, FromStatus: existing.Status, ToStatus: status})
	return fresh, nil
}

// ListRiskHistory returns the status changes of a risk, newest first. It needs read access.
func ListRiskHistory(s *xorm.Session, a web.Auth, id int64, page, perPage int) ([]*RiskStatusHistory, int64, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, 0, ErrGenericForbidden{}
	}

	risk := &Risk{ID: id}
	can, _, err := risk.CanRead(s, a)
	if err != nil {
		return nil, 0, err
	}
	if !can {
		return nil, 0, ErrGenericForbidden{}
	}

	total, err := s.Where("risk_id = ?", id).Count(&RiskStatusHistory{})
	if err != nil {
		return nil, 0, err
	}

	entries := []*RiskStatusHistory{}
	err = s.
		Where("risk_id = ?", id).
		OrderBy("created DESC, id DESC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&entries)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]int64, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ChangedByID)
	}
	users, err := getUsersOrLinkSharesFromIDs(s, ids)
	if err != nil {
		return nil, 0, err
	}
	for _, entry := range entries {
		entry.ChangedBy = users[entry.ChangedByID]
	}
	return entries, total, nil
}

// ---- cleanup --------------------------------------------------------------------------------------

// deleteRisksOfProject removes every risk of a project with its history, when the project is deleted.
func deleteRisksOfProject(s *xorm.Session, projectID int64) error {
	ids := builder.Select("id").From("risks").Where(builder.Eq{"project_id": projectID})
	_, err := s.Where(builder.In("risk_id", ids)).Delete(&RiskStatusHistory{})
	if err != nil {
		return err
	}
	_, err = s.Where("project_id = ?", projectID).Delete(&Risk{})
	return err
}

// clearRiskOwner makes the risks a user owned unowned, when the user is deleted. Who created or
// changed things stays as the id, and shows as a user that does not exist any more.
func clearRiskOwner(s *xorm.Session, userID int64) error {
	_, err := s.
		Where("owner_id = ?", userID).
		Cols("owner_id").
		Update(&Risk{OwnerID: 0})
	return err
}
