package objects

import (
	"danser/audio"
	"danser/bmath"
	. "danser/osuconst"
	"danser/render"
	"danser/settings"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"strconv"
)

type Circle struct {
	objData *BasicData
	sample  int
	Timings *Timings
}

func NewCircle(data []string, number int64) *Circle {
	circle := &Circle{}
	circle.objData = commonParse(data, number)
	f, _ := strconv.ParseInt(data[4], 10, 64)
	circle.sample = int(f)
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	circle.objData.parseExtras(data, 5)
	return circle
}

func NewCirclebyPath(data []string, number int64, isHR bool) *Circle {
	circle := &Circle{}
	circle.objData = commonParsebyPath(data, number, isHR)
	f, _ := strconv.ParseInt(data[4], 10, 64)
	circle.sample = int(f)
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	circle.objData.parseExtras(data, 5)
	return circle
}

func DummyCircle(pos bmath.Vector2d, time int64) *Circle {
	return DummyCircleInherit(pos, time, false)
}

func DummyCircleInherit(pos bmath.Vector2d, time int64, inherit bool) *Circle {
	circle := &Circle{objData: &BasicData{}}
	circle.objData.StartPos = pos
	circle.objData.EndPos = pos
	circle.objData.StartTime = time
	circle.objData.EndTime = time
	circle.objData.EndPos = circle.objData.StartPos
	circle.objData.SliderPoint = inherit
	return circle
}

func (self Circle) GetBasicData() *BasicData {
	return self.objData
}

func (self *Circle) Update(time int64) bool {

	index := self.objData.customIndex

	if index == 0 {
		index = self.Timings.Current.SampleIndex
	}

	if self.objData.sampleSet == 0 {
		audio.PlaySample(self.Timings.Current.SampleSet, self.objData.additionSet, self.sample, index, self.Timings.Current.SampleVolume)
	} else {
		audio.PlaySample(self.objData.sampleSet, self.objData.additionSet, self.sample, index, self.Timings.Current.SampleVolume)
	}

	return true
}

func (self *Circle) SetTiming(timings *Timings) {
	self.Timings = timings
	self.objData.JudgeTime = self.objData.StartTime
}

func (self *Circle) GetPosition() bmath.Vector2d {
	return self.objData.StartPos
}

func (self *Circle) Draw(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch) bool {
	alpha := 1.0
	fadeInStart := float64(self.objData.StartTime) - preempt
	fadeInEnd := math.Min(float64(self.objData.StartTime), fadeInStart+fadeIn)

	if settings.VSplayer.Mods.EnableHD {
		hiddenFadeInStart := float64(self.objData.StartTime) - preempt
		hiddenFadeInEnd := hiddenFadeInStart + preempt*FADE_IN_DURATION_MULTIPLIER

		hiddenFadeOutStart := hiddenFadeInEnd
		hiddenFadeOutEnd := hiddenFadeOutStart + preempt*FADE_OUT_DURATION_MULTIPLIER
		if float64(time) < hiddenFadeInEnd && float64(time) >= hiddenFadeInStart {
			alpha = Clamp(1.0-(hiddenFadeInEnd-float64(time))/(hiddenFadeInEnd-hiddenFadeInStart), 0.0, 1.0)
		} else if float64(time) >= hiddenFadeOutStart {
			alpha = Clamp((hiddenFadeOutEnd-float64(time))/(hiddenFadeOutEnd-hiddenFadeOutStart), 0.0, 1.0)
		} else {
			alpha = float64(color[3])
		}
	} else {
		if time < self.objData.StartTime && float64(time) >= fadeInStart {
			alpha = Clamp(1.0-(fadeInEnd-float64(time))/fadeIn, 0.0, 1.0)
		} else if time >= self.objData.StartTime {
			alpha = Clamp(1.0-float64(time-self.objData.StartTime)/(preempt/2), 0.0, 1.0)
		} else {
			alpha = float64(color[3])
		}
	}

	batch.SetTranslation(self.objData.StartPos)

	if time >= self.objData.StartTime {
		batch.SetSubScale(1+(1.0-alpha)*0.5, 1+(1.0-alpha)*0.5)
	}

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		batch.DrawUnit(*render.CircleFull)
	} else {
		batch.DrawUnit(*render.Circle)
	}

	if settings.VSplayer.PlayerFieldUI.ShowHitCircleNumber {
		// 绘制圈内数字
		widthratio := float64(render.Circle0.Width) / float64(render.Circle.Width)
		heightratio := float64(render.Circle0.Height) / float64(render.Circle.Height)
		batch.SetNumberScale(widthratio, heightratio)
		batch.SetColor(1, 1, 1, alpha)

		if self.objData.Number < 10 {
			// 编号一位数
			DrawHitCircleNumber(self.objData.Number, self.objData.StartPos, batch)
		} else {
			// 只考虑编号两位数的情况

			// 计算十位数和个位数
			tenDigit := self.objData.Number / 10
			unitDigit := self.objData.Number % 10

			// 计算十位数和个位数的位置
			screenratio := PLAYFIELD_HEIGHT / 600
			baseX := self.objData.StartPos.X
			baseY := self.objData.StartPos.Y
			tenDigitWidth := int64(GetHitCircleNumberWidth(tenDigit))
			unitDigitWidth := int64(GetHitCircleNumberWidth(unitDigit))
			tenBaseX := baseX + float64(render.HitCircleOverlap-tenDigitWidth)/2*screenratio
			unitBaseX := baseX - float64(render.HitCircleOverlap-unitDigitWidth)/2*screenratio

			DrawHitCircleNumber(tenDigit, bmath.Vector2d{tenBaseX, baseY}, batch)
			DrawHitCircleNumber(unitDigit, bmath.Vector2d{unitBaseX, baseY}, batch)
		}
	}

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {
		batch.SetColor(1, 1, 1, alpha)
		batch.DrawUnit(*render.CircleOverlay)
	}

	batch.SetSubScale(1, 1)

	if time >= self.objData.StartTime+int64(preempt/2) {
		return true
	}
	return false
}

func (self *Circle) SetDifficulty(preempt, fadeIn float64) {

}

func (self *Circle) DrawApproach(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch) {
	alpha := 1.0
	arr := float64(self.objData.StartTime-time) / preempt

	approachCircleFadeInStart := float64(self.objData.StartTime) - preempt
	approachCircleFadeInEnd := math.Min(float64(self.objData.StartTime), approachCircleFadeInStart+2*fadeIn)

	if time < self.objData.StartTime && float64(time) >= approachCircleFadeInStart {
		alpha = Clamp(1.0-(approachCircleFadeInEnd-float64(time))/fadeIn, 0.0, 1.0)
	} else if time >= self.objData.StartTime {
		alpha = Clamp(1.0-float64(time-self.objData.StartTime)/(preempt/2), 0.0, 1.0)
	} else {
		alpha = float64(color[3])
	}

	batch.SetTranslation(self.objData.StartPos)

	if settings.Objects.DrawApproachCircles && time <= self.objData.StartTime {
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
		batch.SetSubScale(1.0+arr*2, 1.0+arr*2)
		batch.DrawUnitFix(*render.ApproachCircle, float64(128*render.ApproachCircle2x), float64(128*render.ApproachCircle2x))
	}

	batch.SetSubScale(1, 1)
}

func (self *Circle) GetObjectNumber() int64 {
	return self.objData.ObjectNumber
}
