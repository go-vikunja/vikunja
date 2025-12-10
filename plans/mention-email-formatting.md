# Plan: Format User Mentions in Email Notifications

## Problem Statement

When users are mentioned in task descriptions or comments using the `@username` mention feature in the frontend, they receive email notifications. However, the emails contain raw HTML tags like `<mention-user data-id="username">@username</mention-user>` or just `@username` text after sanitization, rather than the user's full display name (e.g., "John Doe" instead of just "@johndoe").

This creates a poor user experience where:
- Emails show technical usernames instead of human-readable names
- The mention appears less personal and harder to read
- The display doesn't match what users see in the frontend

## Current State

### Actual Mention HTML Format
Based on the frontend implementation and real-world usage:

**Current format** (what's actually stored):
```html
<mention-user data-id="konrad" data-label="Konrad" data-mention-suggestion-char="@"></mention-user>
```

**Note**: The mention tag is **self-closing** with NO text content inside. All the display information is in the attributes:
- `data-id`: The username (used for lookup)
- `data-label`: The display name (what should be shown - could be full name or username)
- `data-mention-suggestion-char`: The trigger character (always `@`)

**Old format** (tests still use this):
```html
<mention-user data-id="username">@username</mention-user>
```

### Data Flow
1. **Frontend**: User types `@` and selects a user from dropdown
   - `mentionSuggestion.ts` creates mention with `id: user.username`, `label: getDisplayName(user)`
   - TipTap extension renders it as `<mention-user data-id="..." data-label="..." data-mention-suggestion-char="@">`
2. **Storage**: Saved as self-closing tag with attributes in database
3. **Email Generation**: HTML content passed through bluemonday sanitizer
4. **Email Display**: bluemonday strips the unknown `<mention-user>` tag completely, leaving **nothing** or empty space

### Key Files
- `pkg/models/mentions.go` - Extracts mentioned usernames from HTML using `data-id` attribute
- `pkg/models/notifications.go` - Email notification types (TaskCommentNotification, UserMentionedInTaskNotification)
- `pkg/notifications/mail_render.go` - HTML sanitization and email rendering
- `pkg/user/user.go` - User model with `GetName()` method (returns Name field or Username as fallback)
- `frontend/src/components/input/editor/mention/mentionSuggestion.ts` - Creates mention with id/label
- `frontend/src/components/input/editor/TipTap.vue` - Extends Mention to use custom tag

### Current Limitations
- `extractMentionedUsernames()` only extracts usernames for notification lookup
- `FindMentionedUsersInText()` returns user objects but doesn't modify the HTML
- `ToMail()` methods pass raw HTML/description to email without mention formatting
- bluemonday sanitizer strips `<mention-user>` tags completely (self-closing tags with no content = invisible in email)
- The `data-label` attribute already contains the correct display name but is lost during sanitization

## Solution Design

### Approach
Create a new function `formatMentionsForEmail()` that:
1. Parses HTML content to find `<mention-user>` tags
2. Extracts the `data-label` attribute (which already contains the correct display name)
3. Replaces self-closing mention tags with human-readable formatted text: `<strong>@{data-label}</strong>`
4. Returns formatted HTML ready for email rendering

**Key Insight**: We DON'T need database lookups! The `data-label` attribute already has the display name we need. This is simpler and more efficient than the original approach.

### Implementation Steps

#### 1. Create Mention Formatting Function
**File**: `pkg/models/mentions.go`

Add new function:
```go
// formatMentionsForEmail replaces mention-user tags with human-readable user names.
// Input:  <mention-user data-id="johndoe" data-label="John Doe" data-mention-suggestion-char="@"></mention-user>
// Output: <strong>@John Doe</strong>
// If data-label is missing, falls back to data-id
func formatMentionsForEmail(htmlText string) (string, error)
```

**Implementation Details**:
- Parse HTML using `golang.org/x/net/html` (already used in `extractMentionedUsernames`)
- Find all `<mention-user>` element nodes (note: they're self-closing, no children)
- For each mention node:
  - Extract `data-label` attribute (preferred - contains display name)
  - If `data-label` is empty, fallback to `data-id` attribute
  - Create a new text node with content: `@{label}` (just plain text with @)
  - Create a `<strong>` element wrapping the text node
  - Replace the `<mention-user>` node with the `<strong>` node in the tree
- Render modified HTML tree back to string using `html.Render()`
- Return formatted HTML

**Why `<strong>` tag?**
- Simple, semantic HTML supported by all email clients
- Already allowed by bluemonday UGCPolicy
- Provides visual emphasis similar to frontend rendering
- Alternative considered: `<span style="font-weight:bold">` - not needed, strong is cleaner

**Note on signature**: No `*xorm.Session` parameter needed since we're reading from attributes, not database

#### 2. Update TaskCommentNotification.ToMail()
**File**: `pkg/models/notifications.go` (around line 95-107)

**Current Code**:
```go
func (n *TaskCommentNotification) ToMail(lang string) *notifications.Mail {
	// ...
	mail.HTML(n.Comment.Comment)  // Line 107
	return mail.Action(...)
}
```

**New Code**:
```go
func (n *TaskCommentNotification) ToMail(lang string) *notifications.Mail {
	formattedComment, err := formatMentionsForEmail(n.Comment.Comment)
	if err != nil {
		// Log error but continue with original comment
		log.Errorf("Failed to format mentions in comment %d: %v", n.Comment.ID, err)
		formattedComment = n.Comment.Comment
	}

	mail := notifications.NewMail().
		From(n.Doer.GetNameAndFromEmail()).
		Subject(subject).
		Line(i18n.T(lang, "notifications.task.comment.message", n.Doer.GetName())).
		HTML(formattedComment)  // Use formatted comment

	return mail.Action(...)
}
```

**Note**: No database session needed since we're just parsing HTML attributes

#### 3. Update UserMentionedInTaskNotification.ToMail()
**File**: `pkg/models/notifications.go` (around line 352-368)

**Current Code**:
```go
func (n *UserMentionedInTaskNotification) ToMail(lang string) *notifications.Mail {
	// ...
	mail.HTML(n.Task.Description)  // Line 364
	return mail.Action(...)
}
```

**New Code**:
```go
func (n *UserMentionedInTaskNotification) ToMail(lang string) *notifications.Mail {
	formattedDescription, err := formatMentionsForEmail(n.Task.Description)
	if err != nil {
		log.Errorf("Failed to format mentions in task %d: %v", n.Task.ID, err)
		formattedDescription = n.Task.Description
	}

	mail := notifications.NewMail().
		From(n.Doer.GetNameAndFromEmail()).
		Subject(subject).
		Line(i18n.T(lang, "notifications.task.mentioned.message", n.Doer.GetName())).
		HTML(formattedDescription)  // Use formatted description

	return mail.Action(...)
}
```

**Note**: No database session needed

#### 4. Add Tests
**File**: `pkg/models/mentions_test.go`

Add test cases:
- `TestFormatMentionsForEmail` - Basic formatting scenarios:
  - Single mention with data-label: `<mention-user data-id="konrad" data-label="Konrad" data-mention-suggestion-char="@"></mention-user>` → `<strong>@Konrad</strong>`
  - Multiple mentions in one paragraph
  - Mention with data-label that's a full name: `data-label="John Doe"`
  - Mention without data-label (fallback to data-id): `<mention-user data-id="johndoe"></mention-user>` → `<strong>@johndoe</strong>`
  - Old format with text node inside (backward compatibility): `<mention-user data-id="user1">@user1</mention-user>` → `<strong>@user1</strong>`
  - HTML preservation: verify paragraphs, links, etc. remain intact around mentions
  - Empty/nil input returns unchanged
  - HTML without mentions returns unchanged
  - Self-closing vs. non-self-closing tags both work
  - Special characters in data-label are preserved
  - Mixed old and new format in same HTML

#### 5. Integration Testing
**Manual Testing**:
1. Create test users with full names and username-only accounts
2. Create task with mentions in description
3. Add comment with mentions
4. Verify emails show formatted names
5. Test with multiple mentions of same/different users
6. Verify email clients render correctly (Gmail, Outlook, etc.)

## Edge Cases to Handle

1. **Missing data-label attribute**: Fallback to `data-id` attribute
2. **Missing both data-label and data-id**: Use empty string or skip the mention (log warning)
3. **Permission check**: Already handled - only users with read access are notified
4. **HTML injection**: NOT a concern - we're reading from existing data-label attribute which comes from controlled frontend; html.Render() will properly escape content anyway
5. **Empty/malformed HTML**: Return input unchanged if parsing fails
6. **Old format compatibility**: Handle both `<mention-user data-id="x">@x</mention-user>` (with text content) and new self-closing format
7. **Self-closing vs. element nodes**: HTML parser might parse self-closing differently - handle both cases
8. **Special characters in labels**: Characters like `<`, `>`, `&` in data-label will be preserved through attribute parsing and properly escaped in text node creation
9. **Very long names**: Display as-is (trust frontend validation)

## Error Handling Strategy

- Log errors but never fail email sending due to mention formatting
- On any error in `formatMentionsForEmail()`, return original HTML unchanged
- Errors to handle:
  - HTML parsing errors
  - Database query errors
  - Invalid HTML structure

## Alternative Approaches Considered

### Alternative 1: Look up users from database by data-id
**Rejected**: Unnecessary complexity. The data-label attribute already contains what we need. Why make extra database queries?

### Alternative 2: Frontend-side formatting before save
**Rejected**: Would require changing stored data format, migration of existing data, and changes to frontend editor

### Alternative 3: Replace with plain text "@Full Name" (no strong tag)
**Rejected**: Loses emphasis styling, harder to distinguish mentions from regular text in emails

### Alternative 4: Use custom email template with mention styling
**Rejected**: More complex, requires template changes, harder to maintain

### Alternative 5: Process mentions in mail_render.go during sanitization
**Rejected**: Separates mention logic from existing mention handling, harder to test, would need to hook into bluemonday sanitizer

## Implementation Checklist

- [ ] Implement `formatMentionsForEmail()` in `pkg/models/mentions.go`
- [ ] Update `TaskCommentNotification.ToMail()` to use formatted mentions
- [ ] Update `UserMentionedInTaskNotification.ToMail()` to use formatted mentions
- [ ] Write unit tests for mention formatting function
- [ ] Test with various user name scenarios
- [ ] Test HTML preservation and escaping
- [ ] Manual testing with real email clients
- [ ] Update translation strings if needed (unlikely)

## Rollout Plan

1. Implement and test in development
2. Run existing test suite to ensure no regressions: `mage test`
3. Run lint checks: `mage lint`
4. Test email formatting manually with test users
5. Commit changes following conventional commits format
6. No frontend changes needed
7. No database migration needed
8. No breaking changes - graceful enhancement

## Success Criteria

- ✓ Email notifications show user's display name from data-label
- ✓ Mentions appear as "@John Doe" or "@Konrad" (whatever is in data-label)
- ✓ Mentions appear in bold (using `<strong>` tag)
- ✓ HTML formatting is preserved in emails (paragraphs, links, etc.)
- ✓ Backward compatible with old mention format (with text nodes)
- ✓ No errors in email sending
- ✓ All existing tests pass
- ✓ New tests cover mention formatting scenarios including new self-closing format
- ✓ Works across different email clients

## Future Enhancements (Out of Scope)

- Add avatar images to mentions in emails
- Different styling for mentions (background color, etc.)
- Support for team/group mentions
- Click-through links from mention to user profile
