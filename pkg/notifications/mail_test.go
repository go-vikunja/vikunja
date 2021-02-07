// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, "This is a line", mail.introLines[0])
		assert.Equal(t, "And another one", mail.introLines[1])
		assert.Len(t, mail.outroLines, 2)
		assert.Equal(t, "This should be an outro line", mail.outroLines[0])
		assert.Equal(t, "And one more, because why not?", mail.outroLines[1])
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
		assert.Equal(t, "Hi,", mail.greeting) // Default greeting
		assert.Len(t, mail.introLines, 2)
		assert.Equal(t, "This is a line", mail.introLines[0])
		assert.Equal(t, "And another one", mail.introLines[1])
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
		assert.Equal(t, "This is a line", mail.introLines[0])
		assert.Equal(t, "And another one", mail.introLines[1])
		assert.Equal(t, "This should be an outro line", mail.introLines[2])
		assert.Equal(t, "And one more, because why not?", mail.introLines[3])
	})
}

func TestRenderMail(t *testing.T) {
	mail := NewMail().
		From("test@example.com").
		To("test@otherdomain.com").
		Subject("Testmail").
		Greeting("Hi there,").
		Line("This is a line").
		Line("And another one").
		Action("The action", "https://example.com").
		Line("This should be an outro line").
		Line("And one more, because why not?")

	mailopts, err := RenderMail(mail)
	assert.NoError(t, err)
	assert.Equal(t, mail.from, mailopts.From)
	assert.Equal(t, mail.to, mailopts.To)

	assert.Equal(t, `
Hi there,

This is a line

And another one

The action:
https://example.com

This should be an outro line

And one more, because why not?
`, mailopts.Message)
	assert.Equal(t, `
<!doctype html>
<html style="width: 100%; height: 100%; padding: 0; margin: 0;">
<head>
    <meta name="viewport" content="width: display-width;">
</head>
<body style="width: 100%; padding: 0; margin: 0; background: #f3f4f6">
<div style="width: 100%; font-family: 'Open Sans', sans-serif; text-rendering: optimizeLegibility">
    <div style="width: 600px; margin: 0 auto; text-align: justify;">
        <h1 style="font-size: 30px; text-align: center;">
            <img src="images/logo-full.svg" style="height: 75px;" alt="Vikunja"/>
        </h1>
        <div style="border: 1px solid #dbdbdb; -webkit-box-shadow: 0.3em 0.3em 0.8em #e6e6e6; box-shadow: 0.3em 0.3em 0.8em #e6e6e6; color: #4a4a4a; padding: 5px 25px; border-radius: 3px; background: #fff;">
<p>
	Hi there,
</p>


	<p>
		This is a line
	</p>

	<p>
		And another one
	</p>



	<a href="https://example.com" title="The action"
		style="position: relative;text-decoration:none;display: block;border-radius: 4px;cursor: pointer;padding-bottom: 8px;padding-left: 14px;padding-right: 14px;padding-top: 8px;width:280px;margin:10px auto;text-align: center;white-space: nowrap;border: 0;text-transform: uppercase;font-size: 14px;font-weight: 700;-webkit-box-shadow: 0 3px 6px rgba(107,114,128,.12),0 2px 4px rgba(107,114,128,.1);box-shadow: 0 3px 6px rgba(107,114,128,.12),0 2px 4px rgba(107,114,128,.1);background-color: #1973ff;border-color: transparent;color: #fff;">
		The action
	</a>



	<p>
		This should be an outro line
	</p>

	<p>
		And one more, because why not?
	</p>



	<p style="color: #9CA3AF;font-size:12px;border-top: 1px solid #dbdbdb;margin-top:20px;padding-top:20px;">
		If the button above doesn't work, copy the url below and paste it in your browsers address bar:<br/>
		https://example.com
	</p>

</div>
</div>
</div>
</body>
</html>
`, mailopts.HTMLMessage)
}
