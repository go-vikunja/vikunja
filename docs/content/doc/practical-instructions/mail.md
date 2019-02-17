---
date: "2019-02-12:00:00+02:00"
title: "Mailer"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "practical instructions"
---

# Mailer

This document explains how to use the mailer to send emails and what to do to create a new kind of email to be sent.

## Sending emails

**Note:** You should use mail templates whenever possible (see below).

To send an email, use the function `mail.SendMail(options)`. The options are defined as follows:

{{< highlight golang >}}
type Opts struct {
	To          string // The email address of the recipent
	Subject     string // The subject of the mail
	Message     string // The plaintext message in the mail
	HTMLMessage string // The html message 
	ContentType ContentType // The content type of the mail. Can be either mail.ContentTypePlain, mail.ContentTypeHTML, mail.ContentTypeMultipart. You should set this according to the kind of mail you want to send.
	Boundary    string
	Headers     []*header // Other headers to set in the mail.
}
{{< /highlight >}}

## Sending emails based on a template

For each mail with a template, there are two email templates: One for plaintext emails, one for html emails.

These are located in the `templates/mail` folder and follow the conventions of `template-name.{plain|hmtl}.tmpl`,
both the plaintext and html templates are in the same folder.

To send a mail based on a template, use the function `mail.SendMailWithTemplate(to, subject, tpl string, data map[string]interface{})`.
`to` and `subject` are pretty much self-explanatory, `tpl` is the name of the template, without `.html.tmpl` or `.plain.tmpl`. 
`data` is a map you can pass additional data to your template.

#### Sending a mail with a template

A basic html email template would look like this:

{{< highlight go-html-template >}}
{{template "mail-header.tmpl" .}}
<p>
    Hey there!<br/>
    This is a minimal html email example.<br/>
    {{.Something}}
</p>
{{template "mail-footer.tmpl"}}
{{< /highlight >}}

And the corresponding plaintext template:

{{< highlight go-text-template >}}
Hey there!

This is a minimal html email example.

{{.Something}}
{{< /highlight >}}
You would then call this like so:

{{< highlight golang >}}
data := make(map[string]interface{})
data["Something"] = "I am some computed value"
to := "test@example.com"
subject := "A simple test mail"
tpl := "demo" // Assuming you saved the templates as demo.plain.tmpl and demo.html.tmpl
mail.SendMailWithTemplate(to, subject, tpl, data)
{{< /highlight >}}

The function does not return an error. If an error occures when sending a mail, it is logged but not returned because sending the mail happens asinchrounly.

Notice the `mail-header.tmpl` and `mail-footer.tmpl` in the template. These populate some basic css, a box for your content and the vikunja logo.
All that's left for you is to put the content in, which then will appear in a beautifully-styled box.

Remeber, these are email templates. This is different from normal html/css, you cannot use whatever you want (because most of the clients are wayyy to outdated).

