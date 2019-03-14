package objects

import (
	"strconv"
)

func GetObject(data []string, number int64) (BaseObject, int64) {
	objType, _ := strconv.ParseInt(data[3], 10, 64)
	newnumber := number
	if (objType & CIRCLE) > 0 {
		if objType == NEWCIRCLE {
			// 新的combo
			newnumber = 1
		}else {
			// 继续combo
			newnumber += 1
		}
		return NewCircle(data, newnumber), newnumber
	} else if (objType & SLIDER) > 0 {
		if objType == NEWSLIDER {
			// 新的combo
			newnumber = 1
		}else {
			// 继续combo
			newnumber += 1
		}
		sl := NewSlider(data, newnumber)
		if sl == nil {
			return nil, newnumber
		} else {
			return sl, newnumber
		}
	} else if (objType & SPINNNER) > 0 {
		if objType == NEWSPINNNER{
			// 新的combo
			newnumber = 1
		}else {
			// 继续combo
			newnumber += 1
		}
		return NewSpinner(data, newnumber), newnumber
	}
	return nil, newnumber
}

const (
	CIRCLE int64 = 1
	SLIDER int64 = 2
	SPINNNER int64 = 8
	NEWCIRCLE int64 = 5
	NEWSLIDER int64 = 6
	NEWSPINNNER int64 = 12
)
