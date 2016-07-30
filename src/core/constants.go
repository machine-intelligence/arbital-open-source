package core

type ContentRequestType string

const (
	SlowDown      ContentRequestType = "slowDown"
	SpeedUp       ContentRequestType = "speedUp"
	LessTechnical ContentRequestType = "lessTechnical"
	MoreWords     ContentRequestType = "moreWords"
	MoreTechnical ContentRequestType = "moreTechnical"
	FewerWords    ContentRequestType = "fewerWords"
	ImproveStub   ContentRequestType = "improveStub"
)

var _allContentRequestTypes = []ContentRequestType{
	SlowDown,
	SpeedUp,
	LessTechnical,
	MoreWords,
	MoreTechnical,
	FewerWords,
	ImproveStub,
}

func IsContentRequestTypeValid(t ContentRequestType) bool {
	for _, v := range _allContentRequestTypes {
		if t == v {
			return true
		}
	}
	return false
}
