package commands

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"code.vikunja.io/veans/internal/config"
)

func TestPrimeTemplate_RendersAnchors(t *testing.T) {
	data := primeContext{
		Server:            "https://vikunja.example.com",
		ProjectID:         42,
		ProjectTitle:      "Test Project",
		ProjectIdentifier: "PROJ",
		ViewID:            7,
		Buckets:           config.Buckets{Todo: 11, InProgress: 12, InReview: 13, Done: 14, Scrapped: 15},
		BotUsername:       "bot-myrepo",
		TaskIDExample:     "PROJ-1",
	}
	tpl, err := template.New("prime").Parse(promptTemplate)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	mustContain := []string{
		"<EXTREMELY_IMPORTANT>",
		"</EXTREMELY_IMPORTANT>",
		"bot-myrepo",
		"Test Project",
		"PROJ-1",
		"Refs:",
		"veans claim",
		"veans list --ready",
		"--description-replace-old",
		"Todo",
		"In Progress",
		"In Review",
		"Done",
		"Scrapped",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("rendered prompt missing %q", s)
		}
	}

	// Buckets show concrete IDs.
	for _, want := range []string{"`11`", "`12`", "`13`", "`14`", "`15`"} {
		if !strings.Contains(out, want) {
			t.Errorf("bucket id %s not present in output", want)
		}
	}
}

func TestPrimeTemplate_NoIdentifierFallback(t *testing.T) {
	data := primeContext{
		ProjectTitle:      "No Ident",
		ProjectIdentifier: "",
		BotUsername:       "bot-x",
		TaskIDExample:     "#1",
		Server:            "https://vikunja.example.com",
		ProjectID:         1,
		ViewID:            1,
	}
	tpl, _ := template.New("prime").Parse(promptTemplate)
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "no identifier") {
		t.Errorf("expected fallback copy when project has no identifier; got:\n%s", out)
	}
	if !strings.Contains(out, "#NN") {
		t.Errorf("expected #NN format mention in fallback")
	}
}
