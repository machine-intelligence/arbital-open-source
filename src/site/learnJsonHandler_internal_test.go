package site

import (
	"testing"

	"zanaduu3/src/core"
	"zanaduu3/src/logger"
)

// Make sure we can match a tutor with a requirement
func SimpleTest(t *testing.T) {
	pageIds := []string{"1"}
	requirementMap := map[string]*requirementNode{
		"1": &requirementNode{PageId: "1", TutorIds: []string{"2"}},
	}
	tutorMap := map[string]*tutorNode{
		"2": &tutorNode{PageId: "2"},
	}
	loadOptions := core.EmptyLoadOptions
	returnData := core.NewHandlerData(nil, true)
	computeLearningPath(logger.Glogger{}, pageIds, requirementMap, tutorMap, loadOptions, returnData)
	if requirementMap["1"].BestTutorId != "2" {
		t.Error("Invalid best tutor")
	}
}
