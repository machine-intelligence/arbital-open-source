package core

type ContentRequestType string

const (
	SlowDown                        ContentRequestType = "slowDown"
	SpeedUp                         ContentRequestType = "speedUp"
	LessTechnical                   ContentRequestType = "lessTechnical"
	MoreWords                       ContentRequestType = "moreWords"
	MoreTechnical                   ContentRequestType = "moreTechnical"
	FewerWords                      ContentRequestType = "fewerWords"
	ImproveStub                     ContentRequestType = "improveStub"
	TeachLooseUnderstanding         ContentRequestType = "teachLooseUnderstanding"
	TeachBasicUnderstanding         ContentRequestType = "teachBasicUnderstanding"
	TeachTechnicalUnderstanding     ContentRequestType = "teachTechnicalUnderstanding"
	TeachResearchLevelUnderstanding ContentRequestType = "teachResearchLevelUnderstanding"
	BoostLooseUnderstanding         ContentRequestType = "boostLooseUnderstanding"
	BoostBasicUnderstanding         ContentRequestType = "boostBasicUnderstanding"
	BoostTechnicalUnderstanding     ContentRequestType = "boostTechnicalUnderstanding"
	BoostResearchLevelUnderstanding ContentRequestType = "boostResearchLevelUnderstanding"
)

var _allContentRequestTypes = []ContentRequestType{
	SlowDown,
	SpeedUp,
	LessTechnical,
	MoreWords,
	MoreTechnical,
	FewerWords,
	ImproveStub,
	TeachLooseUnderstanding,
	TeachBasicUnderstanding,
	TeachTechnicalUnderstanding,
	TeachResearchLevelUnderstanding,
	BoostLooseUnderstanding,
	BoostBasicUnderstanding,
	BoostTechnicalUnderstanding,
	BoostResearchLevelUnderstanding,
}

func IsContentRequestTypeValid(t ContentRequestType) bool {
	for _, v := range _allContentRequestTypes {
		if t == v {
			return true
		}
	}
	return false
}
