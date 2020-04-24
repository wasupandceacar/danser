package hitjudge

import (
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/bmath"
	. "danser/osuconst"
	"danser/replay"
	"danser/score"
	"danser/settings"
	"github.com/Mempler/rplpa"
	"github.com/flesnuk/oppai5"
	"log"
	"math"
	"os"
	"sort"
)

var None = rplpa.KeyPressed{
	LeftClick:  false,
	RightClick: false,
	Key1:       false,
	Key2:       false,
}

var UseCSOffset = NO_USE_CS_OFFSET

func ParseMapwithMods(filename string, isHR bool, isEZ bool) *beatmap.BeatMap{
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()
		beatMap := beatmap.ParseBeatMap(file)
		beatmap.ParseObjectsbyPath(beatMap, filename, isHR, isEZ)
		return beatMap
	}else{
		panic(err)
	}
}

func ParseReplay(name string) *rplpa.Replay {
	return replay.ExtractReplay(name)
}

func ParseHits(mapname string, replayname string, errors []Error, addCSoffset float64) (result []ObjectResult, totalresult []TotalResult, mods uint32, allright bool, CSoffset float64) {

	//prof.ProfStart()
	//defer prof.ProfEnd()

	// 是否加入CS偏差
	UseCSOffset = addCSoffset
	if UseCSOffset > 0 {
		log.Println("加入偏差，偏差为", addCSoffset)
	}

	// 加载replay
	pr := ParseReplay(replayname)
	r := pr.ReplayData

	mods = pr.Mods

	// 根据replay的mods加载map
	b := ParseMapwithMods(mapname, (mods&MOD_HR > 0), (mods&MOD_EZ > 0))
	OD300 := b.OD300
	OD100 := b.OD100
	OD50 := b.OD50
	ODMiss := b.ODMiss
	convert_CS := float2unit(32 * (1 - 0.7 * (b.CircleSize - 5) / 5))

	// 如果replay是HR，改变OD和CS
	if mods&MOD_HR > 0 {
		newOD := math.Min(OD_HR_HENSE * b.OD, OD_MAX)
		OD300 = beatmap.AdjustOD(OD_300_BASE - ( newOD * OD_300_MULT ) + OD_PRECISION_FIX)
		OD100 = beatmap.AdjustOD(OD_100_BASE - ( newOD * OD_100_MULT ) + OD_PRECISION_FIX)
		OD50 = beatmap.AdjustOD(OD_50_BASE - ( newOD * OD_50_MULT ) + OD_PRECISION_FIX)
		ODMiss = beatmap.AdjustOD(OD_MISS_BASE + OD_PRECISION_FIX)
		convert_CS = float2unit(32 * (1 - 0.7 * (math.Min(CS_HR_HENSE * b.CircleSize, CS_MAX) - 5) / 5))
	}

	// 如果replay是EZ，改变OD和CS
	if mods&MOD_EZ > 0 {
		newOD := b.OD * OD_EZ_HENSE
		OD300 = beatmap.AdjustOD(OD_300_BASE - ( newOD * OD_300_MULT ) + OD_PRECISION_FIX)
		OD100 = beatmap.AdjustOD(OD_100_BASE - ( newOD * OD_100_MULT ) + OD_PRECISION_FIX)
		OD50 = beatmap.AdjustOD(OD_50_BASE - ( newOD * OD_50_MULT ) + OD_PRECISION_FIX)
		ODMiss = beatmap.AdjustOD(OD_MISS_BASE + OD_PRECISION_FIX)
		convert_CS = float2unit(32 * (1 - 0.7 * (math.Min(b.CircleSize * CS_EZ_HENSE, 10) - 5) / 5))
	}

	// 计数
	count300 := 0
	count100 := 0
	count50 := 0
	countMiss := 0

	// 判定数组
	totalhits := []int64{}
	// maxcombo
	maxcombo := 0
	nowcombo := 0

	// 击打误差数组
	hiterrors := []int64{}

	// 全局键位占用数组
	keysoccupied := []bool{false, false, false, false}

	// CS判定偏差数组
	CSerrors := []float64{}

	// 依次处理HitObject
	keyindex := 3
	time := r[0].Time + r[1].Time + r[2].Time
	for k := 0; k < len(b.HitObjects); k++ {
	//for k := 0; k < 1; k++ {
	//	log.Println("Object", k+1)
		obj :=  b.HitObjects[k]
		if obj != nil {
			// 滑条
			if o, ok := obj.(*objects.Slider); ok {
				//log.Println("Slider info", o.GetBasicData().StartTime, o.GetBasicData().StartPos, o.GetBasicData().EndTime, o.GetBasicData().EndTime - o.TailJudgeOffset, o.TailJudgeOffset, o.TailJudgePoint, o.ScorePoints)
				// 统计滑条的hit数，是否断连
				requirehits := 0
				realhits := 0
				isBreak := false
				// 判断滑条头
				requirehits += 1
				// ticks的判定倍数
				CS_scale := CIRCLE_JUDGE_SCALE
				// 寻找最近的Key
				//log.Println("Slider head find", r[keyindex].Time, time, o.GetBasicData().StartTime, o.GetBasicData().StartPos)
				// 滑条头占有键位，但这个键位依然对这个滑条的ticks有效
				isfind, nearestindex, lasttime, newkeysoccupied, key := findNearestKey(b.Pauses, keyindex, time, r, o.GetBasicData().StartTime, o.GetBasicData().StartPos, ODMiss, OD50, convert_CS * CS_scale, true, float64(o.GetBasicData().StartTime), keysoccupied)
				//log.Println("Slider hit on", key, r[nearestindex].Time, lasttime + r[nearestindex].Time)
				// 后续还可能存在其他的只对滑条ticks有效的键位
				keys := []Key{key}
				copy(keysoccupied, newkeysoccupied)

				// 滑条头前tick判定数组
				var earlytick int
				var earlytickjudge []bool
				var tail bool
				var tailjudge bool
				// 检查从滑条开始到击中之前的sliderball情况，调整CS
				CS_scale, earlytick, earlytickjudge, tail, tailjudge = checkSliderBallBeforeHit(nearestindex, lasttime, r, o, convert_CS, CS_scale)

				if isfind {
					// 如果找到，判断hit结果，设置下一个index+1
					keyhitresult, hiterror := judgeHitResult(nearestindex, lasttime, r, o.GetBasicData().StartTime, ODMiss, OD300, OD100, OD50)
					if keyhitresult != HitMiss {
						hiterrors = append(hiterrors, hiterror)
					}
					switch keyhitresult {
					case Hit300:
						//log.Println("Slider head 300", o.GetBasicData().StartTime, hiterror, ODMiss, OD300, OD100, OD50)
						if hiterror >= 0 {
							CS_scale = checkInSliderBall(o, o.GetBasicData().StartTime+hiterror, bmath.Vector2d{float64(r[nearestindex].MosueX), float64(r[nearestindex].MouseY)}, convert_CS*CS_scale, true)
						}else {
							CS_scale = CIRCLE_JUDGE_SCALE
						}
						realhits += 1
						nowcombo += 1
						break
					case Hit100:
						//log.Println("Slider head 100", o.GetBasicData().StartTime, hiterror, ODMiss, OD300, OD100, OD50)
						if hiterror >= 0 {
							CS_scale = checkInSliderBall(o, o.GetBasicData().StartTime + hiterror, bmath.Vector2d{float64(r[nearestindex].MosueX), float64(r[nearestindex].MouseY)}, convert_CS * CS_scale, true)
						}else {
							CS_scale = CIRCLE_JUDGE_SCALE
						}
						realhits += 1
						nowcombo += 1
						break
					case Hit50:
						//log.Println("Slider head 50", o.GetBasicData().StartTime, hiterror, ODMiss, OD300, OD100, OD50)
						if hiterror >= 0 {
							CS_scale = checkInSliderBall(o, o.GetBasicData().StartTime + hiterror, bmath.Vector2d{float64(r[nearestindex].MosueX), float64(r[nearestindex].MouseY)}, convert_CS * CS_scale, true)
						}else {
							CS_scale = CIRCLE_JUDGE_SCALE
						}
						realhits += 1
						nowcombo += 1
						break
					case HitMiss:
						//log.Println("Slider head miss", o.GetBasicData().StartTime, hiterror, ODMiss, OD300, OD100, OD50)
						CS_scale = CIRCLE_JUDGE_SCALE
						isBreak = true
						nowcombo = 0
						break
					}
					keyindex = nearestindex + 1
					time = lasttime + r[nearestindex].Time
					//log.Println("hit in", time)
				}else {
					// 如果没找到，输出miss，设置下一个index
					//log.Println("Slider head no found", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "Miss", r[keyindex].Time, lasttime)
					CS_scale = CIRCLE_JUDGE_SCALE
					isBreak = true
					nowcombo = 0
					keyindex = nearestindex
					time = lasttime
				}
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				// 判断ticks
				// 预先设置滑条尾判断，以防滑条尾判断早于tick
				istailjudge := true
				istailhit := false
				for i, t := range o.ScorePoints {
				//for _, t := range o.ScorePoints {
					requirehits += 1
					var isHit bool
					if i >= earlytick {
						//log.Println("Normal Check Tick hit", i+1, time, CS_scale * convert_CS)
						newisHit, nextindex, nexttime, newkeysoccupied, newCSscale, newistailjudge, newistailhit, newkeys, newCSerrors := isTickHit(o, keyindex, time, r, t.Time, t.Pos, convert_CS, CS_scale, keysoccupied, keys, o.GetBasicData().EndTime-o.TailJudgeOffset, o.TailJudgePoint, istailjudge, istailhit, CSerrors)
						isHit = newisHit
						keys = newkeys
						istailjudge = newistailjudge
						istailhit = newistailhit
						CS_scale = newCSscale
						copy(keysoccupied, newkeysoccupied)
						copy(CSerrors, newCSerrors)
						keyindex = nextindex
						time = nexttime
					}else {
						//log.Println("Early Check Tick hit", i+1, time)
						isHit = earlytickjudge[i]
					}
					if isHit {
						//log.Println("Tick", i+1, "hit", t.Time, t.Pos)
						CS_scale = TICK_JUDGE_SCALE
						realhits += 1
						nowcombo += 1
					}else {
						//log.Println("Tick", i+1, "not hit", t.Time, t.Pos)
						CS_scale = CIRCLE_JUDGE_SCALE
						isBreak = true
						nowcombo = 0
					}
					maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				}
				// 判断滑条尾
				requirehits += 1
				if !istailjudge {
					// 预先判断的情况
					//log.Println("Quick Check Slider Tail", istailhit)
					if istailhit {
						//log.Println("Quick Check Slider tail hit", o.GetBasicData().EndTime-o.TailJudgeOffset, o.TailJudgePoint)
						realhits += 1
						nowcombo += 1
						// 寻找状态改变后的时间点
						//log.Println("Start find slider release", keyindex, time+r[keyindex].Time)
						keyindex, time, keysoccupied = findRelease(keyindex, time+r[keyindex].Time, r, keysoccupied, true)
						time -= r[keyindex].Time
					}else {
						//log.Println("Quick Check Slider tail not hit", o.GetBasicData().EndTime-o.TailJudgeOffset, o.TailJudgePoint)
						//log.Println("Start find slider release", keyindex, time+r[keyindex].Time)
						keyindex, time, keysoccupied = findRelease(keyindex, time+r[keyindex].Time, r, keysoccupied, true)
						time -= r[keyindex].Time
					}
				}else {
					//log.Println("Slider tail judge", r[keyindex], time, o.GetBasicData().EndTime-o.TailJudgeOffset, o.TailJudgeOffset, o.TailJudgePoint, convert_CS, CS_scale*convert_CS)
					var isHit bool
					var nextindex int
					var nexttime int64

					newisHit, newnextindex, newnexttime, newkeysoccupied, newCSscale, _, _, _, newCSerrors := isTickHit(o, keyindex, time, r, o.GetBasicData().EndTime-o.TailJudgeOffset, o.TailJudgePoint, convert_CS, CS_scale, keysoccupied, keys, 0, bmath.Vector2d{0, 0}, false, false, CSerrors)
					// 滑条尾在滑条头前判断
					if tail {
						//log.Println("Slider head judged later than tail another!!!")
						isHit = tailjudge
					}else {
						isHit = newisHit
					}
					nextindex = newnextindex
					nexttime = newnexttime
					CS_scale = newCSscale
					//log.Println("Slider tail judge result", r[nextindex].Time, nexttime+r[nextindex].Time, nearestindex, nextindex)
					// 滑条头如果判定得比滑条尾还晚！！！
					if nearestindex > nextindex {
						//log.Println("Slider head judged later than tail!!!", nexttime+r[nextindex].Time, lasttime+r[nearestindex].Time)
						nextindex = nearestindex
						nexttime = lasttime
					}
					copy(keysoccupied, newkeysoccupied)
					CSerrors = newCSerrors
					if isHit {
						//log.Println("Slider tail hit", o.GetBasicData().EndTime-o.TailJudgeOffset, o.TailJudgePoint)
						realhits += 1
						nowcombo += 1
						// 寻找状态改变后的时间点
						//log.Println("Start find slider release", r[nextindex].Time, nexttime+ r[nextindex].Time)
						keyindex, time, keysoccupied = findRelease(nextindex, nexttime+r[nextindex].Time, r, keysoccupied, true)
						time -= r[keyindex].Time
					} else {
						//log.Println("Slider tail not hit", o.GetBasicData().EndTime-o.TailJudgeOffset, o.TailJudgePoint)
						//log.Println("Start find slider release", r[nextindex].Time, nexttime+ r[nextindex].Time)
						keyindex, time, keysoccupied = findRelease(nextindex, nexttime+r[nextindex].Time, r, keysoccupied, true)
						time -= r[keyindex].Time
					}
				}
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				// 滑条总体情况
				sliderhitresult := judgeSlider(requirehits, realhits)
				switch sliderhitresult {
				case Hit300:
					//log.Println("Slider count as 300", requirehits, realhits, "Object", k+1)
					count300 += 1
					totalhits = append(totalhits, 300)
					realhits += 1
					break
				case Hit100:
					log.Println("Slider count as 100", requirehits, realhits, "Object", k+1)
					count100 += 1
					totalhits = append(totalhits, 100)
					realhits += 1
					break
				case Hit50:
					log.Println("Slider count as 50", requirehits, realhits, "Object", k+1)
					count50 += 1
					totalhits = append(totalhits, 50)
					realhits += 1
					break
				case HitMiss:
					log.Println("Slider count as Miss", requirehits, realhits, "Object", k+1)
					countMiss += 1
					totalhits = append(totalhits, 0)
					isBreak = true
					break
				}
				if isBreak {
					log.Println("Slider breaks")
				}else {
					//log.Println("Slider no breaks")
				}
				result = append(result, ObjectResult{o.GetBasicData().StartPos, o.GetBasicData().JudgeTime, sliderhitresult, isBreak})
			}
			// note
			if o, ok := obj.(*objects.Circle); ok {
				// 寻找最近的Key
				keyhitresult := HitMiss
				isBreak := true
				// 占用key对note无用
				isfind, nearestindex, lasttime, newkeysoccupied, _ := findNearestKey(b.Pauses, keyindex, time, r, o.GetBasicData().StartTime, o.GetBasicData().StartPos, ODMiss, OD50, convert_CS, false, 0, keysoccupied)
				copy(keysoccupied, newkeysoccupied)
				if isfind {
					// 如果找到，判断hit结果，设置下一个index+1
					var hiterror int64
					keyhitresult, hiterror = judgeHitResult(nearestindex, lasttime, r, o.GetBasicData().StartTime, ODMiss, OD300, OD100, OD50)
					if keyhitresult != HitMiss {
						hiterrors = append(hiterrors, hiterror)
					}
					switch keyhitresult {
					case Hit300:
						//log.Println("Circle count as 300", "Object", k+1)
						count300 += 1
						nowcombo += 1
						totalhits = append(totalhits, 300)
						break
					case Hit100:
						log.Println("Circle count as 100", "Object", k+1)
						count100 += 1
						nowcombo += 1
						totalhits = append(totalhits, 100)
						break
					case Hit50:
						log.Println("Circle count as 50", "Object", k+1)
						count50 += 1
						nowcombo += 1
						totalhits = append(totalhits, 50)
						break
					case HitMiss:
						log.Println("Circle count as Miss", "Object", k+1)
						countMiss += 1
						nowcombo = 0
						totalhits = append(totalhits, 0)
						break
					}
					time = lasttime + r[nearestindex].Time
					//log.Println("hit in", time)
					// 寻找状态改变后的时间点
					keyindex, time, keysoccupied = findRelease(nearestindex, time, r, keysoccupied, true)
					time -= r[keyindex].Time
				}else {
					// 如果没找到，输出miss，设置下一个index
					log.Println("Circle count as Late Miss, Object", k+1)
					countMiss += 1
					nowcombo = 0
					totalhits = append(totalhits, 0)
					keyindex = nearestindex
					time = lasttime
				}
				if keyhitresult != HitMiss {
					isBreak = false
				}
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				result = append(result, ObjectResult{o.GetBasicData().StartPos, o.GetBasicData().JudgeTime, keyhitresult, isBreak})
			}
			// 转盘
			if o, ok := obj.(*objects.Spinner); ok {
				//log.Println("Spinner! skip!", k+1, o.GetBasicData())
				count300 += 1
				nowcombo += 1
				totalhits = append(totalhits, 300)
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				result = append(result, ObjectResult{o.GetBasicData().StartPos, o.GetBasicData().JudgeTime, Hit300, false})
			}
		}
		// 判定修正
		err := shouldfixError(k+1, errors)
		if err != nil {
			// 进行修正
			result, count300, count100, count50, countMiss, maxcombo, nowcombo, totalhits = fixError(*err, result, count300, count100, count50, countMiss, maxcombo, nowcombo, totalhits)
		}
		ur := calculateUnstableRate(hiterrors)
		if math.IsNaN(ur) {
			ur = 0
		}
		tmptotalresult := TotalResult{	uint16(count300),
										uint16(count100),
										uint16(count50),
										uint16(countMiss),
										uint16(maxcombo),
										mods,
										score.CalculateAccuracy(totalhits),
										score.CalculateRank(totalhits, mods),
										oppai.PPv2{},
										ur}
		//tmptotalresult.PP = calculatePP(mapname, tmptotalresult)
		tmptotalresult.PP = calculatePPbyNum(mapname, tmptotalresult, k+1)
		totalresult = append(totalresult, tmptotalresult)
		//log.Println("result", tmptotalresult)
		//log.Println("Now Max Combo:", maxcombo)
		//log.Println("Acc:", score.CalculateAccuracy(totalhits))
		//log.Println("Unstable Rate:", calculateUnstableRate(hiterrors))
	}

	log.Println("Count 300:", count300)
	log.Println("Count 100:", count100)
	log.Println("Count 50:", count50)
	log.Println("Count Miss:", countMiss)
	log.Println("Max Combo:", maxcombo)
	log.Println("Acc:", totalresult[len(totalresult)-1].Acc)
	log.Println("PP:", totalresult[len(totalresult)-1].PP.Total)
	log.Println("UR:", totalresult[len(totalresult)-1].UR)

	// CS修正
	sort.Float64s(CSerrors)
	CSoffset = NO_USE_CS_OFFSET
	// 只在只有滑条尾判断出错时修正
	//if (int(pr.Count100)-count100) == (maxcombo-int(pr.MaxCombo)) {
	//	CSoffset = CSerrors[int(pr.Count100)-count100-1]
	//}

	if settings.VSplayer.ReplayandCache.ReplayDebug {
		// 分析情况和replay记录情况对比检查
		allright = true
		// 检查各个判定个数
		if !checkHits(count300, count100, count50, countMiss, pr.Count300, pr.Count100, pr.Count50, pr.CountMiss) {
			allright = false
			log.Println("判定存在误差！")
			log.Println("300 true:", count300, "replay:", pr.Count300, "error:", count300-int(pr.Count300))
			log.Println("100 true:", count100, "replay:", pr.Count100, "error:", count100-int(pr.Count100))
			log.Println("50 true:", count50, "replay:", pr.Count50, "error:", count50-int(pr.Count50))
			log.Println("Miss true:", countMiss, "replay:", pr.CountMiss, "error:", countMiss-int(pr.CountMiss))
		}
		// 检查最大连击数
		if maxcombo != int(pr.MaxCombo) {
			allright = false
			log.Println("最大连击数存在误差！")
			log.Println("Max combo true:", maxcombo, "replay:", pr.MaxCombo, "error:", maxcombo-int(pr.MaxCombo))
		}
		// 总体情况
		if allright {
			log.Println("检查结果完全一致！")
		}
	}

	return result, totalresult, mods, allright, CSoffset
}

// 定位Key放下的位置
func findRelease(keyindex int, starttime int64, r []*rplpa.ReplayData, keysoccupied []bool, sametimecheck bool) (int, int64, []bool) {
	keypress := r[keyindex].KeyPressed
	index := keyindex
	time := starttime

	// 特殊情况：第一个片段就按下两个键，其中一个键对note生效，另一个note依然留存，即状态已经改变？
	if sametimecheck {
		isanykeyoccupied := keysoccupied[0] || keysoccupied[1] || keysoccupied[2] || keysoccupied[3]
		if isanykeyoccupied {
			keypressstatus := []bool{
				getSameTimePressKeyStatus(keypress.Key1, keysoccupied[0]),
				getSameTimePressKeyStatus(keypress.Key2, keysoccupied[1]),
				getSameTimePressKeyStatus(trueClick(keypress.LeftClick, keypress.Key1), keysoccupied[2]),
				getSameTimePressKeyStatus(trueClick(keypress.RightClick, keypress.Key2), keysoccupied[3]),
			}
			for _, status := range keypressstatus {
				if status {
					//log.Println("Same time press", time, keypress, keysoccupied, keypressstatus)
					return index, time, keysoccupied
				}
			}
		}
	}

	for {
		index++
		// 如果到了replay尽头还没有按下按键，直接返回原数组
		if r[index].Time == REPLAY_END_TIME {
			//log.Println("Reach Replay End, No more!", time)
			return keyindex, starttime, keysoccupied
		}
		time += r[index].Time
		// 如果按键状态改变，则返回
		//log.Println("Key compare", time - r[index].Time, *keypress, time, *r[index].KeyPressed, isPressChanged(*keypress, *r[index].KeyPressed))
		//if time > 29400 {
		//	os.Exit(2)
		//}
		if isPressChanged(*keypress, *r[index].KeyPressed) {
			// 重新设置占用键位情况
			newkeysoccupied, _ := resetKeysOccupied(*keypress, *r[index].KeyPressed, keysoccupied,false, []Key{})
			//log.Println("Reset key occupied release", *keypress, *r[index].KeyPressed, keysoccupied, newkeysoccupied)
			//log.Println("Find release before", r[index].Time, time, sametimecheck)
			// 返回之前已判定的无效键
			return index, time, newkeysoccupied
		}
		keypress = r[index].KeyPressed
	}
}

// 确定是否出现按下状态的改变
func isPressChanged(p1 rplpa.KeyPressed, p2 rplpa.KeyPressed) bool {
	if p1!=p2 {
		// 如果不相等
		if p2==None{
			// 如果没有按键，则肯定状态改变
			return true
		}else {
			// 否则，如果p2按下了某个键，p1必须也按下了这个键，否则状态改变
			if p2.Key1{
				if !p1.Key1{
					return true
				}
			}
			if p2.Key2{
				if !p1.Key2{
					return true
				}
			}
			if trueClick(p2.LeftClick, p2.Key1){
				if !trueClick(p1.LeftClick, p1.Key1){
					return true
				}
			}
			if trueClick(p2.RightClick, p2.Key2){
				if !trueClick(p1.RightClick, p1.Key2){
					return true
				}
			}
			return false
		}
	}else {
		//相等，无改变
		return false
	}
}

// 根据前后两个按键状况重新设置键位占用情况
func resetKeysOccupied(p1 rplpa.KeyPressed, p2 rplpa.KeyPressed, keysoccupied []bool, presscheck bool, keys []Key) ([]bool, []Key) {
	newkeysoccupied := make([]bool, 4)
	copy(newkeysoccupied, keysoccupied)
	p1M1 := trueClick(p1.LeftClick, p1.Key1)
	p2M1 := trueClick(p2.LeftClick, p2.Key1)
	p1M2 := trueClick(p1.RightClick, p1.Key2)
	p2M2 := trueClick(p2.RightClick, p2.Key2)
	for i, keyoccupied := range keysoccupied {
		// 遍历按键，如果被占用
		// 1.前一个按键也按着，后一个按键没按着，则被占位被释放
		// 2.或者前一个按键没按着，说明已经被释放了？
		if keyoccupied {
			switch i {
			case 0:
				if !p1.Key1 || (p1.Key1 && !p2.Key1) {
					newkeysoccupied[0] = false
					//log.Println("K1 no occupied", p1.Key1, p2.Key1, p1.Key1 && !p2.Key1, keysoccupied, newkeysoccupied)
				}
				break
			case 1:
				if !p1.Key2 || (p1.Key2 && !p2.Key2) {
					newkeysoccupied[1] = false
					//log.Println("K2 no occupied", p1.Key2, p2.Key2, p1.Key2 && !p2.Key2, keysoccupied, newkeysoccupied)
				}
				break
			case 2:
				if !p1M1 || (p1M1 && !p2M1) {
					newkeysoccupied[2] = false
					//log.Println("M1 no occupied", keysoccupied, newkeysoccupied)
				}
				break
			case 3:
				if !p1M2 || (p1M2 && !p2M2) {
					newkeysoccupied[3] = false
					//log.Println("M2 no occupied", keysoccupied, newkeysoccupied)
				}
				break
			}
		}
	}
	if presscheck {
		for i, keyoccupied := range keysoccupied {
			// 遍历按键，如果未被占用
			// 1.前一个按键未按着，后一个按键按着，则被占用
			if !keyoccupied {
				switch i {
				case 0:
					if !p1.Key1 && p2.Key1 {
						newkeysoccupied[0] = true
						keys = append(keys, Key1)
						//log.Println("K1 occupied", p1.Key1, p2.Key1, p1.Key1 && !p2.Key1, keysoccupied, newkeysoccupied, keys)
					}
					break
				case 1:
					if !p1.Key2 && p2.Key2 {
						newkeysoccupied[1] = true
						keys = append(keys, Key2)
						//log.Println("K2 occupied", p1.Key2, p2.Key2, p1.Key2 && !p2.Key2, keysoccupied, newkeysoccupied)
					}
					break
				case 2:
					if !p1M1 && p2M1 {
						newkeysoccupied[2] = true
						keys = append(keys, Mouse1)
						//log.Println("M1 occupied", keysoccupied, newkeysoccupied)
					}
					break
				case 3:
					if !p1M2 && p2M2 {
						newkeysoccupied[3] = true
						keys = append(keys, Mouse2)
						//log.Println("M2 occupied", keysoccupied, newkeysoccupied)
					}
					break
				}
			}
		}
	}
	return newkeysoccupied, keys
}

// 寻找最近的击中的Key
func findNearestKey(pauses []objects.BaseObject, start int, starttime int64, r []*rplpa.ReplayData, requirehittime int64, requirepos bmath.Vector2d, ODMiss float64, OD50 float64, CS float64, isNextTick bool, headtime float64, keysoccupied []bool) (bool, int, int64, []bool, Key) {
	index := start
	time := starttime
	for {
		hit := r[index]
		//newkeysoccupied := checkBeforePress(hit, keysoccupied)
		//ispress, nnewkeysoccupied, _ := isPressed(hit, newkeysoccupied)
		//tmpisincircle, _ := isInCircle(hit, requirepos, CS)
		//log.Println("Find move", hit.Time + time, requirehittime, tmpisincircle, ispress, newkeysoccupied, nnewkeysoccupied, bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos), ODMiss, OD50, CS + 0.01, keysoccupied)

		// 如果已经搜索到replay结尾，返回搜索的开头？
		if hit.Time == REPLAY_END_TIME {
			//log.Println("Search reach replay end!", start, starttime)
			return false, start, starttime, keysoccupied, NoKey
		}

		// 如果时间已经超过最后时间，直接返回
		realhittime := hit.Time + time
		if float64(realhittime) > float64(requirehittime) + OD50 {
			//log.Println("Find move already too late", realhittime, float64(requirehittime) + OD50)
			// 占用状态不变，直接返回原占用数组
			if isNextTick {
				// （tick、滑条尾）返回上一个生效点
				index, time = findFirstAfterHead(headtime, r)
				time -= r[index].Time
				//log.Println("Return to head", r[index].Time, time)
			}
			// 此键位未被占用
			return false, index, time, keysoccupied, NoKey
		}
		// 判断是否在圈内
		isincircle, _ := isInCircle(hit, requirepos, CS)
		if isincircle {
			// 如果在圈内，且按下按键
			newkeysoccupied := checkBeforePress(hit, keysoccupied)
			copy(keysoccupied, newkeysoccupied)
			ispressed, newkeysoccupied, key := isPressed(hit, keysoccupied)
			if ispressed {
				realhittime := hit.Time + time
				// 判断这个时间点和object时间点的关系
				//log.Println("Judge", realhittime, requirehittime, ODMiss)
				if isHitOver(realhittime, requirehittime, ODMiss) {
					// 如果已经超过这个object的最后hit时间，则未找到最接近的Key，直接返回这个时间点
					// 这种情况不占用键位
					//log.Println("isHitOver")
					return false, index, time, keysoccupied, NoKey
				}else if isHitMiss(realhittime, requirehittime, ODMiss){
					// 如果落在这个object的区域内，则找到Key，返回这个Key的时间点
					// 这种情况，无论30010050miss都会占用键位
					// 在休息段提前按下
					if isNotInPauses(realhittime, requirehittime, pauses, OD50) {
						return true, index, time, newkeysoccupied, key
					}else {
						// 跳过返回，但需要复制键位占用情况
						//log.Println("Pauses Skip", realhittime)
						copy(keysoccupied, newkeysoccupied)
					}
				}
			}
		}else {
			newkeysoccupied := checkBeforePress(hit, keysoccupied)
			copy(keysoccupied, newkeysoccupied)
			ispressed, newkeysoccupied, key := isPressed(hit, keysoccupied)
			// 如果不在圈内，且按下按键
			if ispressed {
				realhittime := hit.Time + time
				// 判断这个时间点和object时间点的关系
				if float64(realhittime) > float64(requirehittime) + OD50 {
					// 如果在最后时间之后按下，没效果，等于没找到
					// 最后时间为最后能按出50的时间
					//log.Println("Hit too late", realhittime, requirehittime)
					if isNextTick {
						// （tick、滑条尾）返回上一个生效点
						index, time = findFirstAfterHead(headtime, r)
						time -= r[index].Time
						//log.Println("Return to last head", r[index].Time, time)
					}
					// 此键位未被占用
					return false, index, time, newkeysoccupied, NoKey
				}else {
					// 如果最后时间前按下，没效果，此键位失去对下一个非tick的object（note、滑条头）的效果，寻找下一个按键放下的地方
					//log.Println("Tap out is no use!", index, realhittime)
					index, time, keysoccupied = findRelease(index, realhittime, r, newkeysoccupied, false)
					time -= r[index].Time
					// （tick、滑条尾）如果这个时间大于最后时间，则用最后时间重新定位tick生效位置
					if float64(time) > float64(requirehittime) + OD50 {
						if isNextTick {
							index, time = findFirstAfterHead(headtime, r)
							time -= r[index].Time
							//log.Println("Return to last head", r[index].Time, time)
						}
						// 此键位被占用
						return false, index, time, keysoccupied, key
					}
					continue
				}
			}
		}
		index++
		time += hit.Time
	}
}

// 按下前查看松开的键位
func checkBeforePress(hit *rplpa.ReplayData, keysoccupied []bool) ([]bool) {
	press := hit.KeyPressed
	newkeysoccupied := make([]bool, 4)
	copy(newkeysoccupied, keysoccupied)
	for i, keyoccupied := range keysoccupied {
		if keyoccupied {
			// 首先遍历按键，如果占用且被放开，修改占用数组
			switch i {
			case 0:
				if !press.Key1 {
					newkeysoccupied[0] = false
				}
				break
			case 1:
				if !press.Key2 {
					newkeysoccupied[1] = false
				}
				break
			case 2:
				if !trueClick(press.LeftClick, press.Key1) {
					newkeysoccupied[2] = false
				}
				break
			case 3:
				if !trueClick(press.RightClick, press.Key2) {
					newkeysoccupied[3] = false
				}
				break
			}
		}
	}
	return newkeysoccupied
}

// 该时间点是否按下按键
func isPressed(hit *rplpa.ReplayData, keysoccupied []bool) (bool, []bool, Key) {
	press := hit.KeyPressed
	newkeysoccupied := make([]bool, 4)
	copy(newkeysoccupied, keysoccupied)
	for i, keyoccupied := range keysoccupied {
		if !keyoccupied {
			// 遍历按键，如果不占用且按下，立即返回，并修改占用数组
			switch i {
			case 0:
				if press.Key1 {
					newkeysoccupied[0] = true
					//log.Println("K1 occupied", keysoccupied, newkeysoccupied)
					return true, newkeysoccupied, Key1
				}
				break
			case 1:
				if press.Key2 {
					newkeysoccupied[1] = true
					///log.Println("K2 occupied", keysoccupied, newkeysoccupied)
					return true, newkeysoccupied, Key2
				}
				break
			case 2:
				if trueClick(press.LeftClick, press.Key1) {
					newkeysoccupied[2] = true
					//log.Println("M1 occupied", keysoccupied, newkeysoccupied)
					return true, newkeysoccupied, Mouse1
				}
				break
			case 3:
				if trueClick(press.RightClick, press.Key2) {
					newkeysoccupied[3] = true
					//log.Println("M2 occupied", keysoccupied, newkeysoccupied)
					return true, newkeysoccupied, Mouse2
				}
				break
			}
		}
	}
	// 无按下，返回原占用数组
	//log.Println("No key occupied", keysoccupied, newkeysoccupied)
	return false, newkeysoccupied, NoKey
}

// 该时间点是否按下按键（ticks）
func isPressedTick(hit *rplpa.ReplayData, keysoccupied []bool, keys []Key) bool {
	press := hit.KeyPressed
	tickkeysoccupied := getTickKeysOccupied(keysoccupied, keys)
	for i, keyoccupied := range tickkeysoccupied {
		if !keyoccupied {
			// 遍历按键，如果不占用且按下，立即返回，并修改占用数组
			switch i {
			case 0:
				if press.Key1 {
					//log.Println("Tick K1 pressed")
					return true
				}
				break
			case 1:
				if press.Key2 {
					//log.Println("Tick K2 pressed")
					return true
				}
				break
			case 2:
				if trueClick(press.LeftClick, press.Key1) {
					//log.Println("Tick M1 pressed")
					return true
				}
				break
			case 3:
				if trueClick(press.RightClick, press.Key2) {
					//log.Println("Tick M2 pressed")
					return true
				}
				break
			}
		}
	}
	// 无按下
	//log.Println("Tick no key pressed")
	return false
}

// 该时间点是否按下按键（非ticks）
// 不是任何object：判断是否按下，只需要判断是否按下按键
func isPressedNoTick(hit *rplpa.ReplayData) bool {
	return hit.KeyPressed.LeftClick || hit.KeyPressed.RightClick || hit.KeyPressed.Key1 || hit.KeyPressed.Key2
}

// 获取ticks的键位占有数组
func getTickKeysOccupied(keysoccupied []bool, keys []Key) []bool {
	//log.Println("Tick KeysOccupied", keysoccupied, key)
	newkeysoccupied := make([]bool, 4)
	copy(newkeysoccupied, keysoccupied)
	for _, key := range keys {
		for i, keyoccupied := range keysoccupied {
			if keyoccupied {
				// 遍历按键，如果占用且符合滑条头的占用键位，去除
				switch i {
				case 0:
					if key == Key1 {
						newkeysoccupied[0] = false
						//log.Println("K1 occupied delete", keysoccupied, newkeysoccupied)
					}
					break
				case 1:
					if key == Key2 {
						newkeysoccupied[1] = false
						//log.Println("K2 occupied delete", keysoccupied, newkeysoccupied)
					}
					break
				case 2:
					if key == Mouse1 {
						newkeysoccupied[2] = false
						//log.Println("M1 occupied delete", keysoccupied, newkeysoccupied)
					}
					break
				case 3:
					if key == Mouse2 {
						newkeysoccupied[3] = false
						//log.Println("M2 occupied delete", keysoccupied, newkeysoccupied)
					}
					break
				}
			}
		}
	}
	return newkeysoccupied
}

func isInCircle(hit *rplpa.ReplayData, requirepos bmath.Vector2d, CS float64) (judge bool, error float64) {
	adjustCS := CS + 0.01
	dist := bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos)
	if UseCSOffset > 0 {
		error = adjustCS - UseCSOffset - dist
	} else {
		error = adjustCS - dist
	}
	judge = (error >= 0)
	return judge, error
}

// 是否超过object的最后时间点
func isHitOver(realhittime int64, requirehittime int64, ODMiss float64) bool {
	return float64(realhittime) > float64(requirehittime) + ODMiss
}

// 判断hit结果
func judgeHitResult(index int, lasttime int64, r []*rplpa.ReplayData, requirehittime int64, ODMiss float64, OD300 float64, OD100 float64, OD50 float64) (HitResult, int64){
	realhittime := r[index].Time + lasttime
	//log.Println("Judge Hit", realhittime, requirehittime, OD300, OD100, OD50, ODMiss)
	if isHit300(realhittime, requirehittime, OD300) {
		return Hit300, realhittime - requirehittime
	}else if isHit100(realhittime, requirehittime, OD100) {
		return Hit100, realhittime - requirehittime
	}else if isHit50(realhittime, requirehittime, OD50) {
		return Hit50, realhittime - requirehittime
	}else if isHitMiss(realhittime, requirehittime, ODMiss) {
		return HitMiss, realhittime - requirehittime
	}else {
		return HitMiss, realhittime - requirehittime
	}
}

func isHitMiss(realhittime int64, requirehittime int64, ODMiss float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - ODMiss) && (float64(realhittime) <= float64(requirehittime) + ODMiss)
}

func isHit50(realhittime int64, requirehittime int64, OD50 float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - OD50) && (float64(realhittime) <= float64(requirehittime) + OD50)
}

func isHit100(realhittime int64, requirehittime int64, OD100 float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - OD100) && (float64(realhittime) <= float64(requirehittime) + OD100)
}

func isHit300(realhittime int64, requirehittime int64, OD300 float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - OD300) && (float64(realhittime) <= float64(requirehittime) + OD300)
}

// 判断tick是否被击中并按下
func isTickHit(slider *objects.Slider, start int, starttime int64, r []*rplpa.ReplayData, requirehittime int64, requirepos bmath.Vector2d, CS float64, initialCSscale float64, keysoccupied []bool, keys []Key, tailrequirehittime int64, tailrequirepos bmath.Vector2d, judgetail bool, hittail bool, CSerrors []float64) (bool, int, int64, []bool, float64, bool, bool, []Key, []float64) {
	index := start
	time := starttime
	// 初始的CS范围
	CSscale := initialCSscale
	istailjugde := judgetail
	istailhit := hittail
	newCSerrors := make([]float64, len(CSerrors))
	copy(newCSerrors, CSerrors)
	for {
		//寻找正好的一点或者区间
		hit := r[index]
		realhittime := hit.Time + time
		//log.Println("Tick Judge", hit, realhittime, starttime)

		// 【猜测】滑条曲线顶点过多的情况下，会在滑条开头造成几毫秒的卡顿
		//if slider.GetCurvePointsLen() >= 100 && realhittime - slider.GetBasicData().StartTime < 10 {
		//	log.Println("Skip Tick Judge check", slider.GetCurvePointsLen(), realhittime - slider.GetBasicData().StartTime)
		//}else {
		//
		//}

		// 判断前，每次都要检查键位占用情况
		// 为什么要这样做，见 https://github.com/ppy/osu/blob/f134b9c886edcd42eb1aa8a6e232789a017a61aa/osu.Game.Rulesets.Osu/Objects/Drawables/Pieces/SliderBall.cs#L153
		newkeysoccupied, newkeys := resetKeysOccupied(*r[index-1].KeyPressed, *r[index].KeyPressed, keysoccupied, true, keys)
		keys = newkeys
		//log.Println("Reset key occupied tick", *r[index-1].KeyPressed, *r[index].KeyPressed, keysoccupied, newkeysoccupied, keys, newkeys)
		copy(keysoccupied, newkeysoccupied)

		// 每次判断前，还要检查这一时刻的replay光标在不在sliderball中，调整CS范围
		// 在滑条头真实时间开始后再检查
		if realhittime >= slider.GetBasicData().StartTime {
			//log.Println("Check if in sliderball", realhittime, bmath.Vector2d{float64(hit.MosueX), float64(hit.MouseY)}, CSscale)
			CSscale = checkInSliderBall(slider, realhittime, bmath.Vector2d{float64(hit.MosueX), float64(hit.MouseY)}, CS*CSscale, isPressedTick(hit, keysoccupied, keys))
		}

		// 存在滑条尾早于tick的情况，所以在判断tick时顺便判断滑条
		if judgetail {
			if realhittime == tailrequirehittime {
				// 找到正好的一点
				isincircle, cserror := isInCircle(hit, tailrequirepos, CS*CSscale)
				if cserror >= 0 {
					newCSerrors = append(newCSerrors, cserror)
				}
				//log.Println(" Tail Tick Judge Tap", tailrequirehittime, realhittime, bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), tailrequirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), tailrequirepos), CS*CSscale)
				if isincircle {
					// 在圈内
					// tick判断不会改变占用情况
					ispressed := isPressedTick(hit, keysoccupied, keys)
					//log.Println("Tail Tick Judge Tap Press", ispressed, hit.KeyPressed, keysoccupied)
					if ispressed {
						//按下，则击中成功
						//滑条尾判断结束，后面不再判断
						istailhit = true
					}
				}
				istailjugde = false
			} else if realhittime < tailrequirehittime && realhittime+r[index+1].Time > tailrequirehittime {
				// 找到正好的区间
				//log.Println("Tail Tick Judge Range", tailrequirehittime, realhittime, realhittime+r[index+1].Time, hit, r[index+1])
				// 寻找正确的光标位置
				// 1.起始点
				// 2.结束点
				// 3.开始和结束算出的中间点
				// 暂时使用中间点
				realhit := getTickRangeJudgePoint(tailrequirehittime, r[index], r[index+1], realhittime)
				isincircle, cserror := isInCircle(realhit, tailrequirepos, CS*CSscale)
				if cserror >= 0 {
					newCSerrors = append(newCSerrors, cserror)
				}
				//realhit := r[index]
				//log.Println("Forward Tail Tick Judge Range Find Require Point", realhit.KeyPressed, realhit, tailrequirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(realhit.MosueX), float64(realhit.MouseY)), tailrequirepos), CS*CSscale)
				if isincircle {
					// 在圈内
					// tick判断不会改变占用情况
					// 寻找正确的按下放开位置
					// 1.起始点
					// 2.结束点
					// 3.离哪个近用哪个
					// 暂时使用离哪个近用哪个，最中间点可以算按下或者未按下
					// 暂时最中间算未按下
					presshit := hit
					if tailrequirehittime-realhittime >= realhittime+r[index+1].Time-tailrequirehittime {
						presshit = r[index+1]
					}
					ispressed := isPressedTick(presshit, keysoccupied, keys)
					//log.Println("Tail Tick Judge Range Press", ispressed, presshit.KeyPressed, keysoccupied)
					if ispressed {
						//按下，则击中成功
						istailhit = true
					}
				}
				istailjugde = false
			}
		}

		if realhittime == requirehittime {
			// 找到正好的一点
			isincircle, cserror := isInCircle(hit, requirepos, CS*CSscale)
			if cserror >= 0 {
				newCSerrors = append(newCSerrors, cserror)
			}
			//log.Println("Tick Judge Tap", requirehittime, realhittime, bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos), CS * CSscale)
			// 【猜测】滑条曲线顶点太多造成的判定延迟
			if slider.GetCurvePointsLen() >= 100 {
				// 判断下一个时间点的坐标是否和这次一样，一样就可能是延迟判定
				laghit := r[index+1]
				if bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)) == bmath.NewVec2d(float64(laghit.MosueX), float64(laghit.MouseY)) {
					laghittime := realhittime + laghit.Time
					lagisincircle, _ := isInCircle(laghit, slider.GetPointAt(laghittime), CS*CSscale)
					//log.Println("Tick Judge Tap Lag", laghittime, bmath.NewVec2d(float64(laghit.MosueX), float64(laghit.MouseY)), slider.GetPointAt(laghittime), bmath.Vector2d.Dst(bmath.NewVec2d(float64(laghit.MosueX), float64(laghit.MouseY)), slider.GetPointAt(laghittime)), CS * CSscale)
					if lagisincircle {
						// 在圈内
						// tick判断不会改变占用情况
						lagkeysoccupied, lagkeys := resetKeysOccupied(*r[index].KeyPressed, *r[index+1].KeyPressed, keysoccupied, true, keys)
						lagispressed := isPressedTick(laghit, lagkeysoccupied, lagkeys)
						//log.Println("Tick Judge Tap Lag Press", lagispressed, laghit.KeyPressed, lagkeysoccupied)
						if lagispressed {
							//按下，则击中成功
							return true, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
						}
					}
				}
			}
			if isincircle {
				// 在圈内
				// tick判断不会改变占用情况
				ispressed := isPressedTick(hit, keysoccupied, keys)
				//log.Println("Tick Judge Tap Press", ispressed, hit.KeyPressed, keysoccupied)
				if ispressed {
					//按下，则击中成功
					return true, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
				}
			}
			return false, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
		}else if realhittime < requirehittime && realhittime + r[index+1].Time > requirehittime{
			// 找到正好的区间
			//log.Println("Tick Judge Range", requirehittime, realhittime, realhittime + r[index+1].Time, hit, r[index+1])
			// 寻找正确的光标位置
			// 1.起始点
			// 2.结束点
			// 3.开始和结束算出的中间点
			// 暂时使用中间点
			realhit := getTickRangeJudgePoint(requirehittime, r[index], r[index+1], realhittime)
			isincircle, cserror := isInCircle(realhit, requirepos, CS*CSscale)
			if cserror >= 0 {
				newCSerrors = append(newCSerrors, cserror)
			}
			//realhit := r[index]
			//log.Println("Forward Tick Judge Range Find Require Point", realhit.KeyPressed, realhit, requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(realhit.MosueX), float64(realhit.MouseY)), requirepos), CS * CSscale)
			if isincircle {
				// 在圈内
				// tick判断不会改变占用情况
				// 寻找正确的按下放开位置
				// 1.起始点
				// 2.结束点
				// 3.离哪个近用哪个
				// 暂时使用离哪个近用哪个，最中间点可以算按下或者未按下
				// 暂时最中间算未按下
				presshit := hit
				if requirehittime - realhittime >= realhittime + r[index+1].Time - requirehittime {
					presshit = r[index+1]
				}
				ispressed := isPressedTick(presshit, keysoccupied, keys)
				//log.Println("Tick Judge Range Press", ispressed, presshit.KeyPressed, keysoccupied)
				if ispressed {
					//按下，则击中成功
					return true, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
				}
			}
			return false, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
		}else if realhittime > requirehittime {
			// 时间点已经超过需要的击中时间
			// 1.已经无法击中
			// 2.时间过短在同一个片段被击中
			// 向前一个进行检测
			if start == index {
				index--
				realhittime -= hit.Time
				//log.Println("Forward Tick Judge")
				if realhittime == requirehittime {
					// 找到正好的一点
					isincircle, cserror := isInCircle(hit, requirepos, CS*CSscale)
					if cserror >= 0 {
						newCSerrors = append(newCSerrors, cserror)
					}
					//log.Println("Forward Tick Judge Tap", requirehittime, realhittime, bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos), CS * CSscale)
					if isincircle {
						// 在圈内
						// tick判断不会改变占用情况
						ispressed := isPressedTick(hit, keysoccupied, keys)
						//log.Println("Forward Tick Judge Tap Press", ispressed, hit.KeyPressed, keysoccupied)
						if ispressed {
							//按下，则击中成功
							return true, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
						}
					}
					return false, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
				}else if realhittime < requirehittime && realhittime + r[index+1].Time > requirehittime{
					// 找到正好的区间
					//log.Println("Forward Tick Judge Range", requirehittime, realhittime, realhittime + r[index+1].Time, r[index], r[index+1])
					// 寻找正确的光标位置
					// 1.起始点
					// 2.结束点
					// 3.开始和结束算出的中间点
					// 暂时使用中间点
					realhit := getTickRangeJudgePoint(requirehittime, r[index], r[index+1], realhittime)
					isincircle, cserror := isInCircle(realhit, requirepos, CS*CSscale)
					if cserror >= 0 {
						newCSerrors = append(newCSerrors, cserror)
					}
					//realhit := r[index]
					//log.Println("Forward Tick Judge Range Find Require Point", realhit.KeyPressed, realhit, requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(realhit.MosueX), float64(realhit.MouseY)), requirepos), CS * CSscale)
					if isincircle {
						// 在圈内
						// tick判断不会改变占用情况
						// 寻找正确的按下放开位置
						// 1.起始点
						// 2.结束点
						// 3.离哪个近用哪个
						// 暂时使用离哪个近用哪个
						presshit := hit
						if requirehittime - realhittime > realhittime + r[index+1].Time - requirehittime {
							presshit = r[index+1]
						}
						ispressed := isPressedTick(presshit, keysoccupied, keys)
						//log.Println("Forward Tick Judge Range Press", ispressed, presshit.KeyPressed, keysoccupied)
						if ispressed {
							//按下，则击中成功
							return true, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
						}
					}
					return false, index + 1, realhittime, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
				}
			}

			// 无法击中
			//log.Println("Too late to hit tick", realhittime, requirehittime)
			return false, index, realhittime - r[index].Time, keysoccupied, CSscale, istailjugde, istailhit, keys, newCSerrors
		}
		index++
		time += hit.Time
	}
}

// 判断滑条最终情况
func judgeSlider(requirehits int, realhits int) HitResult {
	// 一个滑条的击中比例
	hitfraction := float64(realhits) / float64(requirehits)
	if hitfraction==1 {
		// 击中比例等于1，输出300
		return Hit300
	}else if hitfraction >=0.5 {
		// 击中比例大于等于0.5，输出100
		return Hit100
	}else if hitfraction >0 {
		// 击中比例大于0，输出50
		return Hit50
	}else {
		// 击中比例为0，输出miss
		return HitMiss
	}
}

// 通过最后时间找第一个滑条头生效位置
func findFirstAfterHead(headtime float64, r []*rplpa.ReplayData) (int, int64) {
	index := 3
	time := r[0].Time + r[1].Time + r[2].Time
	for {
		time += r[index].Time
		if float64(time) > headtime {
			//log.Println("Find FirstbeforeHead before", r[index].Time, time, headtime)
			time -= r[index].Time
			return index - 1, time
		}
		index++
	}
}

// 根据区间上下界计算tick进行区间判定时的准确位置
func getTickRangeJudgePoint(time int64, hit1 *rplpa.ReplayData, hit2 *rplpa.ReplayData, realhittime int64) *rplpa.ReplayData {
	mult := float64(time - realhittime) / float64(hit2.Time)
	return &rplpa.ReplayData{
		Time: time - realhittime,
		MosueX: hit1.MosueX + float32(mult) * (hit2.MosueX - hit1.MosueX),
		MouseY: hit1.MouseY + float32(mult) * (hit2.MouseY - hit1.MouseY),
		KeyPressed: hit1.KeyPressed,
	}
}

// 计算部分的pp
func calculatePPbyNum(filename string, result TotalResult, objnum int) oppai.PPv2 {
	return oppai.PPInfo(score.LoadMapbyNum(filename, objnum), &oppai.Parameters{
		Combo:  result.Combo,
		Mods:   result.Mods,
		N300:   result.N300,
		N100:   result.N100,
		N50:    result.N50,
		Misses: result.Misses,
	}).PP
}

// 判断是否需要修正误差
func shouldfixError(objectindex int, errors []Error) *Error {
	for _, err := range errors {
		if err.ObjectIndex == objectindex {
			// 需要修正
			return &err
		}
	}
	return nil
}

// 修正误差
func fixError(error Error, result []ObjectResult, count300 int, count100 int, count50 int, countMiss int, maxcombo int, nowcombo int, totalhits []int64) (reresult []ObjectResult, recount300 int, recount100 int, recount50 int, recountMiss int, remaxcombo int, renowcombo int, retotalhits []int64){
	lastresult := result[len(result)-1]
	recount300 = count300
	recount100 = count100
	recount50 = count50
	recountMiss = countMiss
	// 修正判定计数和判定数组
	switch lastresult.Result {
	case Hit300:
		log.Println("Fix minus 300")
		recount300 -= 1
		break
	case Hit100:
		log.Println("Fix minus 100")
		recount100 -= 1
		break
	case Hit50:
		log.Println("Fix minus 50")
		recount50 -= 1
		break
	case HitMiss:
		log.Println("Fix minus miss")
		recountMiss -= 1
		break
	}
	switch error.Result {
	case Hit300:
		log.Println("Fix plus 300")
		retotalhits = append(totalhits[:len(totalhits)-2], 300)
		recount300 += 1
		break
	case Hit100:
		log.Println("Fix plus 100")
		retotalhits = append(totalhits[:len(totalhits)-2], 100)
		recount100 += 1
		break
	case Hit50:
		log.Println("Fix plus 50")
		retotalhits = append(totalhits[:len(totalhits)-2], 50)
		recount50 += 1
		break
	case HitMiss:
		log.Println("Fix plus miss")
		retotalhits = append(totalhits[:len(totalhits)-2], 0)
		recountMiss += 1
		break
	}
	// 修正结果数组
	reresult = append(result[:len(result)-2], ObjectResult{lastresult.JudgePos, lastresult.JudgeTime, error.Result, error.IsBreak})
	// 修正combo
	remaxcombo = maxcombo + error.MaxComboOffset
	renowcombo = nowcombo + error.NowComboOffset
	return reresult, recount300, recount100, recount50, recountMiss, remaxcombo, renowcombo, retotalhits
}

// 保留小数点后两位数字
func float2unit(num float64) float64 {
	return math.Ceil(num*100) / 100
}

// 计算不稳定率
func calculateUnstableRate(array []int64) (unstablerate float64) {
	var sum int64
	l := float64(len(array))
	for _, a := range array {
		sum += a
	}
	avg := float64(sum) / l
	for _, a := range array {
		unstablerate += math.Pow(float64(a) - avg, 2)
	}
	return 10 * math.Sqrt(unstablerate / l)
}

// 检查判定个数
func checkHits(true300 int, true100 int, true50 int, trueMiss int, replay300 uint16, replay100 uint16, replay50 uint16, replayMiss uint16) bool {
	return (true300 == int(replay300)) && (true100 == int(replay100)) && (true50 == int(replay50)) && (trueMiss == int(replayMiss))
}

// 判断是否真实为鼠标按键
func trueClick(click bool, key bool) bool {
	return click && !key
}

// 检查replay某一瞬间在不在sliderball中
func checkInSliderBall(slider *objects.Slider, time int64, realpos bmath.Vector2d, scale float64, ispress bool) float64 {
	if bmath.Vector2d.Dst(slider.GetPointAt(time), realpos) < scale && ispress {
		//log.Println("Hit in sliderball", time, slider.GetPointAt(time), realpos, bmath.Vector2d.Dst(slider.GetPointAt(time), realpos), scale + 0.01)
		return TICK_JUDGE_SCALE
	}else {
		//log.Println("Hit but not in sliderball", time, slider.GetPointAt(time), realpos, bmath.Vector2d.Dst(slider.GetPointAt(time), realpos), scale + 0.01)
		return CIRCLE_JUDGE_SCALE
	}
}

// 检查滑条开始到击中前的sliderball
func checkSliderBallBeforeHit(index int, lasttime int64, r []*rplpa.ReplayData, slider *objects.Slider, CS float64, scale float64) (float64, int, []bool, bool, bool) {
	// 击中时间
	realhittime := r[index].Time + lasttime
	// 根据滑条开始时间确定开始查找的replay片段
	// 滑条开始时间
	sliderstarttime := slider.GetBasicData().StartTime
	i := 3
	time := r[0].Time + r[1].Time + r[2].Time
	CSscale := scale

	// 暂时的变量
	// 跳过滑条开始时间后的若干个replay片段？这段时间也许为sliderball的加载时间
	skiptimes := 0

	// 初始tick索引，判定数组
	tickindex := 0
	tickjudge := []bool{}

	// 初始slidertail判定
	slidertailtime := slider.GetBasicData().EndTime - slider.TailJudgeOffset
	tailjudge := false
	tailresult := false

	for {
		time += r[i].Time
		// 大于击中时间，检查结束
		if time > realhittime {
			return CSscale, tickindex, tickjudge, tailjudge, tailresult
		}
		// 大于滑条开始时间，开始检查
		if time >= sliderstarttime {
			if skiptimes == 0 {
				//log.Println("Check sliderball before hit", r[index].Time, time, sliderstarttime)
				CSscale = checkInSliderBall(slider, time, bmath.Vector2d{float64(r[i].MosueX), float64(r[i].MouseY)}, CS*CSscale, isPressedNoTick(r[i]))
				if tickindex < len(slider.ScorePoints) {
					ticktime := slider.ScorePoints[tickindex].Time
					// 如果当前时间已经大于现tick时间，判断tick并指向下一个tick
					if time >= ticktime {
						tickindex++
						if CSscale == TICK_JUDGE_SCALE {
							//log.Println("Early Tick Hit", tickindex, ticktime, r[index].Time, time)
							tickjudge = append(tickjudge, true)
						}else if CSscale == CIRCLE_JUDGE_SCALE {
							//log.Println("Early Tick Not Hit", tickindex, ticktime, r[index].Time, time)
							tickjudge = append(tickjudge, false)
						}
					}
				}
				// 如果当前时间已经大于滑条尾时间，判断滑条尾击中情况
				if time >= slidertailtime {
					tailjudge = true
					if CSscale == TICK_JUDGE_SCALE {
						//log.Println("Early Slider Tail Hit", slidertailtime, r[index].Time, time)
						tailresult = true
					}else if CSscale == CIRCLE_JUDGE_SCALE {
						//log.Println("Early Slider Tail Not Hit", slidertailtime, r[index].Time, time)
						tailresult = false
					}
				}
			}else {
				skiptimes--
			}
		}
		i++
	}
}

// 获取是否重复按下的键位状态数组
func getSameTimePressKeyStatus(key bool, occupied bool) bool {
	return (key && bmath.Xor(key, occupied))
}

// 判断时间点是否不在休息时段中
func isNotInPauses(realhittime int64, requirehittime int64, pauses []objects.BaseObject, offset float64) bool {
	// 休息段会提前一段时间结束并开始判定，这段时间暂时猜测为OD50
	pauseoffset := offset
	for _, pause := range pauses {
		// 击打时间在休息段中，物件时间不在休息段中
		if (realhittime >= pause.GetBasicData().StartTime && float64(realhittime) <= float64(pause.GetBasicData().EndTime) - pauseoffset) && (requirehittime < pause.GetBasicData().StartTime || requirehittime > pause.GetBasicData().EndTime) {
			return false
		}
	}
	return true
}

