// Provide data for a projects page.

package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	ProjectProposalsPageID = "5zq"
	RequestedProjectState  = "requested"
	InProgressProjectState = "inProgress"
	FinishedProjectState   = "finished"
)

type projectsParams struct {
}

var projectsHandler = siteHandler{
	URI:         "/json/projects/",
	HandlerFunc: projectsHandlerFunc,
	Options:     pages.PageOptions{},
}

type ProjectsData struct {
	// All projects
	Projects []*Project `json:"projects"`
}

type Project struct {
	ID    string `json:"id"`
	State string `json:"state"`
	// Upvote for "will write" and "want to read"
	WriteLikes *core.Likeable `json:"writeLikes"`
	ReadLikes  *core.Likeable `json:"readLikes"`
	// Project plan page id
	ProjectPageID string `json:"projectPageId"`
	// Start page id
	StartPageID string `json:"startPageId"`
}

type FinishedProject struct {
	// LOAD: start page summary
	// LOAD: Read by #
	// LOAD: Number of pages in the arc
}

type InProgressProject struct {
	// LOAD: Percent complete
	// LOAD: Edits and editors in the past week
}

type RequestedProject struct {
}

func projectsHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data projectsParams
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	projectsData := &ProjectsData{}

	// Load all projects
	projectsData.Projects, err = loadProjects(db, returnData)
	if err != nil {
		return pages.Fail("Error loading project info", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["projectsData"] = projectsData
	return pages.Success(returnData)
}

// Load info for all projects
func loadProjects(db *database.DB, returnData *core.CommonHandlerData) ([]*Project, error) {
	likeablesMap := make(map[int64]*core.Likeable)
	projects := make([]*Project, 0)
	rows := database.NewQuery(`
		SELECT id,projectPageId,startPageId,state,readLikeableId,writeLikeableId
		FROM projects`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var project Project
		project.ReadLikes = core.NewLikeable(core.ProjectReadLikeableType)
		project.WriteLikes = core.NewLikeable(core.ProjectWriteLikeableType)
		err := rows.Scan(&project.ID, &project.ProjectPageID, &project.StartPageID, &project.State,
			&project.ReadLikes.LikeableID, &project.WriteLikes.LikeableID)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		projects = append(projects, &project)
		if project.ReadLikes.LikeableID > 0 {
			likeablesMap[project.ReadLikes.LikeableID] = project.ReadLikes
		}
		if project.WriteLikes.LikeableID > 0 {
			likeablesMap[project.WriteLikes.LikeableID] = project.WriteLikes
		}
		core.AddPageToMap(project.ProjectPageID, returnData.PageMap, &core.PageLoadOptions{
			Summaries: true,
		})
		if project.StartPageID != "" && project.State == FinishedProjectState {
			core.AddPageToMap(project.StartPageID, returnData.PageMap, &core.PageLoadOptions{
				// TODO: load read by #
				Summaries: true,
				Path:      true,
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load projects: %v", err)
	}

	err = core.LoadLikes(db, returnData.User, likeablesMap, nil, returnData.UserMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load project likes: %v", err)
	}

	return projects, nil
}
