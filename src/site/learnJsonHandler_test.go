package site

import (
	"testing"

	"zanaduu3/src/core"
	"zanaduu3/src/logger"
)

// Make sure we can match a tutor with a requirement
func TestOneTutor(t *testing.T) {
	pageIds := []string{"1"}
	requirementMap := map[string]*requirementNode{
		"1": {PageID: "1", TutorIds: []string{"2"}},
	}
	tutorMap := map[string]*tutorNode{
		"2": {PageID: "2"},
	}
	loadOptions := core.EmptyLoadOptions
	returnData := core.NewHandlerData(nil)
	computeLearningPath(logger.StdLogger{}, pageIds, requirementMap, tutorMap, loadOptions, returnData)
	if requirementMap["1"].BestTutorID != "2" {
		t.Errorf("Invalid best tutor: %v, expected 2", requirementMap["1"].BestTutorID)
	}
}

// Make sure we pick the best tutor, given that some of them have unteachable requirements.
func TestTeachableReqs(t *testing.T) {
	pageIds := []string{"1"}
	requirementMap := map[string]*requirementNode{
		"1": {PageID: "1", TutorIds: []string{"3", "4", "5"}},
		"2": {PageID: "2", TutorIds: []string{}},
	}
	tutorMap := map[string]*tutorNode{
		"3": {PageID: "3", RequirementIds: []string{"2"}},
		"4": {PageID: "4"},
		"5": {PageID: "5", RequirementIds: []string{"2"}},
	}
	loadOptions := core.EmptyLoadOptions
	returnData := core.NewHandlerData(nil)
	computeLearningPath(logger.StdLogger{}, pageIds, requirementMap, tutorMap, loadOptions, returnData)
	if requirementMap["1"].BestTutorID != "4" {
		t.Errorf("Invalid best tutor: %v, expected 4", requirementMap["1"].BestTutorID)
	}
}

// Make sure we pick the best tutor, given that some of them have more requirements.
func TestTutorWithLeastReqs(t *testing.T) {
	pageIds := []string{"1"}
	requirementMap := map[string]*requirementNode{
		"1": {PageID: "1", TutorIds: []string{"4", "5", "6"}},
		"2": {PageID: "2", TutorIds: []string{"7"}},
		"3": {PageID: "3", TutorIds: []string{"8"}},
	}
	tutorMap := map[string]*tutorNode{
		"4": {PageID: "4", RequirementIds: []string{"2", "3"}},
		"5": {PageID: "5", RequirementIds: []string{"2"}},
		"6": {PageID: "6", RequirementIds: []string{"2", "3"}},
		"7": {PageID: "7"},
		"8": {PageID: "8"},
	}
	loadOptions := core.EmptyLoadOptions
	returnData := core.NewHandlerData(nil)
	computeLearningPath(logger.StdLogger{}, pageIds, requirementMap, tutorMap, loadOptions, returnData)
	if requirementMap["1"].BestTutorID != "5" {
		t.Errorf("Invalid best tutor: %v, expected 5", requirementMap["1"].BestTutorID)
	}
}
