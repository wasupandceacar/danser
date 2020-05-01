package aop

import (
	"danser/beatmap"
	"danser/bmath"
	"danser/hitjudge"
	"danser/osuconst"
	"danser/score"
	"danser/settings"
	"github.com/Mempler/rplpa"
	oppai "github.com/flesnuk/oppai5"
	"log"
	"math"
	"os"
)

func parseBeatmapWithMods(bm *beatmap.BeatMap, isHR bool, isEZ bool) *beatmap.BeatMap {
	file, err := os.Open(settings.General.OsuSongsDir + bm.Dir + "/" + bm.File)
	if err == nil {
		defer file.Close()
		beatMap := beatmap.ParseBeatMap(file)
		beatmap.ParseObjectsByPath(beatMap, settings.General.OsuSongsDir+bm.Dir+"/"+bm.File, isHR, isEZ)
		return beatMap
	} else {
		panic(err)
	}
}

func configureDifficultySettings(bm *beatmap.BeatMap, rep *rplpa.Replay) DifficultySettings {
	log.Printf("od=%v", bm.OverallDifficulty)
	ds := NewDifficultySettings(bm.OverallDifficulty, bm.CircleSize)
	ds.AdjustForReplay(rep)
	return ds
}

func Judge(bm *beatmap.BeatMap, rep *rplpa.Replay) ([]hitjudge.ObjectResult, []hitjudge.TotalResult) {
	//--------------------------------------------------------------------
	//Step 1: Init
	var objectResults []hitjudge.ObjectResult
	var totalResults []hitjudge.TotalResult
	GetCursorPositionAt := interpolateCubicSpline(rep)
	bm = parseBeatmapWithMods(bm, rep.Mods&osuconst.MOD_HR > 0, rep.Mods&osuconst.MOD_EZ > 0)
	difficultySettings := configureDifficultySettings(bm, rep)
	judgeObjects := ConvertToJudgeObjects(bm.HitObjects)
	currentTime := rep.ReplayData[0].Time + rep.ReplayData[1].Time
	l, r := 0, 0

	log.Printf("cs=%vpx, fcs=%vpx, hwMiss=%vms", difficultySettings.RealCircleSize, difficultySettings.FollowCircleSize, difficultySettings.HitWindowMiss)
	log.Printf("hw300=%vms, hw100=%vms, hw50=%vms", difficultySettings.HitWindow300, difficultySettings.HitWindow100, difficultySettings.HitWindow50)

	//--------------------------------------------------------------------
	//Step 2: Judge each judgeObject
	//TODO: Special case when you release the key while doing sliders

	for i := 2; i < len(rep.ReplayData)-1; i++ {

	}

	for i := 2; i < len(rep.ReplayData)-1; i++ {
		currentTime += rep.ReplayData[i].Time
		for true {
			if r+1 >= len(judgeObjects) {
				break
			}
			nextJudgeObject := &judgeObjects[r+1]
			if nextJudgeObject.JudgeType == InstantJudge {
				log.Printf("Current Frame %v (%vms)", i, currentTime)
				log.Printf("| Adding New Instant Judge, objNum=%v", nextJudgeObject.ObjectIndex)
				r++
			} else if nextJudgeObject.JudgeType == NormalJudge {
				if float64(nextJudgeObject.JudgeTime-currentTime) <= difficultySettings.HitWindowMiss {
					log.Printf("Current Frame %v (%vms)", i, currentTime)
					log.Printf("| deltaTime=%vms", float64(nextJudgeObject.JudgeTime-currentTime))
					log.Printf("| Adding New Normal Judge, objNum=%v", nextJudgeObject.ObjectIndex)
					r++
				} else {
					break
				}
			}
		}
		for true {
			if l > r {
				break
			}
			leftmostJudgeObject := &judgeObjects[l]
			if leftmostJudgeObject.JudgeType == InstantJudge {
				if leftmostJudgeObject.JudgeTime < currentTime {
					log.Printf("Current Frame %v (%vms)", i, currentTime)
					log.Printf("| Outdating an instant judge, objNum=%v", leftmostJudgeObject.ObjectIndex)
					if isAnyKeyPressedAt(rep.ReplayData, leftmostJudgeObject.JudgeTime, i, currentTime) {
						log.Printf("| Instant Hit timing judge passed")
						if judgeFollowCirclePosition(leftmostJudgeObject, GetCursorPositionAt(leftmostJudgeObject.JudgeTime), difficultySettings) {
							log.Printf("| Instant Hit position judge passed")
							log.Printf("| cx=%v, cy=%v, jx=%v, jy=%v", GetCursorPositionAt(leftmostJudgeObject.JudgeTime).X, GetCursorPositionAt(leftmostJudgeObject.JudgeTime).Y, leftmostJudgeObject.JudgePosition.X, leftmostJudgeObject.JudgePosition.Y)
							log.Printf("| dst=%v", GetCursorPositionAt(leftmostJudgeObject.JudgeTime).Dst(leftmostJudgeObject.JudgePosition))
							leftmostJudgeObject.JudgeResult = hitjudge.Hit300
						} else {
							log.Printf("| Instant Hit position judge passed")
							leftmostJudgeObject.JudgeResult = hitjudge.HitMiss
						}
					} else {
						log.Printf("| Instant Hit timing judge failed")
						leftmostJudgeObject.JudgeResult = hitjudge.HitMiss
					}
					l++
				} else {
					break
				}
			} else if leftmostJudgeObject.JudgeType == NormalJudge {
				if float64(currentTime-leftmostJudgeObject.JudgeTime) > difficultySettings.HitWindow50 {
					log.Printf("Current Frame %v (%vms)", i, currentTime)
					log.Printf("| Outdating an normal judge, objNum=%v", leftmostJudgeObject.ObjectIndex)
					if leftmostJudgeObject.JudgeResult == hitjudge.Unjudged {
						log.Printf("| Detected unjudged normal hit, marked as missed")
						leftmostJudgeObject.JudgeResult = hitjudge.HitMiss
					}
					l++
				} else {
					break
				}
			}
		}
		if isAnyNewKeyPressed(rep.ReplayData[i-1].KeyPressed, rep.ReplayData[i].KeyPressed) {
			log.Printf("Current Frame %v (%vms)", i, currentTime)
			log.Printf("| Detected New Key Pressed")
			for j := l; j <= r; j++ {
				currentJudgeObject := &judgeObjects[j]
				if currentJudgeObject.JudgeResult != hitjudge.Unjudged || currentJudgeObject.JudgeType == InstantJudge {
					continue
				}
				if currentJudgeObject.JudgeType == NormalJudge {
					log.Printf("| Processing normal judge, objNum=%v", currentJudgeObject.ObjectIndex)
					if judgeCirclePosition(currentJudgeObject, GetCursorPositionAt(currentTime), difficultySettings) {
						log.Printf("| Position judge passed, cx=%v, cy=%v", GetCursorPositionAt(currentTime).X, GetCursorPositionAt(currentTime).Y)
						log.Printf("| jx=%v, jy=%v, dst=%vpx", currentJudgeObject.JudgePosition.X, currentJudgeObject.JudgePosition.Y, currentJudgeObject.JudgePosition.Dst(GetCursorPositionAt(currentTime)))
						currentJudgeObject.JudgeResult = judgeTiming(currentTime, currentJudgeObject.JudgeTime, difficultySettings)
						log.Printf("| Judged as %v, deltaT=%vms", currentJudgeObject.JudgeResult, currentTime-currentJudgeObject.JudgeTime)
					} else {
						log.Printf("| Position judge failed, cx=%v, cy=%v", GetCursorPositionAt(currentTime).X, GetCursorPositionAt(currentTime).Y)
						log.Printf("| jx=%v, jy=%v, dst=%vpx", currentJudgeObject.JudgePosition.X, currentJudgeObject.JudgePosition.Y, currentJudgeObject.JudgePosition.Dst(GetCursorPositionAt(currentTime)))
						log.Printf("| deltaT=%vms", currentTime-currentJudgeObject.JudgeTime)
					}
					break
				}
			}
		}
	}
	log.Printf("Cleaning...")

	for i := l; i < len(judgeObjects); i++ {
		judgeObjects[i].JudgeResult = hitjudge.HitMiss
	}
	//--------------------------------------------------------------------
	//Step 3: Calculate slider results and hit results

	log.Printf("Calculating...")
	log.Printf("| len(judgeObjects)=%v", len(judgeObjects))

	c300, c100, c50, cMiss, nowCombo, maxCombo := 0, 0, 0, 0, 0, 0

	for i := 0; i < len(judgeObjects); {
		if judgeObjects[i].SliderIndex == -1 {
			objectResults = append(objectResults, hitjudge.ObjectResult{
				JudgePos:  judgeObjects[i].JudgePosition,
				JudgeTime: judgeObjects[i].JudgeTime,
				Result:    judgeObjects[i].JudgeResult,
				IsBreak:   judgeObjects[i].JudgeResult == hitjudge.HitMiss,
			})
			switch judgeObjects[i].JudgeResult {
			case hitjudge.HitMiss:
				cMiss++
				nowCombo = 0
				break
			case hitjudge.Hit300:
				c300++
				nowCombo++
				break
			case hitjudge.Hit100:
				c100++
				nowCombo++
				break
			case hitjudge.Hit50:
				c50++
				nowCombo++
				break
			}
			if nowCombo > maxCombo {
				maxCombo = nowCombo
			}
			i++
		} else {
			var tmp []JudgeObject
			currentSliderIndex := judgeObjects[i].SliderIndex
			currentObjectIndex := judgeObjects[i].ObjectIndex
			log.Printf("| Calculating slider %v, objNum = %v", currentSliderIndex, currentObjectIndex)
			for ; judgeObjects[i].SliderIndex == currentSliderIndex && i < len(judgeObjects); i++ {
				tmp = append(tmp, judgeObjects[i])
			}
			hitResult, dc, mc, isSliderBreak := judgeSlider(tmp)
			objectResults = append(objectResults, hitjudge.ObjectResult{
				JudgePos:  tmp[0].JudgePosition,
				JudgeTime: tmp[len(tmp)-1].JudgeTime,
				Result:    hitResult,
				IsBreak:   isSliderBreak,
			})
			switch hitResult {
			case hitjudge.HitMiss:
				cMiss++
				break
			case hitjudge.Hit300:
				c300++
				break
			case hitjudge.Hit100:
				c100++
				break
			case hitjudge.Hit50:
				c50++
				break
			}
			if isSliderBreak {
				nowCombo = dc
			} else {
				nowCombo += dc
			}
			if nowCombo > maxCombo {
				maxCombo = nowCombo
			}
			if mc > maxCombo {
				maxCombo = mc
			}
		}
		totalResults = append(totalResults, hitjudge.TotalResult{
			N300:   uint16(c300),
			N100:   uint16(c100),
			N50:    uint16(c50),
			Misses: uint16(cMiss),
			Combo:  uint16(maxCombo),
			Mods:   rep.Mods,
			Acc:    calculateAccuracy(c300, c100, c50, cMiss),
			Rank:   calculateRank(c300, c100, c50, cMiss, rep.Mods),
			PP:     calculatePP(settings.General.OsuSongsDir+bm.Dir+"/"+bm.File, c300, c100, c50, cMiss, maxCombo, rep.Mods, int(judgeObjects[i-1].ObjectIndex)),
			UR:     calculateUR(),
		})
	}

	return objectResults, totalResults
}

func isAnyKeyPressed(kp *rplpa.KeyPressed) bool {
	return kp.Key1 || kp.Key2 || kp.LeftClick || kp.RightClick
}

//TODO: Change to use binary search
func isAnyKeyPressedAt(rd []*rplpa.ReplayData, time int64, currentIndex int, currentTime int64) bool {
	if time < currentTime {
		//start from the previous frame
		currentIndex--
		for ; currentIndex >= 2; currentIndex-- {
			currentTime -= rd[currentIndex].Time
			if currentTime <= time {
				return isAnyKeyPressed(rd[currentIndex].KeyPressed)
			}
		}
		return false
	} else if time == currentTime {
		return isAnyKeyPressed(rd[currentIndex].KeyPressed)
	} else {
		//start from the next frame
		currentIndex++
		for ; currentIndex < len(rd); currentIndex++ {
			currentTime += rd[currentIndex].Time
			if currentTime == time {
				return isAnyKeyPressed(rd[currentIndex].KeyPressed)
			} else if currentTime > time && currentIndex-1 >= 2 {
				//use the previous frame
				return isAnyKeyPressed(rd[currentIndex-1].KeyPressed)
			}
		}
		return false
	}
}

func isAnyNewKeyPressed(lastFrame *rplpa.KeyPressed, thisFrame *rplpa.KeyPressed) bool {
	return (!lastFrame.Key1 && thisFrame.Key1) || (!lastFrame.Key2 && thisFrame.Key2) || (!lastFrame.LeftClick && thisFrame.LeftClick) || (!lastFrame.RightClick && thisFrame.RightClick)
}

func judgeTiming(currentTime int64, judgeTime int64, difficultySettings DifficultySettings) hitjudge.HitResult {
	delta := float64(currentTime - judgeTime)
	if delta < 0 {
		delta = math.Abs(delta)
		if delta <= difficultySettings.HitWindow300 {
			return hitjudge.Hit300
		} else if delta <= difficultySettings.HitWindow100 {
			return hitjudge.Hit100
		} else if delta <= difficultySettings.HitWindow50 {
			return hitjudge.Hit50
		} else if delta <= difficultySettings.HitWindowMiss {
			return hitjudge.HitMiss
		} else {
			return hitjudge.Unjudged
		}
	} else {
		delta = math.Abs(delta)
		if delta <= difficultySettings.HitWindow300 {
			return hitjudge.Hit300
		} else if delta <= difficultySettings.HitWindow100 {
			return hitjudge.Hit100
		} else if delta <= difficultySettings.HitWindow50 {
			return hitjudge.Hit50
		} else {
			return hitjudge.HitMiss
		}
	}
}

func judgeCirclePosition(object *JudgeObject, cursorPosition bmath.Vector2d, difficultySettings DifficultySettings) bool {
	if object.JudgePosition.Dst(cursorPosition) <= difficultySettings.RealCircleSize {
		return true
	} else {
		return false
	}
}

func judgeFollowCirclePosition(object *JudgeObject, cursorPosition bmath.Vector2d, difficultySettings DifficultySettings) bool {
	if object.JudgePosition.Dst(cursorPosition) <= difficultySettings.FollowCircleSize {
		return true
	} else {
		return false
	}
}

func judgeSlider(judgeObjects []JudgeObject) (hitjudge.HitResult, int, int, bool) {
	tickMissedCount := 0
	combo := 0
	maxCombo := 0
	isSliderBreak := false
	for i := 0; i < len(judgeObjects)-1; i++ {
		if judgeObjects[i].JudgeResult == hitjudge.HitMiss {
			tickMissedCount++
			combo = 0
			isSliderBreak = true
		} else {
			combo++
			if combo > maxCombo {
				maxCombo = combo
			}
		}
	}
	if judgeObjects[len(judgeObjects)-1].JudgeResult == hitjudge.HitMiss {
		tickMissedCount++
	} else {
		combo++
		if combo > maxCombo {
			maxCombo = combo
		}
	}
	percentage := float64(tickMissedCount) / float64(len(judgeObjects))
	if percentage == 1 {
		return hitjudge.HitMiss, combo, maxCombo, isSliderBreak
	} else if percentage >= 0.5 {
		return hitjudge.Hit50, combo, maxCombo, isSliderBreak
	} else if percentage > 0 {
		return hitjudge.Hit100, combo, maxCombo, isSliderBreak
	} else {
		return hitjudge.Hit300, combo, maxCombo, isSliderBreak
	}
}

func calculateAccuracy(c300, c100, c50, cMiss int) float64 {
	return (float64(c300)*300.0 + float64(c100)*100.0 + float64(c50)*50.0) / (float64(c300+c100+c50+cMiss) * 300.0)
}

func calculateRank(c300, c100, c50, cMiss int, mods uint32) score.Rank {
	countAll := c300 + c100 + c50 + cMiss
	if c300 == countAll {
		if isSilver(mods) {
			return score.SSH
		} else {
			return score.SS
		}
	} else if ((float64(c300) / float64(countAll)) > 0.9) && ((float64(c50) / float64(countAll)) < 0.01) && (cMiss == 0) {
		if isSilver(mods) {
			return score.SH
		} else {
			return score.S
		}
	} else if ((float64(c300) / float64(countAll)) > 0.9) || (((float64(c300) / float64(countAll)) > 0.8) && (cMiss == 0)) {
		return score.A
	} else if ((float64(c300) / float64(countAll)) > 0.8) || (((float64(c300) / float64(countAll)) > 0.7) && (cMiss == 0)) {
		return score.B
	} else if (float64(c300) / float64(countAll)) > 0.6 {
		return score.C
	} else {
		return score.D
	}
}

func isSilver(mods uint32) bool {
	return (mods&osuconst.MOD_HD > 0) || (mods&osuconst.MOD_FL > 0)
}

func calculatePP(filename string, c300, c100, c50, cMiss, combo int, mods uint32, objNum int) oppai.PPv2 {
	//return oppai.PPInfo(score.LoadMapByNum(filename, objNum), &oppai.Parameters{
	//	Combo:  uint16(combo),
	//	Mods:   mods,
	//	N300:   uint16(c300),
	//	N100:   uint16(c100),
	//	N50:    uint16(c50),
	//	Misses: uint16(cMiss),
	//}).PP
	return oppai.PPv2{}
}

//TODO
func calculateUR() float64 {
	return 0.0
}
