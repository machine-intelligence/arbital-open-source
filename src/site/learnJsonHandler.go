// learnJsonHandler.go returns the learn of pages needed for understanding a page
package site

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var learnHandler = siteHandler{
	URI:         "/json/learn/",
	HandlerFunc: learnJsonHandler,
}

type learnJsonData struct {
	PageId string
}

// A node in the output tree for the learning path
type learnNode struct {
	// User wants to understand PageId
	PageId string `json:"pageId"`
	// To understand it, the user will read TaughtById
	TaughtById string `json:"taughtById"`
	// To understand TaughtById, the user needs to have the following Requirements
	RequirementIds []string `json:"requirementIds"`
}

// Requirement the user needs to acquire in order to read a tutor page
type requirementNode struct {
	PageId string
	// Which pages can teach this requirement
	TutorIds []string
	// Best tutor
	BestTutorId string
	// Cost assigned to learning this node
	Cost int
	// Set to true when the node has been processed
	Processed bool
}

// Page that will teach the user about stuff.
type tutorNode struct {
	PageId string
	// To read this page, the user needs these requirements
	RequirementIds []string
	// Cost assigned to learning this node
	Cost int
	// Set to true when the node has been processed
	Processed bool

	// Need to set this map for sorting to work
	RequirementMap map[string]*requirementNode
}

// Sort node's requirements
func (t *tutorNode) Len() int { return len(t.RequirementIds) }
func (t *tutorNode) Swap(i, j int) {
	t.RequirementIds[i], t.RequirementIds[j] = t.RequirementIds[j], t.RequirementIds[i]
}
func (t *tutorNode) Less(i, j int) bool {
	return t.RequirementMap[t.RequirementIds[i]].Cost < t.RequirementMap[t.RequirementIds[j]].Cost
}

func newLearnNode(pageId string) *learnNode {
	return &learnNode{PageId: pageId, TaughtById: "", RequirementIds: make([]string, 0)}
}

func newRequirementNode(pageId string) *requirementNode {
	return &requirementNode{PageId: pageId, TutorIds: make([]string, 0)}
}

func newTutorNode(pageId string) *tutorNode {
	return &tutorNode{PageId: pageId, RequirementIds: make([]string, 0)}
}

func learnJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	c := params.C

	// Decode data
	var data learnJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Need a valid pageId", nil)
	}

	masteryMap := make(map[string]*core.Mastery)
	// Load masteryMap from the cookie, if present
	cookie, err := params.R.Cookie("masteryMap")
	if err == nil {
		jsonStr, _ := url.QueryUnescape(cookie.Value)
		err = json.Unmarshal([]byte(jsonStr), &masteryMap)
		if err != nil {
			params.C.Warningf("Couldn't unmarshal masteryMap cookie: %v", err)
		}
	}

	returnData := newHandlerData(true)
	returnData.User = u

	// Check if the user already has this requirement
	hasMastery := false
	if mastery, ok := masteryMap[data.PageId]; ok {
		hasMastery = mastery.Has
	}
	if !hasMastery && u.Id != "" {
		row := database.NewQuery(`
		SELECT ifnull(max(has),false)
		FROM userMasteryPairs
		WHERE userId=?`, u.Id).Add(`AND masteryId=?`, data.PageId).ToStatement(db).QueryRow()
		_, err = row.Scan(&hasMastery)
		if err != nil {
			return pages.HandlerErrorFail("Error while checking if already knows", err)
		}
	}

	// Track which requirements we need to process in the next step
	requirementIds := make([]string, 0)
	if !hasMastery {
		requirementIds = append(requirementIds, data.PageId)
	}
	// Which tutor pages to load in the next step
	tutorIds := make([]string, 0)

	// Create the maps which will store all the nodes: page id -> node
	tutorMap := make(map[string]*tutorNode)
	requirementMap := make(map[string]*requirementNode)
	requirementMap[data.PageId] = newRequirementNode(data.PageId)

	// What to load for the pages
	loadOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)

	// Recursively find which pages the user has to read
	for maxCount := 0; len(requirementIds) > 0 && maxCount < 20; maxCount++ {
		c.Debugf("TequirementIds: %+v", requirementIds)
		// Load which pages teach the requirements
		tutorIds = make([]string, 0)
		rows := database.NewQuery(`
			SELECT pp.parentId,pp.childId
			FROM pagePairs AS pp
			WHERE pp.parentId IN`).AddArgsGroupStr(requirementIds).Add(`
				AND pp.type=?`, core.SubjectPagePairType).Add(`
			`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentId, childId string
			err := rows.Scan(&parentId, &childId)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			c.Debugf("Found tutor: %s %s", parentId, childId)
			// Get the requirement node and update its tutors
			requirementNode := requirementMap[parentId]
			requirementNode.TutorIds = append(requirementNode.TutorIds, childId)
			c.Debugf("Updated requirement node: %+v", requirementNode)
			// Recursively load requirements for the tutor, unless we already processed it
			if _, ok := tutorMap[childId]; !ok {
				tutorIds = append(tutorIds, childId)
				tutorMap[childId] = newTutorNode(childId)
			}
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("Error while loading tutors", err)
		}
		c.Debugf("TutorIds: %+v", tutorIds)
		if len(tutorIds) <= 0 {
			break
		}

		// Load the requirements for the tutors
		requirementIds = make([]string, 0)
		rows = database.NewQuery(`
			SELECT pp.parentId,pp.childId
			FROM pagePairs AS pp
			LEFT JOIN userMasteryPairs AS mp
			ON (pp.parentId=mp.masteryId AND mp.userId=?)`, u.Id).Add(`
			WHERE pp.childId IN`).AddArgsGroupStr(tutorIds).Add(`
				AND pp.type=?`, core.RequirementPagePairType).Add(`
				AND (NOT mp.has OR isnull(mp.has))`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentId, childId string
			err := rows.Scan(&parentId, &childId)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			c.Debugf("Found requirement: %s %s", parentId, childId)

			// Get the tutor node and update its requirements
			tutorNode := tutorMap[childId]
			tutorNode.RequirementIds = append(tutorNode.RequirementIds, parentId)
			c.Debugf("Updated tutor node: %+v", tutorNode)
			if _, ok := requirementMap[parentId]; !ok {
				requirementIds = append(requirementIds, parentId)
				requirementMap[parentId] = newRequirementNode(parentId)
			}
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("Error while loading requirements", err)
		}
		if maxCount >= 18 {
			c.Warningf("Max count is close to maximum: %d", maxCount)
		}
	}

	var processRequirement func(reqId string) *requirementNode

	// Process the tutor node
	processTutor := func(tutorId string) *tutorNode {
		t := tutorMap[tutorId]
		if t.Processed {
			return t
		}
		t.Cost = 10000 // set the cost high in case there is a loop
		t.Processed = true
		costSum := 0
		c.Debugf("Processing tutor: %s", tutorId)
		// Do processing
		for _, reqId := range t.RequirementIds {
			costSum += processRequirement(reqId).Cost
		}
		t.Cost = costSum + 1
		t.RequirementMap = requirementMap
		sort.Sort(t)
		c.Debugf("Tutor %s has cost: %d", tutorId, t.Cost)
		return t
	}

	// Process the requirement node
	processRequirement = func(reqId string) *requirementNode {
		r := requirementMap[reqId]
		if r.Processed {
			return r
		}
		r.Cost = 10000 // cost of a requirement without a tutor
		r.Processed = true
		noTutor := true // make sure we take at least one tutor, no matter how bad
		c.Debugf("Processing requirement: %s", reqId)
		for _, tutorId := range r.TutorIds {
			cost := processTutor(tutorId).Cost
			if r.Cost > cost || noTutor {
				r.Cost = cost
				r.BestTutorId = tutorId
				noTutor = false
			}
		}
		c.Debugf("Best tutor for requirement %s is %s (%d)", reqId, r.BestTutorId, r.Cost)
		return r
	}

	// Process all the nodes
	processRequirement(data.PageId)

	// Create the learning map, and add the used pages to pageMap
	learnMap := make(map[string]*learnNode)
	requirementIds = make([]string, 0)
	requirementIds = append(requirementIds, data.PageId)
	for len(requirementIds) > 0 {
		copyReqIds := append(make([]string, 0), requirementIds...)
		requirementIds = make([]string, 0)
		// Process each requirement node
		for _, reqId := range copyReqIds {
			learnNode, ok := learnMap[reqId]
			if ok {
				continue
			}
			learnNode = newLearnNode(reqId)
			learnMap[reqId] = learnNode
			core.AddPageToMap(learnNode.PageId, returnData.PageMap, loadOptions)
			// Get corresponding requirement node
			reqNode := requirementMap[reqId]
			// Pick the best tutor
			learnNode.TaughtById = reqNode.BestTutorId
			if learnNode.TaughtById != "" {
				core.AddPageToMap(learnNode.TaughtById, returnData.PageMap, loadOptions)
				// Now populate the requirementIds
				tutorNode, ok := tutorMap[learnNode.TaughtById]
				if ok {
					learnNode.RequirementIds = tutorNode.RequirementIds
					requirementIds = append(requirementIds, tutorNode.RequirementIds...)
				}
			}
		}
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["learnMap"] = learnMap
	return pages.StatusOK(returnData.toJson())
}
