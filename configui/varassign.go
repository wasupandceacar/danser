package configui

import (
	"github.com/lxn/walk"
	"strconv"
)

func assign(ivarpt interface{}, ivar interface{}, component interface{}) {
	switch ivar.(type) {
	case bool:
		*ivarpt.(*bool) = component.(*walk.CheckBox).Checked()
		break
	case string:
		*ivarpt.(*string) = component.(*walk.LineEdit).Text()
		break
	case int:
		intvar, err := strconv.Atoi(component.(*walk.LineEdit).Text())
		if err != nil {
			panic(err)
		}
		*ivarpt.(*int) = intvar
		break
	case int64:
		intvar, err := strconv.Atoi(component.(*walk.LineEdit).Text())
		if err != nil {
			panic(err)
		}
		*ivarpt.(*int64) = int64(intvar)
		break
	case float64:
		float64var, err := strconv.ParseFloat(component.(*walk.LineEdit).Text(), 64)
		if err != nil {
			panic(err)
		}
		*ivarpt.(*float64) = float64var
		break
	}
	return
}

