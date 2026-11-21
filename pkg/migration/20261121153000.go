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
	"src.techknowlogick.com/xormigrate"
	"time"
	"xorm.io/xorm"
)

type completionMail20261121153000 struct {
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

func (completionMail20261121153000) TableName() string { return "completion_mail" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261121153000",
		Description: "Add durable task completion email delivery queue",
		Migrate:     func(tx *xorm.Engine) error { return tx.Sync(completionMail20261121153000{}) },
		Rollback:    func(tx *xorm.Engine) error { return tx.DropTables(completionMail20261121153000{}) },
	})
}
