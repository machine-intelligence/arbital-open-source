package core

type ContentRequestType string

const (
	SlowDown ContentRequestType = "slowDown"
	SpeedUp  ContentRequestType = "speedUp"
)

var _allContentRequestTypes = []ContentRequestType{SlowDown, SpeedUp}

func IsContentRequestTypeValid(t ContentRequestType) bool {
	for _, v := range _allContentRequestTypes {
		if t == v {
			return true
		}
	}
	return false
}
