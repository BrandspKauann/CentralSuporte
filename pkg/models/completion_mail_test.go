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
	"errors"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	vikunjamail "code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/require"
)

func completionTestSettings(t *testing.T) completionMailSettings {
	t.Helper()
	t.Setenv("HIRAYAMA_COMPLETION_EMAIL_ENABLED", "true")
	t.Setenv("HIRAYAMA_COMPLETION_EMAIL_PROJECT_ID", "1")
	t.Setenv("HIRAYAMA_COMPLETION_EMAIL_ENABLED_AFTER", "2026-11-21T00:00:00Z")
	t.Setenv("HIRAYAMA_COMPLETION_EMAIL_TO", "client@example.com,other@example.com,client@example.com")
	settings, err := readCompletionMailSettings()
	require.NoError(t, err)
	return settings
}

func TestHirayamaCompletionBoundaries(t *testing.T) {
	settings := completionTestSettings(t)
	require.Len(t, settings.recipients, 2)
	task := &Task{
		ID:        1,
		ProjectID: 1,
		Done:      true,
		DoneAt:    settings.enabledAfter.Add(time.Hour),
	}
	require.True(t, completionMailEligible(task, settings))
	task.ProjectID = 2
	require.False(t, completionMailEligible(task, settings))
	task.ProjectID = 1
	task.Done = false
	require.False(t, completionMailEligible(task, settings))
	task.Done = true
	task.DoneAt = settings.enabledAfter.Add(-time.Second)
	require.False(t, completionMailEligible(task, settings), "historical completion must not send")
	t.Setenv("HIRAYAMA_COMPLETION_EMAIL_TO", "bad\r\nBcc: hidden@example.com")
	_, err := readCompletionMailSettings()
	require.Error(t, err)
	t.Setenv("HIRAYAMA_COMPLETION_EMAIL_ENABLED", "false")
	disabled, err := readCompletionMailSettings()
	require.NoError(t, err)
	require.Zero(t, disabled.projectID)
}

func TestHirayamaCompletionContent(t *testing.T) {
	delivery := &CompletionMail{
		TaskID:    42,
		Recipient: "client@example.com",
		Title:     "<script>Task</script>\r\nSubject: injected",
	}
	mail := completionMailContent(delivery)
	require.NotContains(t, mail.HTMLMessage, "<script>")
	require.NotContains(t, mail.Subject, "\r")
	require.NotContains(t, mail.Subject, "\n")
	require.Contains(t, mail.Message, "/tasks/42")
	require.Equal(t, delivery.Recipient, mail.To)
	require.False(t, strings.Contains(mail.HTMLMessage, "comentário interno"))
}

func TestHirayamaCompletionQueue(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	cleanup := db.NewSession()
	_, cleanupErr := cleanup.Where("id > ?", 0).Delete(&CompletionMail{})
	require.NoError(t, cleanupErr)
	require.NoError(t, cleanup.Commit())
	require.NoError(t, cleanup.Close())
	settings := completionTestSettings(t)
	s := db.NewSession()
	task, err := GetTaskByIDSimple(s, 1)
	require.NoError(t, err)
	task.Done = true
	task.DoneAt = time.Now()
	_, err = s.ID(task.ID).Cols("done", "done_at").Update(&task)
	require.NoError(t, err)
	require.NoError(t, enqueueCompletionMail(s, &task, settings))
	require.NoError(t, enqueueCompletionMail(s, &task, settings))
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())
	read := db.NewSession()
	var deliveries []*CompletionMail
	require.NoError(t, read.Where("task_id = ?", task.ID).Find(&deliveries))
	require.Len(t, deliveries, 2, "one persistent delivery per recipient")
	require.NoError(t, read.Close())
	oldSend := sendCompletionMail
	t.Cleanup(func() { sendCompletionMail = oldSend })
	sent := 0
	sendCompletionMail = func(opts *vikunjamail.Opts) error { sent++; return nil }
	require.NoError(t, deliverCompletionMail(deliveries[0], settings))
	require.NoError(t, deliverCompletionMail(deliveries[0], settings))
	require.Equal(t, 1, sent, "sent row must not send again")
	verify := db.NewSession()
	var persisted CompletionMail
	_, err = verify.ID(deliveries[0].ID).Get(&persisted)
	require.NoError(t, err)
	require.Equal(t, "sent", persisted.Status)
	require.False(t, persisted.SentAt.IsZero())
	require.NoError(t, verify.Close())

	sendCompletionMail = func(opts *vikunjamail.Opts) error { return errors.New("temporary SMTP failure") }
	require.NoError(t, deliverCompletionMail(deliveries[1], settings))
	verify = db.NewSession()
	persisted = CompletionMail{}
	_, err = verify.ID(deliveries[1].ID).Get(&persisted)
	require.NoError(t, err)
	require.Equal(t, "pending", persisted.Status)
	require.Equal(t, 1, persisted.Attempts)
	require.True(t, persisted.NextAttempt.After(time.Now()))
	require.NoError(t, verify.Close())
}

func TestHirayamaCompletionRollback(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	cleanup := db.NewSession()
	_, cleanupErr := cleanup.Where("id > ?", 0).Delete(&CompletionMail{})
	require.NoError(t, cleanupErr)
	require.NoError(t, cleanup.Commit())
	require.NoError(t, cleanup.Close())
	settings := completionTestSettings(t)
	s := db.NewSession()
	task := &Task{
		ID:        1,
		ProjectID: 1,
		Done:      true,
		DoneAt:    time.Now(),
		Title:     "Rollback",
	}
	require.NoError(t, enqueueCompletionMail(s, task, settings))
	require.NoError(t, s.Rollback())
	require.NoError(t, s.Close())
	verify := db.NewSession()
	defer verify.Close()
	count, err := verify.Where("task_id = ?", 1).Count(&CompletionMail{})
	require.NoError(t, err)
	require.Zero(t, count, "rolled back task changes must not leave a delivery")
}

func TestHirayamaCompletionFailedAndReopened(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	cleanup := db.NewSession()
	_, cleanupErr := cleanup.Where("id > ?", 0).Delete(&CompletionMail{})
	require.NoError(t, cleanupErr)
	require.NoError(t, cleanup.Commit())
	require.NoError(t, cleanup.Close())
	settings := completionTestSettings(t)
	s := db.NewSession()
	task, err := GetTaskByIDSimple(s, 1)
	require.NoError(t, err)
	task.Done = true
	task.DoneAt = time.Now()
	_, err = s.ID(1).Cols("done", "done_at").Update(&task)
	require.NoError(t, err)
	require.NoError(t, enqueueCompletionMail(s, &task, settings))
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())
	oldSend := sendCompletionMail
	t.Cleanup(func() { sendCompletionMail = oldSend })
	sendCompletionMail = func(opts *vikunjamail.Opts) error { return errors.New("SMTP unavailable") }
	read := db.NewSession()
	var deliveries []*CompletionMail
	require.NoError(t, read.Where("task_id = ?", 1).Find(&deliveries))
	require.NoError(t, read.Close())
	for range 5 {
		require.NoError(t, deliverCompletionMail(deliveries[0], settings))
	}
	verify := db.NewSession()
	var delivery CompletionMail
	_, err = verify.ID(deliveries[0].ID).Get(&delivery)
	require.NoError(t, err)
	require.Equal(t, "failed", delivery.Status)
	require.Equal(t, 5, delivery.Attempts)
	task.Done = false
	_, err = verify.ID(1).Cols("done").Update(&task)
	require.NoError(t, err)
	require.NoError(t, verify.Commit())
	require.NoError(t, verify.Close())
	sendCompletionMail = func(opts *vikunjamail.Opts) error { t.Fatal("reopened task sent mail"); return nil }
	require.NoError(t, deliverCompletionMail(deliveries[1], settings))
}

func TestHirayamaCompletionConfiguration(t *testing.T) {
	completionTestSettings(t)
	for _, key := range []string{"HIRAYAMA_COMPLETION_EMAIL_PROJECT_ID", "HIRAYAMA_COMPLETION_EMAIL_ENABLED_AFTER", "HIRAYAMA_COMPLETION_EMAIL_TO"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "")
			_, err := readCompletionMailSettings()
			require.Error(t, err)
		})
	}
	t.Setenv("HIRAYAMA_COMPLETION_EMAIL_ENABLED", "false")
	processCompletionMail()
}

func TestHirayamaCompletionTaskUpdate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	cleanup := db.NewSession()
	_, err := cleanup.Where("id > ?", 0).Delete(&CompletionMail{})
	require.NoError(t, err)
	require.NoError(t, cleanup.Commit())
	require.NoError(t, cleanup.Close())
	completionTestSettings(t)
	s := db.NewSession()
	task, err := GetTaskByIDSimple(s, 1)
	require.NoError(t, err)
	task.Done = true
	require.NoError(t, task.Update(s, &user.User{ID: 1}))
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())
	verify := db.NewSession()
	defer verify.Close()
	count, err := verify.Where("task_id = ?", 1).Count(&CompletionMail{})
	require.NoError(t, err)
	require.EqualValues(t, 2, count)
}
