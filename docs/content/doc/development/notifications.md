---
date: 2021-02-07T19:26:34+02:00
title: "Notifications"
toc: true
draft: false
menu:
  sidebar:
    parent: "development"
---

# Notifications

Vikunjs provides a simple abstraction to send notifications per mail and in the database.

{{< table_of_contents >}}

## Definition

Each notification has to implement this interface:

{{< highlight golang >}}
type Notification interface {
    ToMail() *Mail
    ToDB() interface{}
    Name() string
}
{{< /highlight >}}

Both functions return the formatted messages for mail and database.

A notification will only be sent or recorded for those of the two methods which don't return `nil`.
For example, if your notification should not be recorded in the database but only sent out per mail, it is enough to let the `ToDB` function return `nil`.

### Mail notifications

A list of chainable functions is available to compose a mail:

{{< highlight golang >}}
mail := NewMail(). 
    // The optional sender of the mail message.
    From("test@example.com").
	// The optional receipient of the mail message. Uses the mail address of the notifiable if omitted.
    To("test@otherdomain.com").
	// The subject of the mail to send.
    Subject("Testmail").
	// The greeting, or "intro" line of the mail.
    Greeting("Hi there,").
	// A line of text
    Line("This is a line of text").
	// An action can contain a title and a url. It gets rendered as a big button in the mail.
	// Note that you can have only one action per mail.
	// All lines added before an action will appearr in the mail before the button, all lines 
	// added afterwards will appear after it.
    Action("The Action", "https://example.com").
	// Another line of text.
    Line("This should be an outro line").
{{< /highlight >}}

If not provided, the `from` field of the mail contains the value configured in [`mailer.fromemail`](https://vikunja.io/docs/config-options/#fromemail).

### Database notifications

All data returned from the `ToDB()` method is serialized to json and saved into the database, along with the id of the 
notifiable, the name of the notification and a time stamp.
If you don't use the database notification, the `Name()` function can return an empty string.

## Creating a new notification

The easiest way to generate a mail is by using the `mage dev:make-notification` command.

It takes the name of the notification and the package where the notification will be created.

## Notifiables

Notifiables can receive a notification.
A notifiable is defined with this interface:

{{< highlight golang >}}
type Notifiable interface {
    // Should return the email address this notifiable has.
    RouteForMail() string
    // Should return the id of the notifiable entity
    RouteForDB() int64
}
{{< /highlight >}}

The `User` type from the `user` package implements this interface.

## Sending a notification

Sending a notification is done with the `Notify` method from the `notifications` package.
It takes a notifiable and a notification as input.

For example, the email confirm notification when a new user registers is sent like this:

{{< highlight golang >}}
n := &EmailConfirmNotification{
    User:  update.User,
    IsNew: false,
}

err = notifications.Notify(update.User, n)
return
{{< /highlight >}}

## Testing

The `mail` package provides a `Fake()` method which you should call in the `MainTest` functions of your package.
If it was called, no mails are being sent and you can instead assert they have been sent with the `AssertSent` method.

## Example

Take a look at the [pkg/user/notifications.go](https://code.vikunja.io/api/src/branch/main/pkg/user/notifications.go) file for a good example.
