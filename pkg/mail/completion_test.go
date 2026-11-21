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
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHirayamaCompletionTransportValidation(t *testing.T) {
	t.Setenv("HIRAYAMA_COMPLETION_SMTP_PORT", "25")
	require.ErrorContains(t, SendCompletionNotice(&Opts{}), "port")
	t.Setenv("HIRAYAMA_COMPLETION_SMTP_PORT", "465")
	t.Setenv("HIRAYAMA_COMPLETION_SMTP_HOST", "")
	require.ErrorContains(t, SendCompletionNotice(&Opts{}), "not configured")
	t.Setenv("HIRAYAMA_COMPLETION_SMTP_HOST", "smtp.example.com")
	t.Setenv("HIRAYAMA_COMPLETION_SMTP_USERNAME", "sender@example.com")
	t.Setenv("HIRAYAMA_COMPLETION_SMTP_PASSWORD", "test-only-not-a-real-credential")
	require.ErrorContains(t, SendCompletionNotice(&Opts{From: "different@example.com"}), "match authenticated")
}
