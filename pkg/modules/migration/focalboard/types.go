// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package focalboard

import (
	"encoding/json"
	"time"
)

const (
	NormalizedSchemaVersion    = "1"
	AssigneeMapSchemaVersion   = "1"
	ConverterVersion           = "1.0.0"
	DefaultVikunjaVersion      = "2.5.0"
	SupportedFocalboardVersion = 2
	KnownSourceSHA256          = "c568c140fa6d4cf3ebcd40de64d0b98eb4630689c2dab39537bd825adf516f3c"
	RootProjectTitle           = "НТИ — импорт Focalboard 2026-08-03"
)

type DescriptionLinkMethod string

const (
	DescriptionDirect    DescriptionLinkMethod = "direct"
	DescriptionRecovered DescriptionLinkMethod = "recovered_by_timestamp"
	DescriptionNone      DescriptionLinkMethod = "none"
)

type Options struct {
	ExpectedSHA256 string
	Strict         bool
	Timezone       string
	VikunjaVersion string
	RepoSHA        string
	RepoDirty      bool
	Assignees      *AssigneeMap
}

type AssigneeMap struct {
	SchemaVersion string            `json:"schema_version" yaml:"schema_version"`
	Mappings      []AssigneeMapping `json:"mappings" yaml:"mappings"`
}

type AssigneeMapping struct {
	SourceRaw             string `json:"source_raw" yaml:"source_raw"`
	TargetUsernameOrEmail string `json:"target_username_or_email" yaml:"target_username_or_email"`
	Unassigned            bool   `json:"unassigned" yaml:"unassigned"`
}

type NormalizedArchive struct {
	SchemaVersion       string            `json:"schema_version"`
	SourceSystem        string            `json:"source_system"`
	SourceArchiveSHA256 string            `json:"source_archive_sha256"`
	SourceNestedSHA256  string            `json:"source_boardarchive_sha256"`
	SourceVersion       int               `json:"source_focalboard_version"`
	MigrationRun        MigrationRun      `json:"migration_run"`
	Counts              Counts            `json:"counts"`
	Boards              []NormalizedBoard `json:"boards"`
}

type MigrationRun struct {
	RunID          string `json:"run_id"`
	ToolVersion    string `json:"tool_version"`
	Strict         bool   `json:"strict"`
	Timezone       string `json:"timezone,omitempty"`
	VikunjaVersion string `json:"vikunja_version"`
}

type NormalizedBoard struct {
	SourceBoardID   string           `json:"source_board_id"`
	SourceViewIDs   []string         `json:"source_view_ids"`
	StatusOptions   []string         `json:"source_status_options"`
	Title           string           `json:"title"`
	SourceCreatedAt string           `json:"source_created_at"`
	SourceUpdatedAt string           `json:"source_updated_at"`
	Tasks           []NormalizedTask `json:"tasks"`
}

type NormalizedTask struct {
	SourceSystem          string                `json:"source_system"`
	SourceArchiveSHA256   string                `json:"source_archive_sha256"`
	SourceBoardID         string                `json:"source_board_id"`
	SourceCardID          string                `json:"source_card_id"`
	SourceTextID          string                `json:"source_text_id,omitempty"`
	Title                 string                `json:"title"`
	Description           string                `json:"description"`
	DescriptionLinkMethod DescriptionLinkMethod `json:"description_link_method"`
	StatusRaw             string                `json:"status_raw"`
	SourceProperties      map[string]string     `json:"source_properties"`
	AssigneeRaw           string                `json:"assignee_raw"`
	AssigneeTarget        string                `json:"assignee_target,omitempty"`
	AssigneeUnassigned    bool                  `json:"assignee_unassigned"`
	PriorityRaw           string                `json:"priority_raw"`
	NativePriority        int64                 `json:"native_priority"`
	DueRaw                string                `json:"due_raw"`
	ParsedDueCandidate    string                `json:"parsed_due_candidate,omitempty"`
	NativeDueDate         string                `json:"native_due_date,omitempty"`
	SourceCreatedAt       string                `json:"source_created_at"`
	SourceUpdatedAt       string                `json:"source_updated_at"`
}

type Counts struct {
	Boards                  int            `json:"boards"`
	Cards                   int            `json:"cards"`
	TextBlocks              int            `json:"text_blocks"`
	KanbanViews             int            `json:"kanban_views"`
	Attachments             int            `json:"attachments"`
	DeletedCards            int            `json:"deleted_cards"`
	Templates               int            `json:"templates"`
	EmptyTitles             int            `json:"empty_titles"`
	Statuses                map[string]int `json:"statuses"`
	NonEmptyPriorities      int            `json:"non_empty_priorities"`
	NonEmptyAssignees       int            `json:"non_empty_assignees"`
	UniqueAssigneeRawValues int            `json:"unique_assignee_raw_values"`
	NonEmptyDueRawValues    int            `json:"non_empty_due_raw_values"`
	StrictISODueValues      int            `json:"strict_iso_due_values"`
	FreeTextDueValues       int            `json:"free_text_due_values"`
	EmptyDueValues          int            `json:"empty_due_values"`
	DirectDescriptions      int            `json:"direct_descriptions"`
	RecoveredDescriptions   int            `json:"recovered_descriptions"`
	CardsWithoutDescription int            `json:"cards_without_description"`
	FinalDescriptions       int            `json:"final_descriptions"`
	DuplicateTitleExtras    int            `json:"duplicate_title_extras"`
}

type Analysis struct {
	SchemaVersion       string    `json:"schema_version"`
	SourceArchiveSHA256 string    `json:"source_archive_sha256"`
	SourceNestedSHA256  string    `json:"source_boardarchive_sha256"`
	SourceVersion       int       `json:"source_focalboard_version"`
	Counts              Counts    `json:"counts"`
	Warnings            []Warning `json:"warnings"`
}

type Warning struct {
	Code          string `json:"code"`
	SourceBoardID string `json:"source_board_id,omitempty"`
	SourceCardID  string `json:"source_card_id,omitempty"`
	Detail        string `json:"detail,omitempty"`
}

type RecoveredDescription struct {
	SourceBoardID string `json:"source_board_id"`
	SourceCardID  string `json:"source_card_id"`
	SourceTextID  string `json:"source_text_id"`
	DeltaMS       int64  `json:"delta_ms"`
}

type CardMapping struct {
	SourceBoardID         string                `json:"source_board_id"`
	SourceCardID          string                `json:"source_card_id"`
	SourceTextID          string                `json:"source_text_id,omitempty"`
	SourceCreatedAt       string                `json:"source_created_at"`
	SourceUpdatedAt       string                `json:"source_updated_at"`
	TargetProjectLegacyID int64                 `json:"target_project_legacy_id"`
	TargetTaskLegacyID    int64                 `json:"target_task_legacy_id"`
	DescriptionLinkMethod DescriptionLinkMethod `json:"description_link_method"`
	NativePriority        int64                 `json:"native_priority"`
	NativeDueDateSet      bool                  `json:"native_due_date_set"`
	AssigneeMapped        bool                  `json:"assignee_mapped"`
}

type Reconciliation struct {
	SchemaVersion         string                 `json:"schema_version"`
	RunID                 string                 `json:"run_id"`
	SourceArchiveSHA256   string                 `json:"source_archive_sha256"`
	Counts                Counts                 `json:"counts"`
	RecoveredDescriptions []RecoveredDescription `json:"recovered_descriptions"`
	CardMappings          []CardMapping          `json:"card_mappings"`
	UnknownAssignees      []string               `json:"unknown_assignees"`
	Warnings              []Warning              `json:"warnings"`
	Verified              bool                   `json:"verified"`
}

type RunManifest struct {
	SchemaVersion       string            `json:"schema_version"`
	RunID               string            `json:"run_id"`
	RepoSHA             string            `json:"repo_sha,omitempty"`
	RepoDirty           bool              `json:"repo_dirty"`
	ToolVersion         string            `json:"tool_version"`
	GeneratedAt         string            `json:"generated_at"`
	ConfigSHA256        string            `json:"config_sha256"`
	SourceArchiveSHA256 string            `json:"source_archive_sha256"`
	SourceNestedSHA256  string            `json:"source_boardarchive_sha256"`
	SourceVersion       int               `json:"source_focalboard_version"`
	SourceMaxUpdatedAt  string            `json:"source_max_updated_at"`
	Config              ManifestConfig    `json:"config"`
	Counts              Counts            `json:"counts"`
	ArtifactSHA256      map[string]string `json:"artifact_sha256"`
}

type ManifestConfig struct {
	Strict         bool   `json:"strict"`
	Timezone       string `json:"timezone,omitempty"`
	VikunjaVersion string `json:"vikunja_version"`
}

type Result struct {
	Normalized         *NormalizedArchive
	Reconciliation     *Reconciliation
	Manifest           *RunManifest
	NormalizedJSON     []byte
	VikunjaZip         []byte
	ReconciliationJSON []byte
	ReconciliationCSV  []byte
	AssigneeTemplate   []byte
	ManifestJSON       []byte
	README             []byte
}

type rawPropertyOption struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	Color string `json:"color"`
}

type rawCardProperty struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Type    string              `json:"type"`
	Options []rawPropertyOption `json:"options"`
}

type rawBoard struct {
	ID              string            `json:"id"`
	Title           string            `json:"title"`
	Type            string            `json:"type"`
	CreateAt        int64             `json:"createAt"`
	UpdateAt        int64             `json:"updateAt"`
	DeleteAt        int64             `json:"deleteAt"`
	CreatedBy       string            `json:"createdBy"`
	ModifiedBy      string            `json:"modifiedBy"`
	ChannelID       string            `json:"channelId"`
	TeamID          string            `json:"teamId"`
	Description     string            `json:"description"`
	Icon            string            `json:"icon"`
	ShowDescription bool              `json:"showDescription"`
	IsTemplate      bool              `json:"isTemplate"`
	TemplateVersion int64             `json:"templateVersion"`
	MinimumRole     string            `json:"minimumRole"`
	Properties      json.RawMessage   `json:"properties"`
	CardProperties  []rawCardProperty `json:"cardProperties"`
}

type rawBlockFields struct {
	CardOrder          []string                   `json:"cardOrder"`
	CollapsedOptionIDs []string                   `json:"collapsedOptionIds"`
	ColumnCalculations map[string]json.RawMessage `json:"columnCalculations"`
	ColumnWidths       map[string]json.RawMessage `json:"columnWidths"`
	ContentOrder       []string                   `json:"contentOrder"`
	DefaultTemplateID  string                     `json:"defaultTemplateId"`
	Filter             map[string]json.RawMessage `json:"filter"`
	HiddenOptionIDs    []string                   `json:"hiddenOptionIds"`
	Icon               string                     `json:"icon"`
	IsTemplate         bool                       `json:"isTemplate"`
	KanbanCalculations map[string]json.RawMessage `json:"kanbanCalculations"`
	Properties         map[string]any             `json:"properties"`
	SortOptions        []json.RawMessage          `json:"sortOptions"`
	ViewType           string                     `json:"viewType"`
	VisibleOptionIDs   []string                   `json:"visibleOptionIds"`
	VisiblePropertyIDs []string                   `json:"visiblePropertyIds"`
}

type rawBlock struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	BoardID    string         `json:"boardId"`
	ParentID   string         `json:"parentId"`
	Title      string         `json:"title"`
	CreateAt   int64          `json:"createAt"`
	UpdateAt   int64          `json:"updateAt"`
	DeleteAt   int64          `json:"deleteAt"`
	CreatedBy  string         `json:"createdBy"`
	ModifiedBy string         `json:"modifiedBy"`
	Schema     int64          `json:"schema"`
	Fields     rawBlockFields `json:"fields"`
}

type rawBoardMember struct {
	BoardID         string `json:"boardId"`
	UserID          string `json:"userId"`
	MinimumRole     string `json:"minimumRole"`
	Roles           string `json:"roles"`
	SchemeAdmin     bool   `json:"schemeAdmin"`
	SchemeCommenter bool   `json:"schemeCommenter"`
	SchemeEditor    bool   `json:"schemeEditor"`
	SchemeViewer    bool   `json:"schemeViewer"`
	Synthetic       bool   `json:"synthetic"`
}

type parsedArchive struct {
	SourceHash    string
	NestedHash    string
	SourceVersion int
	Boards        []*rawBoard
	Cards         []*rawBlock
	Texts         []*rawBlock
	Views         []*rawBlock
	BoardMembers  []rawBoardMember
	Attachments   int
	MaxUpdated    time.Time
}
