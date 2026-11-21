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

package mail

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"time"

	gomail "github.com/wneessen/go-mail"
)

// SendCompletionNotice has separate credentials so client notices do not enable account-wide reminders.
func SendCompletionNotice(opts *Opts) error {
	port, err := strconv.Atoi(os.Getenv("HIRAYAMA_COMPLETION_SMTP_PORT"))
	if err != nil || (port != 465 && port != 587) {
		return fmt.Errorf("invalid completion SMTP port")
	}
	host := os.Getenv("HIRAYAMA_COMPLETION_SMTP_HOST")
	username := os.Getenv("HIRAYAMA_COMPLETION_SMTP_USERNAME")
	password := os.Getenv("HIRAYAMA_COMPLETION_SMTP_PASSWORD")
	if host == "" || username == "" || password == "" {
		return fmt.Errorf("completion SMTP is not configured")
	}
	from, parseErr := mail.ParseAddress(opts.From)
	if parseErr != nil || from.Address != username {
		return fmt.Errorf("completion SMTP sender must match authenticated mailbox")
	}
	options := []gomail.Option{
		gomail.WithPort(port),
		gomail.WithUsername(username),
		gomail.WithPassword(password),
		gomail.WithSMTPAuth(gomail.SMTPAuthLogin),
		gomail.WithTLSPortPolicy(gomail.TLSMandatory),
		gomail.WithTLSConfig(&tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		}),
		gomail.WithTimeout(20 * time.Second),
	}
	if port == 465 {
		options = append(options, gomail.WithSSLPort(true))
	}
	client, err := gomail.NewClient(host, options...)
	if err != nil {
		return err
	}
	return client.DialAndSend(getMessage(opts))
}
