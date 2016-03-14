// learnJsonHandler.go returns the path of pages needed for understanding some pages
package site

import (
	"encoding/json"
	"fmt"
	"sort"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/logger"
	"zanaduu3/src/pages"
)

var learnHandler = siteHandler{
	URI:         "/json/learn/",
	HandlerFunc: learnJsonHandler,
}

type learnJsonData struct {
	PageAliases []string
	// If set, only learn pages that are marked as wanted
	OnlyWanted bool
}

const (
	PenaltyCost = 10000
	LensCost    = 10
)

type learnOption struct {
	// If true, the page will be appended in the path after its requisites are learned
	AppendToPath bool `json:"appendToPath"`
	// If true, the page will only be processed if it's wanted
	MustBeWanted bool `json:"mustBeWanted"`
}

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
	returnData := core.NewHandlerData(params.U, true)

	// Decode data
	var data learnJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if len(data.PageAliases) <= 0 {
		return pages.StatusOK(nil)
	}

	// Aliases might have various prefixes. Process them.
	pageAliases := make([]string, len(data.PageAliases), len(data.PageAliases))
	aliasOptionsMap := make(map[string]*learnOption)
	for n, alias := range data.PageAliases {
		option := &learnOption{MustBeWanted: data.OnlyWanted}
		actualAlias := alias
		for {
			char := actualAlias[0]
			if char == '@' {
				option.AppendToPath = true
			} else if char == '$' {
				option.MustBeWanted = true
			} else {
				break
			}
			actualAlias = actualAlias[1:]
		}
		pageAliases[n] = actualAlias
		aliasOptionsMap[actualAlias] = option
	}

	// Convert aliases to page ids
	aliasToIdMap, err := core.LoadAliasToPageIdMap(db, pageAliases)
	if err != nil {
		return pages.HandlerErrorFail("error while loading group members", err)
	}

	// Populate the data structures we need keyed on page id (instead of alias)
	optionsMap := make(map[string]*learnOption)
	pageIds := make([]string, 0)
	for _, alias := range pageAliases {
		pageId := aliasToIdMap[alias]
		if !core.IsIdValid(pageId) {
			return pages.HandlerBadRequestFail(fmt.Sprintf("Invalid page id: %s", pageId), nil)
		}
		pageIds = append(pageIds, pageId)
		optionsMap[pageId] = aliasOptionsMap[alias]
	}

	// Remove requirements that the user already has
	masteryMap := make(map[string]*core.Mastery)
	userId := u.GetSomeId()
	if len(pageIds) > 0 && userId != "" {
		rows := database.NewQuery(`
			SELECT masteryId,wants,has
			FROM userMasteryPairs
			WHERE userId=?`, userId).Add(`AND masteryId IN`).AddArgsGroupStr(pageIds).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var masteryId string
			var wants, has bool
			err := rows.Scan(&masteryId, &wants, &has)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			masteryMap[masteryId] = &core.Mastery{PageId: masteryId, Wants: wants, Has: has}
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("Error while checking if already knows", err)
		}
	}

	// What to load for the pages
	loadOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)

	// Track which requirements we need to process in the next step
	requirementIds := make([]string, 0)
	for _, pageId := range pageIds {
		core.AddPageToMap(pageId, returnData.PageMap, loadOptions)
		mastery, ok := masteryMap[pageId]
		add := !ok || !mastery.Has
		if ok && optionsMap[pageId].MustBeWanted && !mastery.Wants {
			add = false
		}
		if add {
			requirementIds = append(requirementIds, pageId)
		}
	}
	// Leave only the page ids we need to process
	pageIds = append(make([]string, 0), requirementIds...)

	// Which tutor pages to load in the next step
	tutorIds := make([]string, 0)

	// Create the maps which will store all the nodes: page id -> node
	tutorMap := make(map[string]*tutorNode)
	requirementMap := make(map[string]*requirementNode)
	for _, reqId := range requirementIds {
		requirementMap[reqId] = newRequirementNode(reqId)
	}

	// Add a tutor for the tutorMap
	var addTutor = func(parentId, childId string, lensIndex int) {
		// Get the requirement node and update its tutors
		requirementNode := requirementMap[parentId]
		requirementNode.TutorIds = append(requirementNode.TutorIds, childId)
		c.Infof("Updated requirement node: %+v", requirementNode)
		// Recursively load requirements for the tutor, unless we already processed it
		if _, ok := tutorMap[childId]; !ok {
			tutorIds = append(tutorIds, childId)
			tutorMap[childId] = newTutorNode(childId)
			tutorMap[childId].LensIndex = lensIndex
		}
	}

	// Recursively find which pages the user has to read
	for maxCount := 0; len(requirementIds) > 0 && maxCount < 20; maxCount++ {
		c.Infof("RequirementIds: %+v", requirementIds)
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
			c.Infof("Found tutor: %s %s", parentId, childId)
			addTutor(parentId, childId, lensIndex)
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("Error while loading tutors", err)
		}
		if maxCount == 0 {
			// If we haven't found a tutor for a page we want to learn, we'll just say
			// that the page can teach itself.
			for reqId, requirementNode := range requirementMap {
				if len(requirementNode.TutorIds) <= 0 {
					c.Infof("No tutor found for %s, so we are making it teach itself.", reqId)
					addTutor(reqId, reqId, 0)
				}
			}
		}
		c.Infof("TutorIds: %+v", tutorIds)
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
			ON (pp.parentId=mp.masteryId AND mp.userId=?)`, userId).Add(`
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
			c.Infof("Found requirement: %s %s", parentId, childId)

			// Get the tutor node and update its requirements
			tutorNode := tutorMap[childId]
			tutorNode.RequirementIds = append(tutorNode.RequirementIds, parentId)
			c.Infof("Updated tutor node: %+v", tutorNode)
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
		if maxCount >= 15 {
			c.Warningf("Max count is close to maximum: %d", maxCount)
		}
	}

	computeLearningPath(c, pageIds, requirementMap, tutorMap, loadOptions, returnData)

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["tutorMap"] = tutorMap
	returnData.ResultMap["requirementMap"] = requirementMap
	returnData.ResultMap["pageIds"] = pageIds
	returnData.ResultMap["optionsMap"] = optionsMap
	return pages.StatusOK(returnData.ToJson())
}

func computeLearningPath(pl logger.Logger,
	pageIds []string,
	requirementMap map[string]*requirementNode,
	tutorMap map[string]*tutorNode,
	loadOptions *core.PageLoadOptions,
	returnData *core.CommonHandlerData) {

	pl.Infof("================ COMPUTING LEARNING PATH  ==================")

	// Mark all requirements with no teachers as processed
	for _, req := range requirementMap {
		if len(req.TutorIds) > 0 {
			continue
		}
		req.Cost = PenaltyCost
		req.Processed = true
		core.AddPageToMap(req.PageId, returnData.PageMap, loadOptions)
		pl.Infof("Requirement '%s' pre-processed with cost %d", req.PageId, req.Cost)
	}

	done := false
	graphChanged := true
	// Keep processing all the nodes until we processed the node we want to learn
	for !done {
		if !graphChanged {
			// We didn't make any progress in the last iteration, which means there is
			// a cycle. Arbitrarily mark a requirement as processed.
			for _, req := range requirementMap {
				if req.Processed {
					continue
				}

				// Print the cycle, but also find a node that's actually definitely in the cycle
				cycleIds := make([]string, 0)
				cycleIds = append(cycleIds, req.PageId)
				cycleReqMap := make(map[string]bool) // store all requirements we've met
				cycleReqMap[req.PageId] = true
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
						req = requirementMap[reqId]
						if !req.Processed {
							cycleIds = append(cycleIds, req.PageId)
							if _, ok := cycleReqMap[req.PageId]; ok {
								continueCycle = false
							} else {
								cycleReqMap[req.PageId] = true
							}
							break
						}
					}
				}
				pl.Infof("CYCLE: %v", cycleIds)

				// Force the picked requirement to be processed
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
				pl.Infof("Requirement '%s' (tutors: %v) forced to processed with cost %d and best tutor '%s'", req.PageId, req.TutorIds, req.Cost, req.BestTutorId)
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
				pl.Infof("Requirement '%s' (tutors: %v) processed with cost %d and best tutor '%s'", req.PageId, req.TutorIds, req.Cost, req.BestTutorId)
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
				pl.Infof("Tutor '%s' processed with cost %d and reqs %v", tutor.PageId, tutor.Cost, tutor.RequirementIds)
			}
		}

		// Check if we are done
		done = true
		for _, pageId := range pageIds {
			if !requirementMap[pageId].Processed {
				done = false
				break
			}
		}
	}
}
