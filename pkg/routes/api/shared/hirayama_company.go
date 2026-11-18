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

package shared

import (
	"strings"
	"unicode/utf8"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/hirayama"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

const CompanyDescription = "Empresa da Central de Suporte"

type CompanyWorkspace struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	ProjectIDs []int64 `json:"project_ids"`
}

// Companies reuse private teams and project ACLs. A company name never grants membership.
func CreateCompanyWorkspace(s *xorm.Session, owner *user.User, name string) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 250 {
		return models.ErrInvalidData{Message: "Informe o nome da empresa (até 250 caracteres)."}
	}
	exists, err := s.Where("LOWER(name) = ?", strings.ToLower(name)).Exist(&models.Team{})
	if err != nil {
		return err
	}
	if exists {
		return models.ErrInvalidData{Message: "Esta empresa já está cadastrada. Entre com uma conta autorizada da empresa."}
	}
	team := &models.Team{Name: name, Description: CompanyDescription, IsPublic: false}
	if err := team.Create(s, owner); err != nil {
		return err
	}
	project := &models.Project{Title: name, Description: "Projetos e chamados da empresa."}
	if err := project.Create(s, owner); err != nil {
		return err
	}
	teamProject := &models.TeamProject{TeamID: team.ID, ProjectID: project.ID, Permission: models.PermissionWrite}
	if err := teamProject.Create(s, owner); err != nil {
		return err
	}
	supportID, err := hirayama.SupportUserID()
	if err != nil {
		return err
	}
	if supportID != owner.ID {
		support, err := user.GetUserByID(s, supportID)
		if err != nil {
			return err
		}
		member := &models.TeamMember{TeamID: team.ID, Username: support.Username, Admin: true}
		if err := member.Create(s, owner); err != nil {
			return err
		}
		admin := &models.ProjectUser{ProjectID: project.ID, Username: support.Username, Permission: models.PermissionAdmin}
		if err := admin.Create(s, owner); err != nil {
			return err
		}
	}
	owner.DefaultProjectID = project.ID
	_, err = s.ID(owner.ID).Cols("default_project_id").Update(owner)
	return err
}

func FindCompanyWorkspace(s *xorm.Session, account *user.User, name string) (*CompanyWorkspace, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, models.ErrInvalidData{Message: "Informe a empresa que deseja acessar."}
	}
	teams := []*models.Team{}
	err := s.Table("teams").Join("INNER", "team_members", "team_members.team_id = teams.id").
		Where("team_members.user_id = ? AND LOWER(teams.name) = ?", account.ID, strings.ToLower(name)).
		Select("teams.*").Find(&teams)
	if err != nil {
		return nil, err
	}
	if len(teams) != 1 {
		return nil, models.ErrGenericForbidden{}
	}
	links := []*models.TeamProject{}
	if err := s.Where("team_id = ?", teams[0].ID).OrderBy("project_id").Find(&links); err != nil {
		return nil, err
	}
	workspace := &CompanyWorkspace{ID: teams[0].ID, Name: teams[0].Name, ProjectIDs: []int64{}}
	for _, link := range links {
		workspace.ProjectIDs = append(workspace.ProjectIDs, link.ProjectID)
	}
	return workspace, nil
}
