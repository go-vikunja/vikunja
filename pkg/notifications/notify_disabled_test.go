package notifications_test

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"
)

type disabledNotifiable struct{}

func (d *disabledNotifiable) RouteForMail() (string, error) { return "test@example.com", nil }
func (d *disabledNotifiable) RouteForDB() int64             { return 1 }
func (d *disabledNotifiable) ShouldNotify() (bool, error)   { return true, nil }
func (d *disabledNotifiable) Lang() string                  { return "en" }

type disabledNotification struct{}

func (n *disabledNotification) ToMail(string) *notifications.Mail {
	return notifications.NewMail().Subject("Test").Line("Test")
}
func (n *disabledNotification) ToDB() interface{} { return nil }
func (n *disabledNotification) Name() string      { return "disabled.notification" }

func TestNotifyMailerDisabledReturns(t *testing.T) {
	config.InitDefaultConfig()
	config.MailerEnabled.Set(false)
	i18n.Init()

	done := make(chan struct{})
	go func() {
		_ = notifications.Notify(&disabledNotifiable{}, &disabledNotification{})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("notify blocked when mailer disabled")
	}
}
