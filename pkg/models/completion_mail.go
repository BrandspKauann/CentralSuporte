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

package models

import (
	"encoding/json"
	"fmt"
	"html"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	vikunjamail "code.vikunja.io/api/pkg/mail"
	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/xorm"
)

// CompletionMail records one delivery per task and recipient across restarts.
type CompletionMail struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID      int64     `xorm:"bigint not null unique(task_recipient)"`
	Recipient   string    `xorm:"varchar(254) not null unique(task_recipient)"`
	Title       string    `xorm:"text not null"`
	Status      string    `xorm:"varchar(20) not null index"`
	Attempts    int       `xorm:"not null"`
	NextAttempt time.Time `xorm:"not null index"`
	LeaseUntil  time.Time `xorm:"not null"`
	SentAt      time.Time `xorm:"null"`
	Created     time.Time `xorm:"created not null"`
}

func (CompletionMail) TableName() string { return "completion_mail" }

type completionMailSettings struct {
	projectID    int64
	recipients   []string
	enabledAfter time.Time
}

func readCompletionMailSettings() (settings completionMailSettings, err error) {
	if os.Getenv("HIRAYAMA_COMPLETION_EMAIL_ENABLED") != "true" {
		return settings, nil
	}
	settings.projectID, err = strconv.ParseInt(os.Getenv("HIRAYAMA_COMPLETION_EMAIL_PROJECT_ID"), 10, 64)
	if err != nil || settings.projectID < 1 {
		return settings, fmt.Errorf("invalid completion email project")
	}
	settings.enabledAfter, err = time.Parse(time.RFC3339, os.Getenv("HIRAYAMA_COMPLETION_EMAIL_ENABLED_AFTER"))
	if err != nil {
		return settings, fmt.Errorf("invalid completion email activation date")
	}
	seen := map[string]bool{}
	for _, address := range strings.Split(os.Getenv("HIRAYAMA_COMPLETION_EMAIL_TO"), ",") {
		address = strings.TrimSpace(address)
		parsed, parseErr := mail.ParseAddress(address)
		if parseErr != nil || parsed.Address != address {
			return settings, fmt.Errorf("invalid completion email recipient")
		}
		if !seen[address] {
			settings.recipients = append(settings.recipients, address)
			seen[address] = true
		}
	}
	if len(settings.recipients) == 0 {
		return settings, fmt.Errorf("completion email requires recipients")
	}
	return settings, nil
}

func completionMailEligible(task *Task, settings completionMailSettings) bool {
	return task != nil && settings.projectID > 0 && task.ProjectID == settings.projectID && task.Done &&
		!task.DoneAt.IsZero() && !task.DoneAt.Before(settings.enabledAfter)
}

type SendCompletionMail struct{}

func (*SendCompletionMail) Name() string { return "hirayama.task.completed.email" }
func (*SendCompletionMail) Handle(msg *message.Message) error {
	settings, err := readCompletionMailSettings()
	if err != nil {
		return err
	}
	if settings.projectID == 0 {
		return nil
	}
	event := &TaskUpdatedEvent{}
	if err = json.Unmarshal(msg.Payload, event); err != nil {
		return err
	}
	if !completionMailEligible(event.Task, settings) {
		return nil
	}
	s := db.NewSession()
	defer s.Close()
	if err = enqueueCompletionMail(s, event.Task, settings); err != nil {
		return err
	}
	if err = s.Commit(); err != nil {
		return err
	}
	processCompletionMail()
	return nil
}

func enqueueCompletionMail(s *xorm.Session, task *Task, settings completionMailSettings) error {
	if !completionMailEligible(task, settings) {
		return nil
	}
	for _, recipient := range settings.recipients {
		exists, err := s.Where("task_id = ? AND recipient = ?", task.ID, recipient).Exist(&CompletionMail{})
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		_, err = s.Insert(&CompletionMail{
			TaskID:      task.ID,
			Recipient:   recipient,
			Title:       task.Title,
			Status:      "pending",
			NextAttempt: time.Now(),
			LeaseUntil:  time.Unix(0, 0),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func completionMailContent(delivery *CompletionMail) *vikunjamail.Opts {
	url := strings.TrimRight(config.ServicePublicURL.GetString(), "/") + "/tasks/" + strconv.FormatInt(delivery.TaskID, 10)
	title := strings.Join(strings.Fields(delivery.Title), " ")
	plain := fmt.Sprintf("Olá, Ewerton.\n\nConcluí a tarefa: %s.\n\nOs detalhes e o registro do trabalho estão disponíveis no suporte:\n%s\n\nKauann Pires\nBazhi", title, url)
	body := fmt.Sprintf(`<div style="font-family:Arial,sans-serif;color:#193c38;max-width:600px;margin:auto;padding:28px"><p style="font-size:12px;letter-spacing:1px">BAZHI · SUPORTE HIRAYAMA</p><h1 style="font-size:24px">Atualização concluída</h1><p>Olá, Ewerton.</p><p>Concluí a tarefa <strong>%s</strong>.</p><p>Você pode acompanhar os detalhes e o registro do trabalho no suporte.</p><p style="margin:28px 0"><a href="%s" style="background:#193c38;color:white;padding:12px 18px;border-radius:6px;text-decoration:none">Ver tarefa concluída</a></p><p>Kauann Pires<br>Bazhi</p></div>`, html.EscapeString(title), html.EscapeString(url))
	return &vikunjamail.Opts{
		From:        "Kauann Pires <" + os.Getenv("HIRAYAMA_COMPLETION_SMTP_FROM") + ">",
		To:          delivery.Recipient,
		Subject:     "Suporte Hirayama | Concluído: " + title,
		Message:     plain,
		HTMLMessage: body,
		ContentType: vikunjamail.ContentTypeMultipart,
	}
}

var sendCompletionMail = vikunjamail.SendCompletionNotice

func registerCompletionMail() {
	settings, err := readCompletionMailSettings()
	if err != nil {
		log.Error("Completion email configuration invalid; notifications disabled")
		return
	}
	if settings.projectID == 0 {
		return
	}
	go processCompletionMail()
	if err = cron.Schedule("* * * * *", processCompletionMail); err != nil {
		log.Error("Could not schedule completion email retries")
	}
}

func processCompletionMail() {
	settings, err := readCompletionMailSettings()
	if err != nil || settings.projectID == 0 {
		return
	}
	s := db.NewSession()
	defer s.Close()
	var deliveries []*CompletionMail
	now := time.Now()
	err = s.Where("status = ? AND next_attempt <= ? AND lease_until <= ?", "pending", now, now).In("recipient", settings.recipients).Limit(20).Asc("id").Find(&deliveries)
	if err != nil {
		log.Error("Could not read completion email queue")
		return
	}
	for _, delivery := range deliveries {
		if err = deliverCompletionMail(delivery, settings); err != nil {
			log.Errorf("Completion email processing failed for task %d", delivery.TaskID)
		}
	}
}

func deliverCompletionMail(delivery *CompletionMail, settings completionMailSettings) error {
	s := db.NewSession()
	defer s.Close()
	now := time.Now()
	delivery.LeaseUntil = now.Add(2 * time.Minute)
	affected, err := s.ID(delivery.ID).Where("status = ? AND lease_until <= ?", "pending", now).
		Cols("lease_until").Update(delivery)
	if err != nil || affected == 0 {
		return err
	}
	if err = s.Commit(); err != nil {
		return err
	}

	check := db.NewSession()
	task, err := GetTaskByIDSimple(check, delivery.TaskID)
	_ = check.Close()
	if err != nil {
		return err
	}
	if !completionMailEligible(&task, settings) {
		delivery.Status = "cancelled"
	} else {
		delivery.Attempts++
		if err = sendCompletionMail(completionMailContent(delivery)); err == nil {
			delivery.Status = "sent"
			delivery.SentAt = time.Now()
		} else {
			// SMTP responses can contain personal data; persist only state and attempt count.
			log.Warningf("SMTP did not confirm completion email for task %d (attempt %d)", delivery.TaskID, delivery.Attempts)
			delivery.Status = "pending"
			if delivery.Attempts >= 5 {
				delivery.Status = "failed"
			}
			delivery.NextAttempt = time.Now().Add(time.Duration(delivery.Attempts*delivery.Attempts) * time.Minute)
		}
	}
	delivery.LeaseUntil = time.Unix(0, 0)
	save := db.NewSession()
	defer save.Close()
	_, err = save.ID(delivery.ID).Cols("status", "attempts", "next_attempt", "lease_until", "sent_at").Update(delivery)
	if err != nil {
		return err
	}
	if err = save.Commit(); err != nil {
		return err
	}
	log.Infof("Completion email task %d: %s", delivery.TaskID, delivery.Status)
	return nil
}

func enqueueConfiguredCompletionMail(s *xorm.Session, task *Task) error {
	settings, err := readCompletionMailSettings()
	if err != nil {
		return nil
	}
	return enqueueCompletionMail(s, task, settings)
}
