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

package migration

import (
	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCompletionMailMigration(t *testing.T) {
	engine, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, engine.Sync(completionMail20261121153000{}))
	first := &completionMail20261121153000{
		TaskID:      999999,
		Recipient:   "client@example.com",
		Title:       "Completion",
		Status:      "pending",
		NextAttempt: time.Now(),
		LeaseUntil:  time.Unix(0, 0),
	}
	_, err = engine.Insert(first)
	require.NoError(t, err)
	duplicate := *first
	duplicate.ID = 0
	_, err = engine.Insert(&duplicate)
	require.Error(t, err, "unique task/recipient constraint must reject duplicate delivery")
	duplicate.Recipient = "other@example.com"
	_, err = engine.Insert(&duplicate)
	require.NoError(t, err)
	require.NoError(t, engine.Sync(completionMail20261121153000{}))
	count, err := engine.Where("task_id = ?", first.TaskID).Count(&completionMail20261121153000{})
	require.NoError(t, err)
	require.EqualValues(t, 2, count, "restarting migration must preserve queued deliveries")
}
