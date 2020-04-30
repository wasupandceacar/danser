package beatmap

import (
	"danser/beatmap/objects"
	"danser/bmath"
	. "danser/osuconst"
	"danser/settings"
	"math"
)

//Original code by: https://github.com/ppy/osu/blob/master/osu.Game.Rulesets.Osu/Beatmaps/OsuBeatmapProcessor.cs

func isSpinnerBreak(obj objects.BaseObject) bool {
	_, ok2 := obj.(*objects.Pause)
	return ok2
}

func isSlider(obj objects.BaseObject) bool {
	_, ok1 := obj.(*objects.Slider)
	return ok1
}

func difficultyRate(diff, min, mid, max float64) float64 {
	if diff > 5 {
		return mid + (max-mid)*(diff-5)/5
	}
	if diff < 5 {
		return mid - (mid-min)*(5-diff)/5
	}
	return mid
}

// OD规范为带0.5的向下取整小数
func AdjustOD(OD float64) float64 {
	return math.Floor(OD+0.5) - 0.5
}

func calculateStackLeniency(b *BeatMap) {
	stack_distance := 3.0

	preempt := difficultyRate(b.ApproachRate, 1800, 1200, 450)
	b.ARms = preempt
	b.FadeIn = difficultyRate(b.ApproachRate, 1200, 800, 300)
	// 加入OD
	b.OD300 = AdjustOD(HITWINDOW_300_BASE - (b.OverallDifficulty * HITWINDOW_300_MULT) + HITWINDOW_OFFSET)
	b.OD100 = AdjustOD(HITWINDOW_100_BASE - (b.OverallDifficulty * HITWINDOW_100_MULT) + HITWINDOW_OFFSET)
	b.OD50 = AdjustOD(HITWINDOW_50_BASE - (b.OverallDifficulty * HITWINDOW_50_MULT) + HITWINDOW_OFFSET)
	b.ODMiss = AdjustOD(HITWINDOW_MISS + HITWINDOW_OFFSET)
	hitObjects := b.HitObjects

	if !settings.Objects.StackEnabled {
		return
	}

	for _, v := range hitObjects {
		v.GetBasicData().StackIndex = 0
	}

	extendedEndIndex := len(hitObjects) - 1
	for i := len(hitObjects) - 1; i >= 0; i-- {
		stackBaseIndex := i

		for n := stackBaseIndex + 1; n < len(hitObjects); n++ {

			stackBaseObject := hitObjects[stackBaseIndex]
			if isSpinnerBreak(stackBaseObject) {
				break
			}

			objectN := hitObjects[n]
			if isSpinnerBreak(objectN) {
				continue
			}

			stackThreshold := preempt * b.StackLeniency

			if objectN.GetBasicData().StartTime-stackBaseObject.GetBasicData().EndTime > int64(stackThreshold) {
				break
			}

			if stackBaseObject.GetBasicData().StartPos.Dst(objectN.GetBasicData().StartPos) < stack_distance || isSlider(stackBaseObject) && stackBaseObject.GetBasicData().EndPos.Dst(objectN.GetBasicData().StartPos) < stack_distance {
				stackBaseIndex = n
				objectN.GetBasicData().StackIndex = 0
			}
		}

		if stackBaseIndex > extendedEndIndex {
			extendedEndIndex = stackBaseIndex
			if extendedEndIndex == len(hitObjects)-1 {
				break
			}
		}

	}

	extendedStartIndex := 0
	for i := extendedEndIndex; i > 0; i-- {
		n := i

		objectI := hitObjects[i]

		if objectI.GetBasicData().StackIndex != 0 || isSpinnerBreak(objectI) {
			continue
		}

		stackThreshold := preempt * b.StackLeniency

		if _, ok := objectI.(*objects.Circle); ok {
			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinnerBreak(objectN) {
					continue
				}

				if objectI.GetBasicData().StartTime-objectN.GetBasicData().EndTime > int64(stackThreshold) {
					break
				}

				if n < extendedStartIndex {
					objectN.GetBasicData().StackIndex = 0
					extendedStartIndex = n
				}

				if isSlider(objectN) && objectN.GetBasicData().EndPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					offset := objectI.GetBasicData().StackIndex - objectN.GetBasicData().StackIndex + 1
					for j := n + 1; j <= i; j++ {
						objectJ := hitObjects[j]
						if objectN.GetBasicData().EndPos.Dst(objectJ.GetBasicData().StartPos) < stack_distance {
							objectJ.GetBasicData().StackIndex -= offset
						}
					}

					break
				}

				if objectN.GetBasicData().StartPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					objectN.GetBasicData().StackIndex = objectI.GetBasicData().StackIndex + 1
					objectI = objectN
				}

			}
		} else if isSlider(objectI) {

			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinnerBreak(objectN) {
					continue
				}

				if objectI.GetBasicData().StartTime-objectN.GetBasicData().StartTime > int64(stackThreshold) {
					break
				}

				if objectN.GetBasicData().StartPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					objectN.GetBasicData().StackIndex = objectI.GetBasicData().StackIndex + 1
					objectI = objectN
				}

			}
		}

	}

	scale := (1.0 - 0.7*(b.CircleSize-5)/5) / 2

	for _, v := range hitObjects {
		if !isSpinnerBreak(v) {
			sc := float64(v.GetBasicData().StackIndex) * scale * -6.4
			v.GetBasicData().StackOffset = bmath.NewVec2d(sc, sc)
			v.GetBasicData().StartPos = v.GetBasicData().StartPos.Add(v.GetBasicData().StackOffset)
			v.GetBasicData().EndPos = v.GetBasicData().EndPos.Add(v.GetBasicData().StackOffset)
		}
	}
}

// 偏移考虑HR和EZ
func calculateStackLeniencywithMods(b *BeatMap, isHR bool, isEZ bool) {
	stack_distance := 3.0

	newAR := b.ApproachRate
	if isHR {
		newAR = math.Min(b.ApproachRate*AR_HR_HENSE, AR_MAX)
	}
	if isEZ {
		newAR = math.Min(b.ApproachRate*AR_EZ_HENSE, AR_MAX)
	}
	preempt := difficultyRate(newAR, 1800, 1200, 450)
	b.ARms = preempt
	b.FadeIn = difficultyRate(newAR, 1200, 800, 300)
	// 加入OD
	b.OD300 = AdjustOD(HITWINDOW_300_BASE - (b.OverallDifficulty * HITWINDOW_300_MULT) + HITWINDOW_OFFSET)
	b.OD100 = AdjustOD(HITWINDOW_100_BASE - (b.OverallDifficulty * HITWINDOW_100_MULT) + HITWINDOW_OFFSET)
	b.OD50 = AdjustOD(HITWINDOW_50_BASE - (b.OverallDifficulty * HITWINDOW_50_MULT) + HITWINDOW_OFFSET)
	b.ODMiss = AdjustOD(HITWINDOW_MISS + HITWINDOW_OFFSET)
	hitObjects := b.HitObjects

	if !settings.Objects.StackEnabled {
		return
	}

	for _, v := range hitObjects {
		v.GetBasicData().StackIndex = 0
	}

	extendedEndIndex := len(hitObjects) - 1
	for i := len(hitObjects) - 1; i >= 0; i-- {
		stackBaseIndex := i

		for n := stackBaseIndex + 1; n < len(hitObjects); n++ {

			stackBaseObject := hitObjects[stackBaseIndex]
			if isSpinnerBreak(stackBaseObject) {
				break
			}

			objectN := hitObjects[n]
			if isSpinnerBreak(objectN) {
				continue
			}

			stackThreshold := preempt * b.StackLeniency

			if objectN.GetBasicData().StartTime-stackBaseObject.GetBasicData().EndTime > int64(stackThreshold) {
				break
			}

			if stackBaseObject.GetBasicData().StartPos.Dst(objectN.GetBasicData().StartPos) < stack_distance || isSlider(stackBaseObject) && stackBaseObject.GetBasicData().EndPos.Dst(objectN.GetBasicData().StartPos) < stack_distance {

				stackBaseIndex = n
				objectN.GetBasicData().StackIndex = 0
			}
		}

		if stackBaseIndex > extendedEndIndex {
			extendedEndIndex = stackBaseIndex
			if extendedEndIndex == len(hitObjects)-1 {
				break
			}
		}

	}

	extendedStartIndex := 0
	for i := extendedEndIndex; i > 0; i-- {
		n := i

		objectI := hitObjects[i]

		if objectI.GetBasicData().StackIndex != 0 || isSpinnerBreak(objectI) {
			continue
		}

		stackThreshold := preempt * b.StackLeniency

		if _, ok := objectI.(*objects.Circle); ok {
			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinnerBreak(objectN) {
					continue
				}

				if objectI.GetBasicData().StartTime-objectN.GetBasicData().EndTime > int64(stackThreshold) {
					break
				}

				if n < extendedStartIndex {
					objectN.GetBasicData().StackIndex = 0
					extendedStartIndex = n
				}

				if isSlider(objectN) && objectN.GetBasicData().EndPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					offset := objectI.GetBasicData().StackIndex - objectN.GetBasicData().StackIndex + 1
					for j := n + 1; j <= i; j++ {
						objectJ := hitObjects[j]
						if objectN.GetBasicData().EndPos.Dst(objectJ.GetBasicData().StartPos) < stack_distance {
							objectJ.GetBasicData().StackIndex -= offset
						}
					}

					break
				}

				if objectN.GetBasicData().StartPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					objectN.GetBasicData().StackIndex = objectI.GetBasicData().StackIndex + 1
					objectI = objectN
				}

			}
		} else if isSlider(objectI) {

			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinnerBreak(objectN) {
					continue
				}

				if objectI.GetBasicData().StartTime-objectN.GetBasicData().StartTime > int64(stackThreshold) {
					break
				}

				if objectN.GetBasicData().StartPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					objectN.GetBasicData().StackIndex = objectI.GetBasicData().StackIndex + 1
					objectI = objectN
				}

			}
		}

	}

	scale := (1.0 - 0.7*(b.CircleSize-5)/5) / 2

	for _, v := range hitObjects {
		if !isSpinnerBreak(v) {
			sc := float64(v.GetBasicData().StackIndex) * scale * BASE_STACK_OFFSET
			v.GetBasicData().StackOffset = bmath.NewVec2d(sc, sc)
			v.GetBasicData().StartPos = v.GetBasicData().StartPos.Add(v.GetBasicData().StackOffset)
			v.GetBasicData().EndPos = v.GetBasicData().EndPos.Add(v.GetBasicData().StackOffset)
		}
	}
}
