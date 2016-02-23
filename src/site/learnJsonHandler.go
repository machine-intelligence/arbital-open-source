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

const (
	PenaltyCost = 10000
	LensCost    = 10
)

// Requirement the user needs to acquire in order to read a tutor page
type requirementNode struct {
	PageId    string `json:"pageId"`
	LensIndex int    `json:"-"`
	// Which pages can teach this requirement
	TutorIds []string `json:"tutorIds"`
	// Best tutor
	BestTutorId string `json:"bestTutorId"`
	// Cost assigned to learning this node
	Cost int `json:"cost"`
	// Set to true when the node has been processed
	Processed bool `json:"-"`
}

// Page that will teach the user about stuff.
type tutorNode struct {
	PageId    string `json:"pageId"`
	LensIndex int    `json:"-"`
	// To read this page, the user needs these requirements
	RequirementIds []string `json:"requirementIds"`
	// Cost assigned to learning this node
	Cost int `json:"cost"`
	// Set to true when the node has been processed
	Processed bool `json:"-"`

	// Need to set this map for sorting to work
	RequirementMap map[string]*requirementNode `json:"-"`
}

// Sort node's requirements
func (t *tutorNode) Len() int { return len(t.RequirementIds) }
func (t *tutorNode) Swap(i, j int) {
	t.RequirementIds[i], t.RequirementIds[j] = t.RequirementIds[j], t.RequirementIds[i]
}
func (t *tutorNode) Less(i, j int) bool {
	return t.RequirementMap[t.RequirementIds[i]].Cost < t.RequirementMap[t.RequirementIds[j]].Cost
}

func newRequirementNode(pageId string) *requirementNode {
	return &requirementNode{PageId: pageId, TutorIds: make([]string, 0), Cost: 10000000}
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
		c.Debugf("RequirementIds: %+v", requirementIds)
		// Load which pages teach the requirements
		tutorIds = make([]string, 0)
		rows := database.NewQuery(`
			SELECT pp.parentId,pp.childId,pi.lensIndex
			FROM pagePairs AS pp
			JOIN pageInfos AS pi
			ON (pp.childId=pi.pageId)
			WHERE pp.parentId IN`).AddArgsGroupStr(requirementIds).Add(`
				AND pp.type=?`, core.SubjectPagePairType).Add(`
			`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentId, childId string
			var lensIndex int
			err := rows.Scan(&parentId, &childId, &lensIndex)
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
				tutorMap[childId].LensIndex = lensIndex
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
			SELECT pp.parentId,pp.childId,pi.lensIndex
			FROM pagePairs AS pp
			JOIN pageInfos AS pi
			ON (pp.parentId=pi.pageId)
			LEFT JOIN userMasteryPairs AS mp
			ON (pp.parentId=mp.masteryId AND mp.userId=?)`, u.Id).Add(`
			WHERE pp.childId IN`).AddArgsGroupStr(tutorIds).Add(`
				AND pp.type=?`, core.RequirementPagePairType).Add(`
				AND (NOT mp.has OR isnull(mp.has))`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentId, childId string
			var lensIndex int
			err := rows.Scan(&parentId, &childId, &lensIndex)
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
				requirementMap[parentId].LensIndex = lensIndex
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

	/*var processRequirement func(reqId string) *requirementNode

	// Process the tutor node
	processTutor := func(tutorId string) *tutorNode {
		t := tutorMap[tutorId]
		if t.Processed {
			return t
		}
		core.AddPageToMap(tutorId, returnData.PageMap, loadOptions)
		t.Cost = PenaltyCost // set the cost high in case there is a loop
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
		core.AddPageToMap(reqId, returnData.PageMap, loadOptions)
		r.Cost = PenaltyCost // cost of a requirement without a tutor
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
	*/

	c.Debugf("================ PROCESSING ==================")

	// Mark all requirements with no teachers as processed
	for _, req := range requirementMap {
		if len(req.TutorIds) > 0 {
			continue
		}
		req.Cost = PenaltyCost
		req.Processed = true
		core.AddPageToMap(req.PageId, returnData.PageMap, loadOptions)
		c.Debugf("Requirement '%s' pre-processed with cost %d", req.PageId, req.Cost)
	}

	graphChanged := true
	// Keep processing all the nodes until we processed the node we want to learn
	for !requirementMap[data.PageId].Processed {
		if !graphChanged {
			// We didn't make any progress in the last iteration, which means there is
			// a cycle. Arbitrarily mark a requirement as processed.
			for _, req := range requirementMap {
				if req.Processed || req.PageId == data.PageId {
					continue
				}

				// Print the cycle
				cycleReq := req
				cycleIds := make([]string, 0)
				cycleIds = append(cycleIds, cycleReq.PageId)
				cycleReqMap := make(map[string]bool) // store all requirements we've met
				cycleReqMap[cycleReq.PageId] = true
				continueCycle := true
				for continueCycle {
					// Get first eligible tutor
					var cycleTutor *tutorNode
					for _, tutorId := range req.TutorIds {
						cycleTutor = tutorMap[tutorId]
						if !cycleTutor.Processed {
							cycleIds = append(cycleIds, cycleTutor.PageId)
							break
						}
					}
					// Get first eligible requirement
					for _, reqId := range cycleTutor.RequirementIds {
						cycleReq := requirementMap[reqId]
						if !cycleReq.Processed {
							cycleIds = append(cycleIds, cycleReq.PageId)
							if _, ok := cycleReqMap[cycleReq.PageId]; ok {
								continueCycle = false
							} else {
								cycleReqMap[cycleReq.PageId] = true
							}
							break
						}
					}
				}
				c.Debugf("CYCLE: %v", cycleIds)

				req.Processed = true
				if req.BestTutorId == "" {
					if len(req.TutorIds) > 0 {
						// Just take the first tutor
						req.BestTutorId = req.TutorIds[0]
					}
					req.Cost = PenaltyCost
				}
				req.Cost += req.LensIndex * LensCost
				core.AddPageToMap(req.PageId, returnData.PageMap, loadOptions)
				c.Debugf("Requirement '%s' (tutors: %v) forced to processed with cost %d and best tutor '%s'", req.PageId, req.TutorIds, req.Cost, req.BestTutorId)
				break
			}
		}

		graphChanged = false
		// Process all requirements
		for _, req := range requirementMap {
			if req.Processed {
				continue
			}
			// We can mark a requirement processed when we processed all its tutors
			allTutorsProcessed := true
			for _, tutorId := range req.TutorIds {
				tutor := tutorMap[tutorId]
				if !tutor.Processed {
					allTutorsProcessed = false
					continue
				}
				if req.Cost > tutor.Cost {
					req.Cost = tutor.Cost
					req.BestTutorId = tutorId
				}
			}
			if allTutorsProcessed {
				req.Cost += req.LensIndex * LensCost
				req.Processed = true
				graphChanged = true
				core.AddPageToMap(req.PageId, returnData.PageMap, loadOptions)
				c.Debugf("Requirement '%s' (tutors: %v) processed with cost %d and best tutor '%s'", req.PageId, req.TutorIds, req.Cost, req.BestTutorId)
			}
		}

		// Process all tutors
		for _, tutor := range tutorMap {
			if tutor.Processed {
				continue
			}
			// We can mark a tutor processed when we processed all its requirements
			allReqsProcessed := true
			tutor.Cost = 0
			for _, reqId := range tutor.RequirementIds {
				requirement := requirementMap[reqId]
				if !requirement.Processed {
					allReqsProcessed = false
					continue
				}
				tutor.Cost += requirement.Cost
			}
			if allReqsProcessed {
				tutor.Cost += tutor.LensIndex * LensCost
				tutor.Processed = true
				tutor.Cost++
				tutor.RequirementMap = requirementMap
				sort.Sort(tutor)
				graphChanged = true
				core.AddPageToMap(tutor.PageId, returnData.PageMap, loadOptions)
				c.Debugf("Tutor '%s' processed with cost %d and reqs %v", tutor.PageId, tutor.Cost, tutor.RequirementIds)
			}
		}
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["tutorMap"] = tutorMap
	returnData.ResultMap["requirementMap"] = requirementMap
	return pages.StatusOK(returnData.toJson())
}
