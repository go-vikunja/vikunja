# Conversational Email Notifications Design

**Date:** 2025-11-27
**Status:** Design
**Goal:** Change email notification styling from formal centered notifications to conversational left-aligned emails (like GitHub issue comments) for interactive notifications only.

---

## Overview

Currently, Vikunja's email notifications use a formal, centered layout with a large logo, boxed content, and prominent blue action button. This design feels like a "system notification" rather than a natural email conversation.

This design introduces a conversational email template for interactive notifications (comments, mentions, assignments) while keeping the formal style for system notifications (reminders, overdue alerts, exports).

### Design Inspiration

This design is directly inspired by GitHub's notification emails, which use:
- **No app logo** - completely minimal
- **User avatar** (20x20px, circular) next to username
- GitHub-style header line: `{Avatar} {Username} {action} ({Repository}#{Issue})`
- Simple, clean layout with system fonts
- Minimal borders and padding
- GitHub blue links (`#0969da`)
- Very simple footer with text links

For Vikunja, we adapt this format to:
- **No Vikunja logo** - just content
- **User avatar** inline with username in header
- Header line: `{Avatar} {Username} {action} ({Project} > {Task})`
- Use project + task title instead of repo#issue
- Maintain the same minimal, conversational aesthetic
- Keep system fonts and GitHub's color palette

---

## Architecture & Template Structure

### Dual Template System

Create a second email template alongside the existing one:
- **Keep** `mailTemplateHTML` for formal/system notifications
- **Add** `mailTemplateConversationalHTML` for interactive notifications
- Each notification's `ToMail()` method chooses which template to use
- Templates share common rendering logic but have different layouts

**Benefits:**
- Clean separation between styles
- Easy to maintain both independently
- Explicit template selection per notification
- No breaking changes to existing notifications

### Template Selection Logic

Modify `RenderMail()` in `pkg/notifications/mail_render.go` to select template based on a flag:

```go
func RenderMail(m *Mail, lang string) (*mail.Opts, error) {
	var htmlTemplate string
	if m.conversational {
		htmlTemplate = mailTemplateConversationalHTML
	} else {
		htmlTemplate = mailTemplateHTML
	}
	// ... rest of rendering logic
}
```

---

## Conversational Template Design

### HTML Structure

Inspired by GitHub's conversational email style with minimal formatting, no app logo, and user avatar inline.

```html
<!doctype html>
<html style="width: 100%; height: 100%; padding: 0; margin: 0;">
<head>
    <meta name="viewport" content="width: display-width;">
    <meta charset="utf-8">
</head>
<body style="width: 100%; padding: 0; margin: 0; background: #f6f8fa; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;">
<div style="max-width: 700px; margin: 0 auto; background: #ffffff;">

    <!-- Conversational content with header line (no logo) -->
    <div style="padding: 20px; color: #24292f; font-size: 14px; line-height: 1.5;">
        {{ range $line := .IntroLinesHTML}}
            {{ $line }}
        {{ end }}

        {{ range $line := .OutroLinesHTML}}
            {{ $line }}
        {{ end }}
    </div>

    <!-- Inline action link instead of button -->
    {{ if .ActionURL }}
    <div style="padding: 0 20px 20px 20px; border-top: 1px solid #d1d9e0; margin-top: 16px; padding-top: 16px;">
        <a href="{{ .ActionURL }}" style="color: #0969da; text-decoration: none; font-weight: 500;">
            {{ .ActionText }} →
        </a>
    </div>
    {{ end }}

    <!-- Footer -->
    {{ if .FooterLinesHTML }}
    <div style="padding: 16px 20px 20px 20px; border-top: 1px solid #d1d9e0; color: #656d76; font-size: 12px;">
        {{ range $line := .FooterLinesHTML }}
            {{ $line }}
        {{ end }}
    </div>
    {{ end }}
</div>
</body>
</html>
```

### Key Styling Differences

| Element | Formal Style | Conversational Style |
|---------|--------------|---------------------|
| **Layout** | Centered 600px box | Max-width 700px container |
| **App Logo** | 75px centered banner | **None** - completely removed |
| **User Avatar** | None | 20x20px circular, inline with username |
| **Font** | Open Sans | System fonts (-apple-system, Segoe UI, etc.) |
| **Font size** | Larger (implicit) | 14px base |
| **Container** | Border + drop shadow | Clean white background with padding |
| **Action** | Large blue button | Inline text link with arrow (GitHub blue `#0969da`) |
| **Borders** | `#dbdbdb` box shadow | Subtle `#d1d9e0` dividers |
| **Background** | `#f3f4f6` | `#f6f8fa` (GitHub's background color) |
| **Greeting** | Included | Omitted (flows naturally) |
| **Header Line** | None | "{Avatar} {Username} {action} ({Project Task})" format |
| **Text Color** | `#4a4a4a` | `#24292f` (GitHub's text color) |

### Notification Content Format

Conversational notifications should use a GitHub-style header line with avatar as the first intro line.

**Format**: `{Avatar} {Username} {action description} ({Project} > {Task title})`

**Examples**:
- `[avatar] konrad left a comment (Project Management > Update documentation)`
- `[avatar] jane mentioned you (Bug Fixes > Fix login issue)`
- `[avatar] bob assigned you (Feature Requests > Add dark mode)`

**Implementation Details**:

The header line is generated as HTML and added via `mail.HTML()`:
```html
<div style="margin-bottom: 16px;">
    <img src="{avatar_url}" height="20" width="20" style="border-radius: 50%; margin-right: 4px; vertical-align: middle;" alt="{Username}"/>
    <strong>{Username}</strong> {action description}
    <a href="{task_url}" style="color: #0969da; text-decoration: none;">
        ({Project} &gt; {Task})
    </a>
</div>
```

Key points:
- Avatar: 20x20px, circular (`border-radius: 50%`), small right margin
- Use `vertical-align: middle` to align avatar with text
- Avatar URL from user's avatar provider (Gravatar, uploaded, or initials)
- Use `<strong>` for username (matches GitHub)
- Link to task uses GitHub blue (`#0969da`)
- HTML entities (`&gt;`) for proper rendering
- Bottom margin for spacing before content
- Task link is clickable for easy navigation

---

## Mail Struct Changes

### Extend Mail Struct

Add conversational flag to `pkg/notifications/mail.go`:

```go
type Mail struct {
	from           string
	to             string
	subject        string
	actionText     string
	actionURL      string
	greeting       string
	introLines     []*mailLine
	outroLines     []*mailLine
	footerLines    []*mailLine
	threadID       string
	conversational bool  // NEW: template selection flag
}
```

### Add Conversational Method

```go
// Conversational sets the email to use conversational styling
func (m *Mail) Conversational() *Mail {
	m.conversational = true
	return m
}
```

### Usage Example

```go
func (n *TaskCommentNotification) ToMail(lang string) *notifications.Mail {
	mail := notifications.NewMail().
		Conversational().  // Enable conversational style
		From(n.Doer.GetNameAndFromEmail()).
		Subject(i18n.T(lang, "notifications.task.comment.subject", n.Task.Title))

	// Add GitHub-style header line with avatar, username, action, and task reference
	avatarURL := n.Doer.GetAvatarURL()  // Get user's avatar from avatar provider
	headerLine := fmt.Sprintf(
		`<div style="margin-bottom: 16px;"><img src="%s" height="20" width="20" style="border-radius: 50%%; margin-right: 4px; vertical-align: middle;" alt="%s"/><strong>%s</strong> left a comment <a href="%s" style="color: #0969da; text-decoration: none;">(%s &gt; %s)</a></div>`,
		avatarURL,
		n.Doer.GetName(),
		n.Doer.GetName(),
		n.Task.GetFrontendURL(),
		n.Task.Project.Title,
		n.Task.Title,
	)
	mail.HTML(headerLine)

	// Add the actual comment content
	mail.Line(n.Comment.Comment)

	// Add action link to view in Vikunja
	mail.Action(
		i18n.T(lang, "notifications.task.comment.action"),
		n.Task.GetFrontendURL(),
	)

	return mail
}
```

---

## Notification Type Mapping

### Interactive Notifications (Conversational Style)

These notifications call `.Conversational()` to use the new template:

1. **TaskCommentNotification** - Comments are inherently conversational
2. **UserMentionedInTaskNotification** - Direct mention/address
3. **TaskAssignedNotification** - Collaborative task handoff

### System Notifications (Formal Style)

These continue using the existing formal template:

1. **ReminderDueNotification** - Time-based system reminder
2. **UndoneTaskOverdueNotification** - Automated overdue alert
3. **UndoneTasksOverdueNotification** - Bulk overdue digest
4. **TaskDeletedNotification** - System event notification
5. **DataExportReadyNotification** - System process completion
6. **TeamMemberAddedNotification** - Administrative action
7. **ProjectCreatedNotification** - Administrative action

---

## Implementation Details

### Files to Modify

1. **`pkg/notifications/mail_render.go`**:
   - Add `mailTemplateConversationalHTML` constant (without logo)
   - Update `RenderMail()` to select template based on flag
   - Optionally add conversational plain text variant
   - Consider conditional logo embedding (only for non-conversational emails)

2. **`pkg/notifications/mail.go`**:
   - Add `conversational bool` field to Mail struct
   - Add `Conversational()` method

3. **`pkg/models/notifications.go`**:
   - Update `TaskCommentNotification.ToMail()`:
     - Call `.Conversational()`
     - Get avatar URL from `n.Doer.GetAvatarURL()`
     - Add GitHub-style header line with avatar: `{Avatar} {Username} left a comment ({Project} > {Task})`
     - Include comment content as main body
   - Update `UserMentionedInTaskNotification.ToMail()`:
     - Call `.Conversational()`
     - Get avatar URL from doer/mentioner
     - Add header line with avatar: `{Avatar} {Username} mentioned you ({Project} > {Task})`
     - Include context of where the mention occurred
   - Update `TaskAssignedNotification.ToMail()`:
     - Call `.Conversational()`
     - Get avatar URL from assigner
     - Add header line with avatar: `{Avatar} {Username} assigned you ({Project} > {Task})`
     - Include task details

### Plain Text Template

The plain text template should also skip greeting for conversational emails, but otherwise minimal changes needed since plain text is already conversational by nature.

Plain text header line format (avatar cannot be displayed in plain text):
```
{Username} {action} ({Project} > {Task})
```

Note: Avatar is omitted in plain text since it's an image.

### Avatar URL Generation

Vikunja already has an avatar system that supports multiple providers:
- Gravatar (based on email hash)
- Uploaded avatars
- Initials/default avatars

The `User.GetAvatarURL()` method (or similar) should return the full HTTP(S) URL to the user's avatar image. This URL will be embedded directly in the email HTML.

**Important considerations**:
- Avatar URLs must be publicly accessible (external URLs work in emails)
- Gravatar URLs work well in emails since they're external
- For uploaded avatars, ensure the URL is publicly accessible or consider fallback to Gravatar
- Initials-based avatars may need to be served as external URLs if generated server-side

### Footer Format

Conversational emails should use a simple, GitHub-style footer with plain text links:

```
Reply to this email directly or view it on Vikunja:
{task_url}

Change your notification settings:
{settings_url}
```

This matches GitHub's minimalist footer approach rather than the more formal footer in system notifications.

### Backward Compatibility

- ✅ All existing notifications default to formal style (`conversational` defaults to `false`)
- ✅ No database changes required
- ✅ No configuration changes required
- ✅ Users see changes immediately on next notification
- ✅ No breaking changes to notification interface

### i18n Considerations

- ✅ No translation string changes needed
- ✅ Styling is presentation-only
- ✅ Existing `i18n.T()` calls work as-is

---

## Testing Strategy

### Unit Tests

Add to `pkg/notifications/mail_test.go`:
- Test conversational flag properly selects template
- Verify both templates render without errors
- Confirm conversational emails omit greeting
- Test plain text variants (without avatar)
- Verify header line format renders correctly with HTML entities
- Test avatar rendering (20x20px, circular, with proper margins)
- Test that action links use correct color (`#0969da`)
- Confirm footer uses simple text link format
- Test avatar URL generation for different avatar providers (Gravatar, uploaded, initials)

### Visual Testing

Send test emails for:
- Each interactive notification type (comment, mention, assignment)
  - Verify avatar displays correctly: 20x20px, circular, aligned with text
  - Check avatar loading from different providers (Gravatar, uploaded)
  - Verify header line renders correctly: `{Avatar} {Username} {action} ({Project} > {Task})`
  - Check that links use GitHub blue color (`#0969da`)
  - Confirm minimal styling matches GitHub's aesthetic
  - Test with users who have different avatar types
- Each system notification type (verify no changes, no avatars)
- Test in multiple email clients:
  - Gmail (web + mobile) - verify avatar loads
  - Outlook (desktop + web) - check image rendering
  - Apple Mail (macOS + iOS) - confirm circular rendering
  - Thunderbird - test avatar display
- Compare side-by-side with actual GitHub notification emails

### Edge Cases

- **Email clients that strip styles**: HTML structure ensures readability
- **Avatar loading failures**: Test with broken avatar URLs, ensure graceful fallback
- **Avatar URLs blocked**: Some email clients block external images by default
- **Circular avatars**: Test `border-radius: 50%` support across email clients
- **Dark mode**: Semantic colors adapt in modern email clients
- **Mobile rendering**: Max-width approach ensures responsive design, avatar scales properly
- **Long content**: Test with long task titles, comments, etc.
- **Long usernames**: Test avatar alignment with long usernames

---

## Success Criteria

1. Interactive notifications (comments, mentions, assignments) use conversational template
2. User avatars display correctly in header line (20x20px, circular)
3. No Vikunja logo in conversational emails (completely removed)
4. Header line format matches GitHub: `{Avatar} {Username} {action} ({Project} > {Task})`
5. System notifications continue using formal template unchanged (with logo)
6. All email clients render conversational template correctly
7. Avatars load from all avatar providers (Gravatar, uploaded, initials)
8. Plain text emails remain readable (without avatar)
9. No regressions in existing notification functionality
10. Email threading (ThreadID) continues working

---

## Future Enhancements

Potential future improvements (out of scope for this design):

- User preference to choose formal vs conversational style
- Reply-by-email functionality for comments
- Show previous comments in thread context
- Rich task metadata display (labels, due dates, etc.)
- Inline comment preview/snippet in notification
