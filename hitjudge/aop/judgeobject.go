package aop

import (
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/osuconst"
)

type JudgeObjectType int

const (
	//normal judge with hitwindow and notelock
	NormalJudge JudgeObjectType = 1

	//instant judge without hitwindow or notelock
	//mainly used for slider ticks
	InstantJudge JudgeObjectType = 2

	//reserved for spinners
	ContinuousJudge JudgeObjectType = 3
)

type JudgeObject struct {
	JudgeTime     int64
	JudgePosition bmath.Vector2d
	JudgeType     JudgeObjectType

	//the index of the slider if it belongs to a slider
	SliderIndex int64
}

func ConvertToJudgeObjects(hitObjects []objects.BaseObject) []JudgeObject {
	var result []JudgeObject
	var sliderIndex int64 = 0
	for i := 0; i < len(hitObjects); i++ {
		hitObject := hitObjects[i]
		if hitObject != nil {
			if slider, succeeded := hitObject.(*objects.Slider); succeeded {
				result = append(result, convertSlider(slider, sliderIndex)...)
				sliderIndex++
			} else if circle, succeeded := hitObject.(*objects.Circle); succeeded {
				result = append(result, convertCircle(circle))
			}
		}
	}
	return result
}

func convertSlider(slider *objects.Slider, sliderIndex int64) []JudgeObject {
	var result []JudgeObject
	dummyCircles := slider.GetAsDummyCircles()
	if len(dummyCircles) > 0 {
		result = append(result, JudgeObject{
			JudgeTime:     dummyCircles[0].GetBasicData().StartTime,
			JudgePosition: dummyCircles[0].GetBasicData().StartPos,
			JudgeType:     NormalJudge,
			SliderIndex:   sliderIndex,
		})
	}
	for i := 1; i < len(dummyCircles)-1; i++ {
		result = append(result, JudgeObject{
			JudgeTime:     dummyCircles[i].GetBasicData().StartTime,
			JudgePosition: dummyCircles[i].GetBasicData().StartPos,
			JudgeType:     InstantJudge,
			SliderIndex:   sliderIndex,
		})
	}
	result = append(result, convertSliderTail(slider, sliderIndex))
	return result
}

func convertSliderTail(slider *objects.Slider, sliderIndex int64) JudgeObject {
	duration := slider.GetBasicData().EndTime - slider.GetBasicData().StartTime
	if duration < 2*osuconst.SLIDER_TAIL_JUDGE_OFFSET {
		return JudgeObject{
			JudgeTime:     slider.GetBasicData().StartTime + duration/2,
			JudgePosition: slider.GetPointAt(slider.GetBasicData().StartTime + duration/2),
			JudgeType:     InstantJudge,
			SliderIndex:   sliderIndex,
		}
	} else {
		return JudgeObject{
			JudgeTime:     slider.GetBasicData().EndTime - osuconst.SLIDER_TAIL_JUDGE_OFFSET,
			JudgePosition: slider.GetPointAt(slider.GetBasicData().EndTime - osuconst.SLIDER_TAIL_JUDGE_OFFSET),
			JudgeType:     InstantJudge,
			SliderIndex:   sliderIndex,
		}
	}
}

func convertCircle(circle *objects.Circle) JudgeObject {
	return JudgeObject{
		JudgeTime:     circle.GetBasicData().StartTime,
		JudgePosition: circle.GetBasicData().StartPos,
		JudgeType:     NormalJudge,
		SliderIndex:   -1,
	}
}
