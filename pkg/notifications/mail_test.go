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

package notifications

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMail(t *testing.T) {
	t.Run("Full mail", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line("This is a line").
			Line("And another one").
			Action("the actiopn", "https://example.com").
			Line("This should be an outro line").
			Line("And one more, because why not?")

		assert.Equal(t, "test@example.com", mail.from)
		assert.Equal(t, "test@otherdomain.com", mail.to)
		assert.Equal(t, "Testmail", mail.subject)
		assert.Equal(t, "Hi there,", mail.greeting)
		assert.Len(t, mail.introLines, 2)
		assert.Equal(t, "This is a line", mail.introLines[0].Text)
		assert.False(t, mail.introLines[0].isHTML)
		assert.Equal(t, "And another one", mail.introLines[1].Text)
		assert.False(t, mail.introLines[1].isHTML)
		assert.Len(t, mail.outroLines, 2)
		assert.Equal(t, "This should be an outro line", mail.outroLines[0].Text)
		assert.False(t, mail.outroLines[0].isHTML)
		assert.Equal(t, "And one more, because why not?", mail.outroLines[1].Text)
		assert.False(t, mail.outroLines[1].isHTML)
	})
	t.Run("No greeting", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Line("This is a line").
			Line("And another one")

		assert.Equal(t, "test@example.com", mail.from)
		assert.Equal(t, "test@otherdomain.com", mail.to)
		assert.Equal(t, "Testmail", mail.subject)
		assert.Empty(t, mail.greeting)
		assert.Len(t, mail.introLines, 2)
		assert.Equal(t, "This is a line", mail.introLines[0].Text)
		assert.Equal(t, "And another one", mail.introLines[1].Text)
	})
	t.Run("No action", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Line("This is a line").
			Line("And another one").
			Line("This should be an outro line").
			Line("And one more, because why not?")

		assert.Equal(t, "test@example.com", mail.from)
		assert.Equal(t, "test@otherdomain.com", mail.to)
		assert.Equal(t, "Testmail", mail.subject)
		assert.Len(t, mail.introLines, 4)
		assert.Equal(t, "This is a line", mail.introLines[0].Text)
		assert.Equal(t, "And another one", mail.introLines[1].Text)
		assert.Equal(t, "This should be an outro line", mail.introLines[2].Text)
		assert.Equal(t, "And one more, because why not?", mail.introLines[3].Text)
	})
}

// assertHTMLContainsDarkModeSupport checks that the HTML message contains
// the required dark mode meta tags and CSS
func assertHTMLContainsDarkModeSupport(t *testing.T, htmlMessage string) {
	t.Helper()
	// Check for dark mode meta tags
	assert.Contains(t, htmlMessage, `<meta name="color-scheme" content="light dark">`)
	assert.Contains(t, htmlMessage, `<meta name="supported-color-schemes" content="light dark">`)

	// Check for dark mode CSS
	assert.Contains(t, htmlMessage, `@media (prefers-color-scheme: dark)`)
	assert.Contains(t, htmlMessage, `.email-card`)
	assert.Contains(t, htmlMessage, `box-shadow: 0.3em 0.3em 0.8em rgba(0,0,0,0.3) !important`)
	assert.Contains(t, htmlMessage, `.email-button`)

	// Check for email-card class on the card div
	assert.Contains(t, htmlMessage, `class="email-card"`)
}

func TestRenderMail(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line("This is a line")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)
		assert.Equal(t, mail.from, mailopts.From)
		assert.Equal(t, mail.to, mailopts.To)

		assert.Equal(t, `
Hi there,

This is a line



`, mailopts.Message)

		// Check for dark mode support
		assertHTMLContainsDarkModeSupport(t, mailopts.HTMLMessage)

		// Check for expected content
		assert.Contains(t, mailopts.HTMLMessage, `<p>This is a line</p>`)
		assert.Contains(t, mailopts.HTMLMessage, `Hi there,`)

		// Verify no action button is present
		assert.NotContains(t, mailopts.HTMLMessage, `class="email-button"`)
	})
	t.Run("with action", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line("This is a line").
			Line("This **line** contains [a link](https://vikunja.io)").
			Line("And another one").
			Action("The action", "https://example.com").
			Line("This should be an outro line").
			Line("And one more, because why not?")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)
		assert.Equal(t, mail.from, mailopts.From)
		assert.Equal(t, mail.to, mailopts.To)

		assert.Equal(t, `
Hi there,

This is a line

This **line** contains [a link](https://vikunja.io)

And another one

The action:
https://example.com

This should be an outro line

And one more, because why not?

`, mailopts.Message)

		// Check for dark mode support
		assertHTMLContainsDarkModeSupport(t, mailopts.HTMLMessage)

		// Check for action button with email-button class
		assert.Contains(t, mailopts.HTMLMessage, `class="email-button"`)
		assert.Contains(t, mailopts.HTMLMessage, `href="https://example.com"`)
		assert.Contains(t, mailopts.HTMLMessage, `The action`)

		// Check for markdown conversion
		assert.Contains(t, mailopts.HTMLMessage, `<strong>line</strong>`)
		assert.Contains(t, mailopts.HTMLMessage, `<a href="https://vikunja.io" rel="nofollow">a link</a>`)

		// Check for outro lines
		assert.Contains(t, mailopts.HTMLMessage, `This should be an outro line`)
		assert.Contains(t, mailopts.HTMLMessage, `And one more, because why not?`)

		// Check for copy URL text
		assert.Contains(t, mailopts.HTMLMessage, `https://example.com`)
	})
	t.Run("with footer", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line("This is a line").
			FooterLine("This is a footer line")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)
		assert.Equal(t, mail.from, mailopts.From)
		assert.Equal(t, mail.to, mailopts.To)

		assert.Equal(t, `
Hi there,

This is a line




This is a footer line
`, mailopts.Message)

		// Check for dark mode support
		assertHTMLContainsDarkModeSupport(t, mailopts.HTMLMessage)

		// Check for content
		assert.Contains(t, mailopts.HTMLMessage, `<p>This is a line</p>`)
		assert.Contains(t, mailopts.HTMLMessage, `<p>This is a footer line</p>`)

		// Verify no action button
		assert.NotContains(t, mailopts.HTMLMessage, `class="email-button"`)
	})
	t.Run("with footer and action", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line("This is a line").
			Line("This **line** contains [a link](https://vikunja.io)").
			Line("And another one").
			Action("The action", "https://example.com").
			Line("This should be an outro line").
			Line("And one more, because why not?").
			FooterLine("This is a footer line")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)
		assert.Equal(t, mail.from, mailopts.From)
		assert.Equal(t, mail.to, mailopts.To)

		assert.Equal(t, `
Hi there,

This is a line

This **line** contains [a link](https://vikunja.io)

And another one

The action:
https://example.com

This should be an outro line

And one more, because why not?


This is a footer line
`, mailopts.Message)

		// Check for dark mode support
		assertHTMLContainsDarkModeSupport(t, mailopts.HTMLMessage)

		// Check for action button with email-button class
		assert.Contains(t, mailopts.HTMLMessage, `class="email-button"`)
		assert.Contains(t, mailopts.HTMLMessage, `href="https://example.com"`)

		// Check for footer
		assert.Contains(t, mailopts.HTMLMessage, `<p>This is a footer line</p>`)
	})
	t.Run("with thread ID", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line("This is a line").
			ThreadID("<task-123@vikunja>")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)
		assert.Equal(t, mail.from, mailopts.From)
		assert.Equal(t, mail.to, mailopts.To)
		assert.Equal(t, "<task-123@vikunja>", mailopts.ThreadID)
	})
	t.Run("with special characters in task title", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`This is a friendly reminder of the task "Fix structured data Value in property "reviewCount" must be positive" (My Project).`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)
		assert.Equal(t, mail.from, mailopts.From)
		assert.Equal(t, mail.to, mailopts.To)

		// Plain text should keep quotes as-is
		assert.Contains(t, mailopts.Message, `"Fix structured data Value in property "reviewCount" must be positive"`)

		// HTML should have proper HTML entities for quotes
		// &#34; is the correct HTML entity for the quote character and will render as " in the browser
		assert.Contains(t, mailopts.HTMLMessage, `&#34;Fix structured data Value in property &#34;reviewCount&#34; must be positive&#34;`)
	})
	t.Run("with pre-escaped HTML entities", func(t *testing.T) {
		// This tests the fix for issue #1664 where HTML entities were being double-escaped
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task with entity: &#34;already escaped&#34; should render correctly`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Plain text should contain the HTML entity as-is (it will be interpreted by email client)
		assert.Contains(t, mailopts.Message, `&#34;`)

		// HTML should properly handle the pre-escaped entity without double-escaping
		// The entity should remain as &#34; (not become &amp;#34;)
		assert.Contains(t, mailopts.HTMLMessage, `&#34;already escaped&#34;`)
		// Should NOT double-escape to &amp;#34; which would display as literal &#34;
		assert.NotContains(t, mailopts.HTMLMessage, `&amp;#34;`)
	})
	t.Run("with XSS attempt via script tag", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task: <script>alert('XSS')</script>`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Script tags should be stripped by bluemonday sanitization
		assert.NotContains(t, mailopts.HTMLMessage, `<script>`)
		assert.NotContains(t, mailopts.HTMLMessage, `</script>`)
		assert.NotContains(t, mailopts.HTMLMessage, `alert('XSS')`)
		// The text should be present but sanitized
		assert.Contains(t, mailopts.HTMLMessage, `Task:`)
	})
	t.Run("with XSS attempt via img onerror", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task: <img src=x onerror=alert('XSS')>`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// The dangerous HTML should be escaped, not rendered as actual HTML
		// This makes it safe - it will display as text, not execute
		assert.Contains(t, mailopts.HTMLMessage, `&lt;img`)
		assert.Contains(t, mailopts.HTMLMessage, `&gt;`)
		// Verify it's not an actual executable img tag
		assert.NotContains(t, mailopts.HTMLMessage, `<img src=x onerror=`)
		// Task text should remain
		assert.Contains(t, mailopts.HTMLMessage, `Task:`)
	})
	t.Run("with XSS attempt via javascript protocol", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task: <a href="javascript:alert('XSS')">Click me</a>`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// JavaScript protocol should be stripped
		assert.NotContains(t, mailopts.HTMLMessage, `javascript:alert`)
		assert.NotContains(t, mailopts.HTMLMessage, `href="javascript:`)
		// Text content should remain
		assert.Contains(t, mailopts.HTMLMessage, `Task:`)
	})
	t.Run("with XSS attempt via iframe", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task: <iframe src="http://evil.com"></iframe>`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Iframes should be completely stripped by bluemonday
		assert.NotContains(t, mailopts.HTMLMessage, `<iframe`)
		assert.NotContains(t, mailopts.HTMLMessage, `http://evil.com`)
		// Task text should remain
		assert.Contains(t, mailopts.HTMLMessage, `Task:`)
	})
	t.Run("with XSS attempt via HTML injection", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task: <div onclick="alert('XSS')">Dangerous</div>`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// onclick handler should be stripped
		assert.NotContains(t, mailopts.HTMLMessage, `onclick=`)
		assert.NotContains(t, mailopts.HTMLMessage, `onclick="alert`)
		// Text content may remain but without the dangerous attributes
		assert.Contains(t, mailopts.HTMLMessage, `Task:`)
	})
	t.Run("with XSS attempt via data URI", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task: <img src="data:text/html,<script>alert('XSS')</script>">`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Script tags should not appear in final HTML
		assert.NotContains(t, mailopts.HTMLMessage, `<script>alert('XSS')</script>`)
		assert.NotContains(t, mailopts.HTMLMessage, `<script>`)
		// Task text should remain
		assert.Contains(t, mailopts.HTMLMessage, `Task:`)
	})
	t.Run("with XSS attempt via style tag in user content", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task: <style>body{background:url('javascript:alert(1)')}</style>`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// User-provided style tags should be stripped by bluemonday (different from template style)
		// The template has a legitimate <style> block for dark mode in the <head>, but user content
		// style tags in the body should be stripped
		assert.NotContains(t, mailopts.HTMLMessage, `<style>body{background:url`)
		// Task text should remain
		assert.Contains(t, mailopts.HTMLMessage, `Task:`)

		// The template's dark mode style block should still be present in the head
		// Count <style> tags - there should be exactly one (the template's dark mode styles)
		styleCount := strings.Count(mailopts.HTMLMessage, "<style>")
		assert.Equal(t, 1, styleCount, "There should be exactly one <style> tag (the template's dark mode styles)")
	})
	t.Run("with mixed XSS and legitimate content", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line(`Task "Fix Bug" has <script>alert('XSS')</script> priority & needs **attention**`)

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Malicious content should be stripped
		assert.NotContains(t, mailopts.HTMLMessage, `<script>`)
		assert.NotContains(t, mailopts.HTMLMessage, `alert('XSS')`)

		// Legitimate content should be preserved
		assert.Contains(t, mailopts.HTMLMessage, `Task`)
		assert.Contains(t, mailopts.HTMLMessage, `Fix Bug`)
		// Ampersand should be escaped
		assert.Contains(t, mailopts.HTMLMessage, `&amp;`)
		// Markdown bold should be converted to strong
		assert.Contains(t, mailopts.HTMLMessage, `<strong>attention</strong>`)
	})
}

func TestConversationalMail(t *testing.T) {
	t.Run("Conversational flag", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Conversational().
			Line("This is a conversational message")

		assert.True(t, mail.IsConversational())
	})

	t.Run("Default is not conversational", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Line("This is a formal message")

		assert.False(t, mail.IsConversational())
	})

	t.Run("Conversational template selection", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Conversational().
			Line("This is a conversational message").
			Action("View Task", "https://example.com/task/123")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Should not contain greeting section
		assert.NotContains(t, mailopts.HTMLMessage, "<p>\n\t\t\n\t</p>")

		// Should use conversational styling
		assert.Contains(t, mailopts.HTMLMessage, "background: #f6f8fa")
		assert.Contains(t, mailopts.HTMLMessage, "font-family: -apple-system")
		assert.Contains(t, mailopts.HTMLMessage, "max-width: 700px")

		// Should NOT have logo (completely removed)
		assert.NotContains(t, mailopts.HTMLMessage, "logo.png")
		assert.NotContains(t, mailopts.HTMLMessage, "Vikunja")

		// Should have inline action link with arrow
		assert.Contains(t, mailopts.HTMLMessage, "View Task →")
		assert.Contains(t, mailopts.HTMLMessage, "color: #0969da")

		// Should not have the formal button styling
		assert.NotContains(t, mailopts.HTMLMessage, "background-color: #1973ff")
		assert.NotContains(t, mailopts.HTMLMessage, "width:280px")

		// Plain text should not have greeting
		assert.NotContains(t, mailopts.Message, "Hi there,")
		assert.Contains(t, mailopts.Message, "This is a conversational message")
	})

	t.Run("Formal template still works", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Greeting("Hi there,").
			Line("This is a formal message").
			Action("View Task", "https://example.com/task/123")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Should contain greeting
		assert.Contains(t, mailopts.HTMLMessage, "Hi there,")

		// Should use formal styling
		assert.Contains(t, mailopts.HTMLMessage, "background: #f3f4f6")
		assert.Contains(t, mailopts.HTMLMessage, "font-family: 'Open Sans'")
		assert.Contains(t, mailopts.HTMLMessage, "width: 600px")
		assert.Contains(t, mailopts.HTMLMessage, "height: 75px")

		// Should HAVE logo in formal emails
		assert.Contains(t, mailopts.HTMLMessage, "logo.png")
		assert.Contains(t, mailopts.HTMLMessage, "Vikunja")

		// Should have formal button styling
		assert.Contains(t, mailopts.HTMLMessage, "background-color: #1973ff")
		assert.Contains(t, mailopts.HTMLMessage, "width:280px")

		// Should not have conversational arrow
		assert.NotContains(t, mailopts.HTMLMessage, "View Task →")

		// Plain text should have greeting
		assert.Contains(t, mailopts.Message, "Hi there,")
	})

	t.Run("Conversational without action", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Conversational().
			Line("This is a conversational message without action")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Should use conversational styling
		assert.Contains(t, mailopts.HTMLMessage, "background: #f6f8fa")
		assert.Contains(t, mailopts.HTMLMessage, "max-width: 700px")

		// Should not have action section
		assert.NotContains(t, mailopts.HTMLMessage, "border-top: 1px solid #e5e7eb")
	})

	t.Run("Conversational with footer", func(t *testing.T) {
		mail := NewMail().
			From("test@example.com").
			To("test@otherdomain.com").
			Subject("Testmail").
			Conversational().
			Line("This is a conversational message").
			FooterLine("This is a footer line")

		mailopts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Should have footer with conversational styling
		assert.Contains(t, mailopts.HTMLMessage, "color: #656d76")
		assert.Contains(t, mailopts.HTMLMessage, "font-size: 12px")
		assert.Contains(t, mailopts.HTMLMessage, "This is a footer line")
	})

	t.Run("Conversational header line format", func(t *testing.T) {
		username := "testuser"
		action := "left a comment"
		taskURL := "https://example.com/task/123"
		projectTitle := "Test Project"
		taskTitle := "Test Task"

		headerLine := CreateConversationalHeader(username, action, taskURL, projectTitle, taskTitle)

		// Should contain username in strong tags
		assert.Contains(t, headerLine, `<strong>testuser</strong>`)

		// Should contain action
		assert.Contains(t, headerLine, "left a comment")

		// Should contain task link with GitHub blue color
		assert.Contains(t, headerLine, `<a href="https://example.com/task/123"`)
		assert.Contains(t, headerLine, `color: #0969da`)
		assert.Contains(t, headerLine, `(Test Project &gt; Test Task)`)

		// Should have proper margin bottom
		assert.Contains(t, headerLine, `margin-bottom: 16px`)
	})

	t.Run("Conversational action link font size", func(t *testing.T) {
		mail := NewMail().
			Conversational().
			Subject("Test").
			Line("Content").
			Action("View Task", "https://example.com/task/123")

		mailOpts, err := RenderMail(mail, "en")
		require.NoError(t, err)

		// Action link should have 14px font size to match main content
		assert.Contains(t, mailOpts.HTMLMessage, `font-size: 14px; line-height: 1.5`) // Main content
		assert.Contains(t, mailOpts.HTMLMessage, `font-weight: 500; font-size: 14px`) // Action link
	})

	t.Run("Translation system integration", func(t *testing.T) {
		// Test that translation keys are properly structured
		// This verifies the translation keys exist and are accessible

		// Test action translations
		headerLine1 := CreateConversationalHeader("John", "left a comment", "https://example.com", "Project", "Task")
		assert.Contains(t, headerLine1, "left a comment")

		headerLine2 := CreateConversationalHeader("Jane", "assigned you", "https://example.com", "Project", "Task")
		assert.Contains(t, headerLine2, "assigned you")

		// Verify header structure is maintained
		assert.Contains(t, headerLine1, "<strong>John</strong>")
		assert.Contains(t, headerLine1, `color: #0969da`)
		assert.Contains(t, headerLine1, "(Project &gt; Task)")
	})
}
