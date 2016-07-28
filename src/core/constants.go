package core

type ContentRequestType string

const (
	SlowDown      ContentRequestType = "slowDown"
	SpeedUp       ContentRequestType = "speedUp"
	LessTechnical ContentRequestType = "lessTechnical"
	MoreWords     ContentRequestType = "moreWords"
	ImproveStub   ContentRequestType = "improveStub"
)

var _allContentRequestTypes = []ContentRequestType{
	SlowDown,
	SpeedUp,
	LessTechnical,
	MoreWords,
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
