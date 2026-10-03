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

package todoist

import (
	"maps"
	"regexp"
	"slices"
)

// repeatLang is the word data of one language for the language-agnostic grammar in
// parseRepeatTokens.
type repeatLang struct {
	intervals  map[string]int64
	adverbs    map[string]int64
	everyWords []string // typos like "elke!" are normalized away before the grammar runs
	// otherWord stays empty for languages with no single fixed form for it.
	otherWord       string
	lastWord        string
	dayWord         string
	weekdays        map[string]struct{}
	months          map[string]struct{}
	ordinals        string // regex alternation of ordinal suffixes: "st|nd|rd|th"; nl: "e|ste|de"
	ordinalDayRegex *regexp.Regexp
}

// langPacks is deliberately data-only: a new language is a new data block here, with zero
// grammar changes. Todoist plans to move its date parsing upstream into a multilingual
// parser, so keep this table and the small parseTodoistRepeat entry point swappable for it.
var langPacks = compileRepeatLangs(map[string]*repeatLang{
	"en": {
		intervals: map[string]int64{
			"day": secondsPerDay, "days": secondsPerDay,
			"week": secondsPerWeek, "weeks": secondsPerWeek,
			"month": secondsPerMonth, "months": secondsPerMonth,
			"year": secondsPerYear, "years": secondsPerYear,
		},
		adverbs: map[string]int64{
			"daily":    secondsPerDay,
			"weekly":   secondsPerWeek,
			"monthly":  secondsPerMonth,
			"yearly":   secondsPerYear,
			"annually": secondsPerYear,
		},
		everyWords: []string{"every"},
		otherWord:  "other",
		lastWord:   "last",
		dayWord:    "day",
		weekdays: map[string]struct{}{
			"monday": {}, "tuesday": {}, "wednesday": {}, "thursday": {},
			"friday": {}, "saturday": {}, "sunday": {},
		},
		months: map[string]struct{}{
			"january": {}, "jan": {}, "february": {}, "feb": {},
			"march": {}, "mar": {}, "april": {}, "apr": {},
			"may": {}, "june": {}, "jun": {}, "july": {}, "jul": {},
			"august": {}, "aug": {}, "september": {}, "sept": {}, "sep": {},
			"october": {}, "oct": {}, "november": {}, "nov": {},
			"december": {}, "dec": {},
		},
		ordinals: "st|nd|rd|th",
	},
	"nl": {
		intervals: map[string]int64{
			"dag": secondsPerDay, "dagen": secondsPerDay,
			"week": secondsPerWeek, "weken": secondsPerWeek,
			"maand": secondsPerMonth, "maanden": secondsPerMonth,
			"jaar": secondsPerYear, "jaren": secondsPerYear,
		},
		adverbs: map[string]int64{
			"dagelijks":   secondsPerDay,
			"wekelijks":   secondsPerWeek,
			"maandelijks": secondsPerMonth,
			"jaarlijks":   secondsPerYear,
		},
		everyWords: []string{"elke", "elk"},
		// Dutch inflects "other" ("elke andere dag"), so one fixed otherWord can only
		// ever match half of the form - leave it unset instead.
		otherWord: "",
		lastWord:  "laatste",
		dayWord:   "dag",
		weekdays: map[string]struct{}{
			"maandag": {}, "dinsdag": {}, "woensdag": {}, "donderdag": {},
			"vrijdag": {}, "zaterdag": {}, "zondag": {},
		},
		months: map[string]struct{}{
			"januari": {}, "jan": {}, "februari": {}, "feb": {},
			"maart": {}, "mrt": {}, "april": {}, "apr": {},
			"mei": {}, "juni": {}, "jun": {}, "juli": {}, "jul": {},
			"augustus": {}, "aug": {}, "september": {}, "sept": {}, "sep": {},
			"oktober": {}, "okt": {}, "november": {}, "nov": {},
			"december": {}, "dec": {},
		},
		ordinals: "e|ste|de",
	},
})

func compileRepeatLangs(packs map[string]*repeatLang) map[string]*repeatLang {
	for _, pack := range packs {
		pack.ordinalDayRegex = regexp.MustCompile(`^\d{1,2}(?:` + pack.ordinals + `)?$`)
	}
	return packs
}

func (l *repeatLang) isEveryWord(token string) bool {
	return slices.Contains(l.everyWords, token)
}

func (l *repeatLang) isYearlyAdverb(token string) bool {
	return l.adverbs[token] == secondsPerYear
}

// langPacksFor returns the word packs to try for a due object's language. Recurring due
// objects from real Todoist exports always carry a lang; when it is missing there is
// nothing to key on and every pack is tried as a whole. Unknown languages fall back to
// English.
func langPacksFor(lang string) []*repeatLang {
	if pack, ok := langPacks[lang]; ok {
		return []*repeatLang{pack}
	}
	if lang != "" {
		return []*repeatLang{langPacks["en"]}
	}

	packs := make([]*repeatLang, 0, len(langPacks))
	for _, l := range slices.Sorted(maps.Keys(langPacks)) {
		packs = append(packs, langPacks[l])
	}
	return packs
}
