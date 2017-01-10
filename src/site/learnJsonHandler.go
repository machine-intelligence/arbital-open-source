// learnJsonHandler.go returns the path of pages needed for understanding some pages

package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/logger"
	"zanaduu3/src/pages"
)

var learnHandler = siteHandler{
	URI:         "/json/learn/",
	HandlerFunc: learnJSONHandler,
}

type learnJSONData struct {
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
	PageID    string `json:"pageId"`
	LensIndex int    `json:"-"`
	// Which pages can teach this requirement
	TutorIDs []string `json:"tutorIds"`
	// Best tutor
	BestTutorID string `json:"bestTutorId"`
	// Cost assigned to learning this node
	Cost int `json:"cost"`
	// Set to true when the node has been processed
	Processed bool `json:"-"`
}

// Page that will teach the user about stuff.
type tutorNode struct {
	PageID    string `json:"pageId"`
	LensIndex int    `json:"-"`
	// To read this page, the user needs these requirements
	RequirementIDs []string `json:"requirementIds"`
	// Cost assigned to learning this node
	Cost int `json:"cost"`
	// Set to true when the node has been processed
	Processed bool `json:"-"`
	// Set when a requsite doesn't have a tutor, so we pretend it teaches itself.
	MadeUp bool `json:"madeUp"`

	// Need to set this map for sorting to work
	RequirementMap map[string]*requirementNode `json:"-"`
}

// Sort node's requirements
func (t *tutorNode) Len() int { return len(t.RequirementIDs) }
func (t *tutorNode) Swap(i, j int) {
	t.RequirementIDs[i], t.RequirementIDs[j] = t.RequirementIDs[j], t.RequirementIDs[i]
}
func (t *tutorNode) Less(i, j int) bool {
	return t.RequirementMap[t.RequirementIDs[i]].Cost < t.RequirementMap[t.RequirementIDs[j]].Cost
}

func newRequirementNode(pageID string) *requirementNode {
	return &requirementNode{PageID: pageID, TutorIDs: make([]string, 0), Cost: 10000000}
}

func newTutorNode(pageID string) *tutorNode {
	return &tutorNode{PageID: pageID, RequirementIDs: make([]string, 0)}
}

func learnJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	c := params.C
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data learnJSONData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if len(data.PageAliases) <= 0 {
		return pages.Success(nil)
	}

	// Aliases might have various prefixes. Process them.
	pageAliases := make([]string, 0)
	aliasOptionsMap := make(map[string]*learnOption)
	for _, alias := range data.PageAliases {
		option := &learnOption{MustBeWanted: data.OnlyWanted}
		actualAlias := strings.ToLower(alias)
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
		pageAliases = append(pageAliases, actualAlias)
		aliasOptionsMap[actualAlias] = option
	}

	// Convert aliases to page ids
	aliasToIDMap, err := core.LoadAliasToPageIDMap(db, u, pageAliases)
	if err != nil {
		return pages.Fail("error while loading group members", err)
	}

	// Populate the data structures we need keyed on page id (instead of alias)
	optionsMap := make(map[string]*learnOption)
	pageIDs := make([]string, 0)
	for _, alias := range pageAliases {
		pageID := aliasToIDMap[alias]
		if !core.IsIDValid(pageID) {
			return pages.Fail(fmt.Sprintf("Invalid page id: %s", pageID), nil).Status(http.StatusBadRequest)
		}
		pageIDs = append(pageIDs, pageID)
		optionsMap[pageID] = aliasOptionsMap[alias]
	}

	// Remove requirements that the user already has
	masteryMap := make(map[string]*core.Mastery)
	userID := u.GetSomeID()
	if len(pageIDs) > 0 && userID != "" {
		rows := database.NewQuery(`
			SELECT masteryId,wants,has
			FROM userMasteryPairs
			WHERE userId=?`, userID).Add(`AND masteryId IN`).AddArgsGroupStr(pageIDs).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var masteryID string
			var wants, has bool
			err := rows.Scan(&masteryID, &wants, &has)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			masteryMap[masteryID] = &core.Mastery{PageID: masteryID, Wants: wants, Has: has}
			return nil
		})
		if err != nil {
			return pages.Fail("Error while checking if already knows", err)
		}
	}

	// What to load for the pages
	loadOptions := (&core.PageLoadOptions{
		Tags: true,
	}).Add(core.TitlePlusLoadOptions)

	// Track which requirements we need to process in the next step
	requirementIDs := make([]string, 0)
	for _, pageID := range pageIDs {
		core.AddPageToMap(pageID, returnData.PageMap, loadOptions)
		mastery, ok := masteryMap[pageID]
		add := !ok || !mastery.Has
		if ok && optionsMap[pageID].MustBeWanted && !mastery.Wants {
			add = false
		}
		if add {
			requirementIDs = append(requirementIDs, pageID)
		}
	}
	// Leave only the page ids we need to process
	pageIDs = append(make([]string, 0), requirementIDs...)

	// Which tutor pages to load in the next step
	tutorIDs := make([]string, 0)

	// Create the maps which will store all the nodes: page id -> node
	tutorMap := make(map[string]*tutorNode)
	requirementMap := make(map[string]*requirementNode)
	for _, reqID := range requirementIDs {
		requirementMap[reqID] = newRequirementNode(reqID)
	}

	// Add a tutor for the tutorMap
	var addTutor = func(parentID, childID string, lensIndex int) {
		// Get the requirement node and update its tutors
		requirementNode := requirementMap[parentID]
		requirementNode.TutorIDs = append(requirementNode.TutorIDs, childID)
		c.Infof("Updated requirement node: %+v", requirementNode)
		// Recursively load requirements for the tutor, unless we already processed it
		if _, ok := tutorMap[childID]; !ok {
			tutorIDs = append(tutorIDs, childID)
			tutorMap[childID] = newTutorNode(childID)
			tutorMap[childID].LensIndex = lensIndex
		}
	}

	// Recursively find which pages the user has to read
	for maxCount := 0; len(requirementIDs) > 0 && maxCount < 20; maxCount++ {
		c.Infof("RequirementIds: %+v", requirementIDs)
		// Load which pages teach the requirements
		tutorIDs = make([]string, 0)
		rows := database.NewQuery(`
			SELECT pp.parentId,pp.childId,IFNULL(l.lensIndex,0)
			FROM pagePairs AS pp
			JOIN pageInfos AS pi
			ON (pp.childId=pi.pageId)
			LEFT JOIN lenses AS l
			ON (pi.pageId=l.lensId)
			WHERE pp.parentId IN`).AddArgsGroupStr(requirementIDs).Add(`
				AND pp.type=?`, core.SubjectPagePairType).Add(`
				AND`).AddPart(core.PageInfosFilter(u)).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentID, childID string
			var lensIndex int
			err := rows.Scan(&parentID, &childID, &lensIndex)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			c.Infof("Found tutor: %s %s", parentID, childID)
			addTutor(parentID, childID, lensIndex)
			return nil
		})
		if err != nil {
			return pages.Fail("Error while loading tutors", err)
		}
		// If we haven't found a tutor for a page we want to learn, we'll pretend
		// that the page can teach itself.
		for reqID, requirementNode := range requirementMap {
			_, ok := tutorMap[reqID]
			if len(requirementNode.TutorIDs) <= 0 && !ok {
				c.Infof("No tutor found for %s, so we are making it teach itself.", reqID)
				addTutor(reqID, reqID, 0)
				tutorMap[reqID].MadeUp = true
			}
		}
		c.Infof("TutorIds: %+v", tutorIDs)
		if len(tutorIDs) <= 0 {
			break
		}

		// Load the requirements for the tutors
		requirementIDs = make([]string, 0)
		rows = database.NewQuery(`
			SELECT pp.parentId,pp.childId,IFNULL(l.lensIndex,0)
			FROM pagePairs AS pp
			JOIN pageInfos AS pi
			ON (pp.parentId=pi.pageId)
			LEFT JOIN userMasteryPairs AS mp
			ON (pp.parentId=mp.masteryId AND mp.userId=?)`, userID).Add(`
			LEFT JOIN lenses AS l
			ON (pi.pageId=l.lensId)
			WHERE pp.childId IN`).AddArgsGroupStr(tutorIDs).Add(`
				AND pp.type=?`, core.RequirementPagePairType).Add(`
				AND (NOT mp.has OR ISNULL(mp.has))
				AND`).AddPart(core.PageInfosFilter(u)).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentID, childID string
			var lensIndex int
			err := rows.Scan(&parentID, &childID, &lensIndex)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			c.Infof("Found requirement: %s %s", parentID, childID)

			// Get the tutor node and update its requirements
			tutorNode := tutorMap[childID]
			tutorNode.RequirementIDs = append(tutorNode.RequirementIDs, parentID)
			c.Infof("Updated tutor node: %+v", tutorNode)
			if _, ok := requirementMap[parentID]; !ok {
				requirementIDs = append(requirementIDs, parentID)
				requirementMap[parentID] = newRequirementNode(parentID)
				requirementMap[parentID].LensIndex = lensIndex
			}
			return nil
		})
		if err != nil {
			return pages.Fail("Error while loading requirements", err)
		}
		if maxCount >= 15 {
			c.Warningf("Max count is close to maximum: %d", maxCount)
		}
	}

	computeLearningPath(c, pageIDs, requirementMap, tutorMap, loadOptions, returnData)

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["tutorMap"] = tutorMap
	returnData.ResultMap["requirementMap"] = requirementMap
	returnData.ResultMap["pageIds"] = pageIDs
	returnData.ResultMap["optionsMap"] = optionsMap
	return pages.Success(returnData)
}

func computeLearningPath(pl logger.Logger,
	pageIDs []string,
	requirementMap map[string]*requirementNode,
	tutorMap map[string]*tutorNode,
	loadOptions *core.PageLoadOptions,
	returnData *core.CommonHandlerData) {

	pl.Infof("================ COMPUTING LEARNING PATH  ==================")

	// Mark all requirements with no teachers as processed. Also set the initial cost.
	for _, req := range requirementMap {
		req.Cost = 10000000
		if len(req.TutorIDs) > 0 {
			continue
		}
		req.Cost = PenaltyCost
		req.Processed = true
		core.AddPageToMap(req.PageID, returnData.PageMap, loadOptions)
		pl.Infof("Requirement '%s' pre-processed with cost %d", req.PageID, req.Cost)
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
				cycleIDs := make([]string, 0)
				cycleIDs = append(cycleIDs, req.PageID)
				cycleReqMap := make(map[string]bool) // store all requirements we've met
				cycleReqMap[req.PageID] = true
				continueCycle := true
				for continueCycle {
					// Get first eligible tutor
					var cycleTutor *tutorNode
					for _, tutorID := range req.TutorIDs {
						cycleTutor = tutorMap[tutorID]
						if !cycleTutor.Processed {
							cycleIDs = append(cycleIDs, cycleTutor.PageID)
							break
						}
					}
					// Get first eligible requirement
					for _, reqID := range cycleTutor.RequirementIDs {
						req = requirementMap[reqID]
						if !req.Processed {
							cycleIDs = append(cycleIDs, req.PageID)
							if _, ok := cycleReqMap[req.PageID]; ok {
								continueCycle = false
							} else {
								cycleReqMap[req.PageID] = true
							}
							break
						}
					}
				}
				pl.Infof("CYCLE: %v", cycleIDs)

				// Force the picked requirement to be processed
				req.Processed = true
				if req.BestTutorID == "" {
					if len(req.TutorIDs) > 0 {
						// Just take the first tutor
						req.BestTutorID = req.TutorIDs[0]
					}
					req.Cost = PenaltyCost
				}
				req.Cost += req.LensIndex * LensCost
				core.AddPageToMap(req.PageID, returnData.PageMap, loadOptions)
				pl.Infof("Requirement '%s' (tutors: %v) forced to processed with cost %d and best tutor '%s'", req.PageID, req.TutorIDs, req.Cost, req.BestTutorID)
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
			for _, tutorID := range req.TutorIDs {
				tutor := tutorMap[tutorID]
				if !tutor.Processed {
					allTutorsProcessed = false
					continue
				}
				if req.Cost > tutor.Cost {
					req.Cost = tutor.Cost
					req.BestTutorID = tutorID
				}
			}
			if allTutorsProcessed {
				req.Cost += req.LensIndex * LensCost
				req.Processed = true
				graphChanged = true
				core.AddPageToMap(req.PageID, returnData.PageMap, loadOptions)
				pl.Infof("Requirement '%s' (tutors: %v) processed with cost %d and best tutor '%s'", req.PageID, req.TutorIDs, req.Cost, req.BestTutorID)
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
			for _, reqID := range tutor.RequirementIDs {
				requirement := requirementMap[reqID]
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
				core.AddPageToMap(tutor.PageID, returnData.PageMap, loadOptions)
				pl.Infof("Tutor '%s' processed with cost %d and reqs %v", tutor.PageID, tutor.Cost, tutor.RequirementIDs)
			}
		}

		// Check if we are done
		done = true
		for _, pageID := range pageIDs {
			if !requirementMap[pageID].Processed {
				done = false
				break
			}
		}
	}
}
