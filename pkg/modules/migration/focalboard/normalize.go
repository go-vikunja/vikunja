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
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var strictDuePattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})(?: 00:00:00)?$`)

var priorityMap = map[string]int64{
	"🔥 P0 Срочно + важно":   5,
	"🧭 P1 Важно, не срочно": 3,
	"⚡ P2 Срочно, не важно": 2,
	"🧊 P3 Backlog":          1,
}

func rawString(value any, strict bool) (string, error) {
	if value == nil {
		return "", nil
	}
	if s, ok := value.(string); ok {
		return s, nil
	}
	if strict {
		return "", fmt.Errorf("property value must be a string or null, got %T", value)
	}
	return fmt.Sprint(value), nil
}

func sourceTimestamp(millis int64) string {
	if millis <= 0 {
		return ""
	}
	return time.UnixMilli(millis).UTC().Format(time.RFC3339Nano)
}

func validateBoardPropertySchema(board *rawBoard) error {
	expected := map[string]string{"Status": "select", "Кто делает": "text", "Приоритет": "text", "Срок": "text"}
	seenIDs := map[string]struct{}{}
	seenNames := map[string]struct{}{}
	for _, property := range board.CardProperties {
		if property.ID == "" || property.Name == "" {
			return fmt.Errorf("board %s has property with empty id or name", board.ID)
		}
		if _, exists := seenIDs[property.ID]; exists {
			return fmt.Errorf("board %s has duplicate property id %q", board.ID, property.ID)
		}
		if _, exists := seenNames[property.Name]; exists {
			return fmt.Errorf("board %s has duplicate property name %q", board.ID, property.Name)
		}
		seenIDs[property.ID] = struct{}{}
		seenNames[property.Name] = struct{}{}
		expectedType, required := expected[property.Name]
		if property.Type != "text" && property.Type != "select" {
			return fmt.Errorf("board %s property %q has unsupported type %q", board.ID, property.Name, property.Type)
		}
		if required && property.Type != expectedType {
			return fmt.Errorf("board %s property %q must have type %q, got %q", board.ID, property.Name, expectedType, property.Type)
		}
		if property.Type == "text" && len(property.Options) != 0 {
			return fmt.Errorf("board %s text property %q must not define options", board.ID, property.Name)
		}
		if property.Type == "select" && len(property.Options) == 0 {
			return fmt.Errorf("board %s select property %q must define options", board.ID, property.Name)
		}
		optionIDs := map[string]struct{}{}
		optionValues := map[string]struct{}{}
		for _, option := range property.Options {
			if option.ID == "" {
				return fmt.Errorf("board %s property %q has option with empty id", board.ID, property.Name)
			}
			if option.Value == "" {
				return fmt.Errorf("board %s property %q has option with empty value", board.ID, property.Name)
			}
			if _, exists := optionIDs[option.ID]; exists {
				return fmt.Errorf("board %s property %q has duplicate option id %q", board.ID, property.Name, option.ID)
			}
			if _, exists := optionValues[option.Value]; exists {
				return fmt.Errorf("board %s property %q has duplicate option value %q", board.ID, property.Name, option.Value)
			}
			optionIDs[option.ID] = struct{}{}
			optionValues[option.Value] = struct{}{}
		}
	}
	for name := range expected {
		if _, exists := seenNames[name]; !exists {
			return fmt.Errorf("board %s is missing required property %q", board.ID, name)
		}
	}
	return nil
}

func validateCardPropertyValues(board *rawBoard, card *rawBlock, strict bool) (map[string]string, error) {
	propertiesByID := make(map[string]*rawCardProperty, len(board.CardProperties))
	for i := range board.CardProperties {
		propertiesByID[board.CardProperties[i].ID] = &board.CardProperties[i]
	}
	values := make(map[string]string, len(card.Fields.Properties))
	for propertyID, value := range card.Fields.Properties {
		property, exists := propertiesByID[propertyID]
		if !exists {
			return nil, fmt.Errorf("card %s references unknown property id %q", card.ID, propertyID)
		}
		raw, err := rawString(value, strict)
		if err != nil {
			return nil, fmt.Errorf("card %s property %q: %w", card.ID, property.Name, err)
		}
		if property.Type == "select" && raw != "" {
			known := false
			for _, option := range property.Options {
				if option.ID == raw {
					known = true
					break
				}
			}
			if !known {
				return nil, fmt.Errorf("card %s property %q references unknown option id %q", card.ID, property.Name, raw)
			}
		}
		values[property.Name] = raw
	}
	return values, nil
}

func propertyByName(board *rawBoard, name string) *rawCardProperty {
	for i := range board.CardProperties {
		if board.CardProperties[i].Name == name {
			return &board.CardProperties[i]
		}
	}
	return nil
}

func validatedStatusOptions(property *rawCardProperty) ([]string, error) {
	if property == nil {
		return nil, fmt.Errorf("board is missing required Status property")
	}
	if property.Type != "select" {
		return nil, fmt.Errorf("status property must have type select")
	}
	values := make([]string, 0, len(property.Options))
	seen := map[string]struct{}{}
	for _, option := range property.Options {
		if option.ID == "" {
			return nil, fmt.Errorf("status property contains an option with empty id")
		}
		if option.Value == "" {
			return nil, fmt.Errorf("status property contains an empty option")
		}
		if _, exists := seen[option.Value]; exists {
			return nil, fmt.Errorf("status property contains duplicate option %q", option.Value)
		}
		seen[option.Value] = struct{}{}
		values = append(values, option.Value)
	}
	matches := func(expected []string) bool {
		if len(seen) != len(expected) {
			return false
		}
		for _, value := range expected {
			if _, ok := seen[value]; !ok {
				return false
			}
		}
		return true
	}
	if !matches([]string{"Todo", "In Progress", "Done"}) && !matches([]string{"Новая", "В работе", "Разобрана"}) {
		return nil, fmt.Errorf("unsupported Status option schema")
	}
	sort.Strings(values)
	return values, nil
}

func requiredTextProperty(board *rawBoard, name string) (*rawCardProperty, error) {
	property := propertyByName(board, name)
	if property == nil {
		return nil, fmt.Errorf("board %s is missing required property %q", board.ID, name)
	}
	if property.Type != "text" {
		return nil, fmt.Errorf("board %s property %q must have type text", board.ID, name)
	}
	return property, nil
}

func propertyRaw(card *rawBlock, property *rawCardProperty, strict bool) (string, error) {
	if property == nil || card.Fields.Properties == nil {
		return "", nil
	}
	return rawString(card.Fields.Properties[property.ID], strict)
}

func statusRaw(card *rawBlock, property *rawCardProperty, strict bool) (string, error) {
	raw, err := propertyRaw(card, property, strict)
	if err != nil || property == nil || raw == "" {
		return raw, err
	}
	for _, option := range property.Options {
		if option.ID == raw {
			return option.Value, nil
		}
	}
	if strict {
		return "", fmt.Errorf("card %s status references unknown option id %q", card.ID, raw)
	}
	return raw, nil
}

func buildAssigneeLookup(mapping *AssigneeMap) (map[string]AssigneeMapping, error) {
	lookup := map[string]AssigneeMapping{}
	seenSources := map[string]struct{}{}
	if mapping == nil {
		return lookup, nil
	}
	if mapping.SchemaVersion != "" && mapping.SchemaVersion != AssigneeMapSchemaVersion {
		return nil, fmt.Errorf("unsupported assignee map schema version %q", mapping.SchemaVersion)
	}
	for _, item := range mapping.Mappings {
		if item.SourceRaw == "" {
			return nil, fmt.Errorf("assignee mapping contains an empty source_raw")
		}
		if _, exists := seenSources[item.SourceRaw]; exists {
			return nil, fmt.Errorf("duplicate assignee mapping source value")
		}
		seenSources[item.SourceRaw] = struct{}{}
		if item.Unassigned && item.TargetUsernameOrEmail != "" {
			return nil, fmt.Errorf("assignee mapping for a source value cannot set both unassigned and a target")
		}
		if !item.Unassigned && item.TargetUsernameOrEmail == "" {
			// Generated template rows are intentionally incomplete. Keep them
			// unmapped so they remain visible in exceptions and warnings.
			continue
		}
		lookup[item.SourceRaw] = item
	}
	return lookup, nil
}

func normalizeParsed(parsed *parsedArchive, opts Options) (*NormalizedArchive, *Reconciliation, error) {
	if opts.VikunjaVersion == "" {
		opts.VikunjaVersion = DefaultVikunjaVersion
	}
	if opts.VikunjaVersion != DefaultVikunjaVersion {
		return nil, nil, fmt.Errorf("unsupported Vikunja export VERSION %q; this converter is pinned to %q", opts.VikunjaVersion, DefaultVikunjaVersion)
	}
	var location *time.Location
	var err error
	if opts.Timezone != "" {
		location, err = time.LoadLocation(opts.Timezone)
		if err != nil {
			return nil, nil, fmt.Errorf("load timezone %q: %w", opts.Timezone, err)
		}
	}
	assigneeLookup, err := buildAssigneeLookup(opts.Assignees)
	if err != nil {
		return nil, nil, err
	}
	boardsByID := make(map[string]*rawBoard, len(parsed.Boards))
	for _, board := range parsed.Boards {
		boardsByID[board.ID] = board
	}
	for _, card := range parsed.Cards {
		if _, exists := boardsByID[card.BoardID]; !exists {
			return nil, nil, fmt.Errorf("card %s references unknown board %s", card.ID, card.BoardID)
		}
	}
	for _, text := range parsed.Texts {
		if _, exists := boardsByID[text.BoardID]; !exists {
			return nil, nil, fmt.Errorf("text %s references unknown board %s", text.ID, text.BoardID)
		}
	}

	counts := Counts{
		Boards:      len(parsed.Boards),
		Cards:       len(parsed.Cards),
		TextBlocks:  len(parsed.Texts),
		KanbanViews: len(parsed.Views),
		Attachments: parsed.Attachments,
		Statuses:    map[string]int{},
	}
	warnings := []Warning{}
	for _, board := range parsed.Boards {
		if board.IsTemplate {
			counts.Templates++
		}
	}
	for _, card := range parsed.Cards {
		if card.DeleteAt != 0 {
			counts.DeletedCards++
		}
		if card.Fields.IsTemplate {
			counts.Templates++
		}
		if strings.TrimSpace(card.Title) == "" {
			counts.EmptyTitles++
		}
	}
	if counts.DeletedCards > 0 || counts.Templates > 0 {
		return nil, nil, fmt.Errorf("deleted/template cards are not imported: deleted=%d templates=%d", counts.DeletedCards, counts.Templates)
	}
	if opts.Strict && counts.EmptyTitles > 0 {
		return nil, nil, fmt.Errorf("cards with empty titles are not importable: %d", counts.EmptyTitles)
	}

	textsByID := make(map[string]*rawBlock, len(parsed.Texts))
	for _, text := range parsed.Texts {
		textsByID[text.ID] = text
	}
	usedTexts := map[string]string{}
	descriptionByCard := map[string]string{}
	methodByCard := map[string]DescriptionLinkMethod{}
	textIDByCard := map[string]string{}
	brokenCards := []*rawBlock{}
	for _, card := range parsed.Cards {
		if len(card.Fields.ContentOrder) == 0 {
			methodByCard[card.ID] = DescriptionNone
			continue
		}
		if len(card.Fields.ContentOrder) != 1 {
			return nil, nil, fmt.Errorf("card %s has unsupported contentOrder length %d", card.ID, len(card.Fields.ContentOrder))
		}
		textID := card.Fields.ContentOrder[0]
		text, exists := textsByID[textID]
		if !exists {
			brokenCards = append(brokenCards, card)
			continue
		}
		if text.BoardID != card.BoardID {
			return nil, nil, fmt.Errorf("card %s links text %s from another board", card.ID, text.ID)
		}
		if previous, exists := usedTexts[text.ID]; exists {
			return nil, nil, fmt.Errorf("direct description link is not one-to-one: text %s is referenced by cards %s and %s", text.ID, previous, card.ID)
		}
		usedTexts[text.ID] = card.ID
		descriptionByCard[card.ID] = text.Title
		methodByCard[card.ID] = DescriptionDirect
		textIDByCard[card.ID] = text.ID
		counts.DirectDescriptions++
	}

	orphansByBoard := map[string][]*rawBlock{}
	for _, text := range parsed.Texts {
		if _, used := usedTexts[text.ID]; !used {
			orphansByBoard[text.BoardID] = append(orphansByBoard[text.BoardID], text)
		}
	}
	type recoverySelection struct {
		text  *rawBlock
		delta int64
	}
	selections := map[string]recoverySelection{}
	consumed := map[string]string{}
	orphanCount := 0
	for _, texts := range orphansByBoard {
		orphanCount += len(texts)
	}
	for _, card := range brokenCards {
		var nearest []*rawBlock
		nearestDelta := int64(1501)
		for _, text := range orphansByBoard[card.BoardID] {
			delta := text.CreateAt - card.CreateAt
			if delta <= 0 || delta > 1500 {
				continue
			}
			if delta < nearestDelta {
				nearestDelta = delta
				nearest = []*rawBlock{text}
			} else if delta == nearestDelta {
				nearest = append(nearest, text)
			}
		}
		if len(nearest) != 1 {
			return nil, nil, fmt.Errorf("description recovery for card %s is ambiguous or missing: nearest_candidates=%d", card.ID, len(nearest))
		}
		text := nearest[0]
		if previous, exists := consumed[text.ID]; exists {
			return nil, nil, fmt.Errorf("description recovery is not one-to-one: text %s is the nearest candidate for cards %s and %s", text.ID, previous, card.ID)
		}
		consumed[text.ID] = card.ID
		selections[card.ID] = recoverySelection{text: text, delta: nearestDelta}
	}
	// Every orphan must participate in the unique global one-to-one nearest
	// assignment. This catches unequal secondary candidates too: they cannot be
	// silently discarded after selecting the nearer text.
	if len(selections) != orphanCount {
		return nil, nil, fmt.Errorf("orphan text blocks remain after recovery: %d", orphanCount-len(selections))
	}
	recovered := make([]RecoveredDescription, 0, len(brokenCards))
	for _, card := range brokenCards {
		selection := selections[card.ID]
		text := selection.text
		usedTexts[text.ID] = card.ID
		descriptionByCard[card.ID] = text.Title
		methodByCard[card.ID] = DescriptionRecovered
		textIDByCard[card.ID] = text.ID
		recovered = append(recovered, RecoveredDescription{SourceBoardID: card.BoardID, SourceCardID: card.ID, SourceTextID: text.ID, DeltaMS: selection.delta})
		counts.RecoveredDescriptions++
	}
	if len(recovered) != len(consumed) {
		return nil, nil, fmt.Errorf("recovery count mismatch: recovered=%d consumed=%d", len(recovered), len(consumed))
	}
	if len(usedTexts) != len(parsed.Texts) {
		return nil, nil, fmt.Errorf("orphan text blocks remain after recovery: %d", len(parsed.Texts)-len(usedTexts))
	}
	sort.Slice(recovered, func(i, j int) bool {
		if recovered[i].SourceBoardID == recovered[j].SourceBoardID {
			return recovered[i].SourceCardID < recovered[j].SourceCardID
		}
		return recovered[i].SourceBoardID < recovered[j].SourceBoardID
	})
	counts.FinalDescriptions = counts.DirectDescriptions + counts.RecoveredDescriptions
	counts.CardsWithoutDescription = counts.Cards - counts.FinalDescriptions

	cardsByBoard := map[string][]*rawBlock{}
	for _, card := range parsed.Cards {
		cardsByBoard[card.BoardID] = append(cardsByBoard[card.BoardID], card)
	}
	assigneeValues := map[string]struct{}{}
	unknownAssignees := map[string]struct{}{}
	mappingRows := []CardMapping{}
	viewIDsByBoard := map[string][]string{}
	for _, view := range parsed.Views {
		viewIDsByBoard[view.BoardID] = append(viewIDsByBoard[view.BoardID], view.ID)
	}
	for boardID := range viewIDsByBoard {
		sort.Strings(viewIDsByBoard[boardID])
	}
	normalizedBoards := make([]NormalizedBoard, 0, len(parsed.Boards))
	legacyTaskID := int64(1)
	for boardIndex, board := range parsed.Boards {
		if err := validateBoardPropertySchema(board); err != nil {
			return nil, nil, err
		}
		statusProperty := propertyByName(board, "Status")
		statusOptions, statusErr := validatedStatusOptions(statusProperty)
		if statusErr != nil {
			return nil, nil, fmt.Errorf("board %s: %w", board.ID, statusErr)
		}
		assigneeProperty, propertyErr := requiredTextProperty(board, "Кто делает")
		if propertyErr != nil {
			return nil, nil, propertyErr
		}
		priorityProperty, propertyErr := requiredTextProperty(board, "Приоритет")
		if propertyErr != nil {
			return nil, nil, propertyErr
		}
		dueProperty, propertyErr := requiredTextProperty(board, "Срок")
		if propertyErr != nil {
			return nil, nil, propertyErr
		}
		titleCounts := map[string]int{}
		normalizedBoard := NormalizedBoard{
			SourceBoardID:   board.ID,
			SourceViewIDs:   append([]string{}, viewIDsByBoard[board.ID]...),
			StatusOptions:   statusOptions,
			Title:           board.Title,
			SourceCreatedAt: sourceTimestamp(board.CreateAt),
			SourceUpdatedAt: sourceTimestamp(board.UpdateAt),
			Tasks:           []NormalizedTask{},
		}
		for _, card := range cardsByBoard[board.ID] {
			sourceProperties, valueErr := validateCardPropertyValues(board, card, opts.Strict)
			if valueErr != nil {
				return nil, nil, valueErr
			}
			status, valueErr := statusRaw(card, statusProperty, opts.Strict)
			if valueErr != nil {
				return nil, nil, fmt.Errorf("card %s Status: %w", card.ID, valueErr)
			}
			counts.Statuses[status]++
			if status == "" {
				if opts.Strict {
					return nil, nil, fmt.Errorf("card %s has empty status", card.ID)
				}
				warnings = append(warnings, Warning{Code: "empty_status", SourceBoardID: board.ID, SourceCardID: card.ID})
			}
			priorityRaw, valueErr := propertyRaw(card, priorityProperty, opts.Strict)
			if valueErr != nil {
				return nil, nil, fmt.Errorf("card %s priority: %w", card.ID, valueErr)
			}
			priority, knownPriority := priorityMap[priorityRaw]
			if priorityRaw != "" {
				counts.NonEmptyPriorities++
			}
			if !knownPriority {
				warnings = append(warnings, Warning{Code: "priority_unset", SourceBoardID: board.ID, SourceCardID: card.ID})
			}
			assigneeRaw, valueErr := propertyRaw(card, assigneeProperty, opts.Strict)
			if valueErr != nil {
				return nil, nil, fmt.Errorf("card %s assignee: %w", card.ID, valueErr)
			}
			assigneeTarget := ""
			assigneeUnassigned := assigneeRaw == ""
			assigneeMapped := false
			if assigneeRaw != "" {
				counts.NonEmptyAssignees++
				assigneeValues[assigneeRaw] = struct{}{}
				if item, exists := assigneeLookup[assigneeRaw]; exists {
					assigneeTarget = item.TargetUsernameOrEmail
					assigneeUnassigned = item.Unassigned
					assigneeMapped = item.Unassigned || item.TargetUsernameOrEmail != ""
				} else {
					unknownAssignees[assigneeRaw] = struct{}{}
					warnings = append(warnings, Warning{Code: "unmapped_assignee", SourceBoardID: board.ID, SourceCardID: card.ID})
				}
			}
			dueRaw, valueErr := propertyRaw(card, dueProperty, opts.Strict)
			if valueErr != nil {
				return nil, nil, fmt.Errorf("card %s due: %w", card.ID, valueErr)
			}
			parsedDue := ""
			nativeDue := ""
			if dueRaw == "" {
				counts.EmptyDueValues++
			} else {
				counts.NonEmptyDueRawValues++
				matches := strictDuePattern.FindStringSubmatch(dueRaw)
				if matches == nil {
					counts.FreeTextDueValues++
					warnings = append(warnings, Warning{Code: "free_text_due_preserved", SourceBoardID: board.ID, SourceCardID: card.ID})
				} else {
					date, parseErr := time.Parse("2006-01-02", matches[1])
					if parseErr != nil {
						counts.FreeTextDueValues++
						warnings = append(warnings, Warning{Code: "invalid_iso_due_preserved", SourceBoardID: board.ID, SourceCardID: card.ID})
					} else {
						counts.StrictISODueValues++
						parsedDue = date.Format("2006-01-02")
						if location == nil {
							warnings = append(warnings, Warning{Code: "timezone_required_for_native_due", SourceBoardID: board.ID, SourceCardID: card.ID})
						} else {
							nativeDue = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location).Format(time.RFC3339)
						}
					}
				}
			}
			titleCounts[card.Title]++
			normalizedTask := NormalizedTask{
				SourceSystem:          "focalboard",
				SourceArchiveSHA256:   parsed.SourceHash,
				SourceBoardID:         board.ID,
				SourceCardID:          card.ID,
				SourceTextID:          textIDByCard[card.ID],
				Title:                 card.Title,
				Description:           descriptionByCard[card.ID],
				DescriptionLinkMethod: methodByCard[card.ID],
				StatusRaw:             status,
				SourceProperties:      sourceProperties,
				AssigneeRaw:           assigneeRaw,
				AssigneeTarget:        assigneeTarget,
				AssigneeUnassigned:    assigneeUnassigned,
				PriorityRaw:           priorityRaw,
				NativePriority:        priority,
				DueRaw:                dueRaw,
				ParsedDueCandidate:    parsedDue,
				NativeDueDate:         nativeDue,
				SourceCreatedAt:       sourceTimestamp(card.CreateAt),
				SourceUpdatedAt:       sourceTimestamp(card.UpdateAt),
			}
			normalizedBoard.Tasks = append(normalizedBoard.Tasks, normalizedTask)
			mappingRows = append(mappingRows, CardMapping{
				SourceBoardID:         board.ID,
				SourceCardID:          card.ID,
				SourceTextID:          normalizedTask.SourceTextID,
				SourceCreatedAt:       normalizedTask.SourceCreatedAt,
				SourceUpdatedAt:       normalizedTask.SourceUpdatedAt,
				TargetProjectLegacyID: int64(boardIndex + 2),
				TargetTaskLegacyID:    legacyTaskID,
				DescriptionLinkMethod: normalizedTask.DescriptionLinkMethod,
				NativePriority:        priority,
				NativeDueDateSet:      nativeDue != "",
				AssigneeMapped:        assigneeMapped,
			})
			legacyTaskID++
		}
		for _, n := range titleCounts {
			if n > 1 {
				counts.DuplicateTitleExtras += n - 1
			}
		}
		normalizedBoards = append(normalizedBoards, normalizedBoard)
	}
	counts.UniqueAssigneeRawValues = len(assigneeValues)

	unknown := make([]string, 0, len(unknownAssignees))
	for value := range unknownAssignees {
		unknown = append(unknown, value)
	}
	sort.Strings(unknown)
	sort.Slice(mappingRows, func(i, j int) bool {
		if mappingRows[i].SourceBoardID == mappingRows[j].SourceBoardID {
			return mappingRows[i].SourceCardID < mappingRows[j].SourceCardID
		}
		return mappingRows[i].SourceBoardID < mappingRows[j].SourceBoardID
	})
	runID, err := deterministicRunID(parsed, opts)
	if err != nil {
		return nil, nil, err
	}
	normalized := &NormalizedArchive{
		SchemaVersion:       NormalizedSchemaVersion,
		SourceSystem:        "focalboard",
		SourceArchiveSHA256: parsed.SourceHash,
		SourceNestedSHA256:  parsed.NestedHash,
		SourceVersion:       parsed.SourceVersion,
		MigrationRun: MigrationRun{
			RunID:          runID,
			ToolVersion:    ConverterVersion,
			Strict:         opts.Strict,
			Timezone:       opts.Timezone,
			VikunjaVersion: opts.VikunjaVersion,
		},
		Counts: counts,
		Boards: normalizedBoards,
	}
	reconciliation := &Reconciliation{
		SchemaVersion:         NormalizedSchemaVersion,
		RunID:                 runID,
		SourceArchiveSHA256:   parsed.SourceHash,
		Counts:                counts,
		RecoveredDescriptions: recovered,
		CardMappings:          mappingRows,
		UnknownAssignees:      unknown,
		Warnings:              warnings,
		Verified:              false,
	}
	if opts.Strict && parsed.SourceHash == KnownSourceSHA256 {
		if err := validateKnownCounts(counts); err != nil {
			return nil, nil, err
		}
	}
	return normalized, reconciliation, nil
}

func canonicalAssigneeMap(mapping *AssigneeMap) AssigneeMap {
	if mapping == nil {
		return AssigneeMap{SchemaVersion: AssigneeMapSchemaVersion, Mappings: []AssigneeMapping{}}
	}
	copyMap := AssigneeMap{SchemaVersion: mapping.SchemaVersion}
	if copyMap.SchemaVersion == "" {
		copyMap.SchemaVersion = AssigneeMapSchemaVersion
	}
	copyMap.Mappings = append([]AssigneeMapping{}, mapping.Mappings...)
	sort.Slice(copyMap.Mappings, func(i, j int) bool { return copyMap.Mappings[i].SourceRaw < copyMap.Mappings[j].SourceRaw })
	return copyMap
}

func configurationHash(opts Options) (string, error) {
	config, err := json.Marshal(struct {
		Strict    bool        `json:"strict"`
		Timezone  string      `json:"timezone"`
		Version   string      `json:"version"`
		Assignees AssigneeMap `json:"assignees"`
	}{opts.Strict, opts.Timezone, opts.VikunjaVersion, canonicalAssigneeMap(opts.Assignees)})
	if err != nil {
		return "", fmt.Errorf("marshal converter configuration: %w", err)
	}
	return sha256Hex(config), nil
}

func deterministicRunID(parsed *parsedArchive, opts Options) (string, error) {
	configHash, err := configurationHash(opts)
	if err != nil {
		return "", err
	}
	identity := []byte(parsed.SourceHash + "\n" + parsed.NestedHash + "\n" + configHash + "\n" + ConverterVersion)
	return sha256Hex(identity)[:24], nil
}

func validateKnownCounts(actual Counts) error {
	expected := Counts{
		Boards: 9, Cards: 858, TextBlocks: 852, KanbanViews: 8, Attachments: 0,
		DeletedCards: 0, Templates: 0, EmptyTitles: 0,
		Statuses:           map[string]int{"Todo": 808, "Новая": 50},
		NonEmptyPriorities: 802, NonEmptyAssignees: 804, UniqueAssigneeRawValues: 28,
		NonEmptyDueRawValues: 305, StrictISODueValues: 25, FreeTextDueValues: 280, EmptyDueValues: 553,
		DirectDescriptions: 776, RecoveredDescriptions: 76, CardsWithoutDescription: 6, FinalDescriptions: 852,
		DuplicateTitleExtras: 2,
	}
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return fmt.Errorf("marshal actual source counts: %w", err)
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return fmt.Errorf("marshal expected source counts: %w", err)
	}
	if string(actualJSON) != string(expectedJSON) {
		return fmt.Errorf("known archive counts differ: expected=%s actual=%s", expectedJSON, actualJSON)
	}
	return nil
}

func Analyze(inputPath string, opts Options) (*Analysis, error) {
	parsed, err := parsePackage(inputPath, opts)
	if err != nil {
		return nil, err
	}
	normalized, reconciliation, err := normalizeParsed(parsed, opts)
	if err != nil {
		return nil, err
	}
	return &Analysis{
		SchemaVersion:       NormalizedSchemaVersion,
		SourceArchiveSHA256: parsed.SourceHash,
		SourceNestedSHA256:  parsed.NestedHash,
		SourceVersion:       parsed.SourceVersion,
		Counts:              normalized.Counts,
		Warnings:            reconciliation.Warnings,
	}, nil
}
