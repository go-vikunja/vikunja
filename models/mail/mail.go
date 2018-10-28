package mail

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
	"time"
)

// Queue is the mail queue
var Queue chan *gomail.Message

// StartMailDaemon starts the mail daemon
func StartMailDaemon() {
	Queue = make(chan *gomail.Message, viper.GetInt("mailer.queuelength"))

	if viper.GetString("mailer.host") == "" {
		//models.Log.Warning("Mailer seems to be not configured! Please see the config docs for more details.")
		fmt.Println("Mailer seems to be not configured! Please see the config docs for more details.")
		return
	}

	go func() {
		d := gomail.NewDialer(viper.GetString("mailer.host"), viper.GetInt("mailer.port"), viper.GetString("mailer.username"), viper.GetString("mailer.password"))
		d.TLSConfig = &tls.Config{InsecureSkipVerify: viper.GetBool("mailer.skiptlsverify")}

		var s gomail.SendCloser
		var err error
		open := false
		for {
			select {
			case m, ok := <-Queue:
				if !ok {
					return
				}
				if !open {
					if s, err = d.Dial(); err != nil {
						// models.Log.Error("Error during connect to smtp server: %s", err)
						fmt.Printf("Error during connect to smtp server: %s \n", err)
					}
					open = true
				}
				if err := gomail.Send(s, m); err != nil {
					// models.Log.Error("Error when sending mail: %s", err)
					fmt.Printf("Error when sending mail: %s \n", err)
				}
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
			case <-time.After(viper.GetDuration("mailer.queuetimeout") * time.Second):
				if open {
					if err := s.Close(); err != nil {
						fmt.Printf("Error closing the mail server connection: %s\n", err)
					}
					fmt.Println("Closed connection to mailserver")
					open = false
				}
			}
		}
	}()
}

// StopMailDaemon closes the mail queue channel, aka stops the daemon
func StopMailDaemon() {
	close(Queue)
}
