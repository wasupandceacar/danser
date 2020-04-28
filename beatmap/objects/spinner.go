package objects

import (
	"danser/audio"
	"danser/bmath"
	. "danser/osuconst"
	"danser/render"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"strconv"
)

type Spinner struct {
	objData         *basicData
	pos             bmath.Vector2d
	Timings         *Timings
	sample          int
	renderStartTime int64
}

func NewSpinner(data []string, number int64) *Spinner {
	spinner := &Spinner{}
	spinner.objData = commonParse(data, number)
	endtime, _ := strconv.ParseInt(data[5], 10, 64)
	spinner.objData.EndTime = int64(endtime)
	spinner.pos = bmath.Vector2d{PLAYFIELD_WIDTH / 2, PLAYFIELD_HEIGHT / 2}

	sample, _ := strconv.ParseInt(data[4], 10, 64)
	spinner.sample = int(sample)

	spinner.renderStartTime = REPLAY_END_TIME
	return spinner
}

func (self Spinner) GetBasicData() *basicData {
	return self.objData
}

func (self *Spinner) SetTiming(timings *Timings) {
	self.Timings = timings
	self.objData.JudgeTime = self.objData.EndTime
}

func (self *Spinner) GetPosition() bmath.Vector2d {
	return self.pos
}

func (self *Spinner) Update(time int64) bool {
	if time < self.objData.EndTime {
		return false
	}

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

func (self *Spinner) Draw(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch) bool {
	if self.renderStartTime == REPLAY_END_TIME {
		self.renderStartTime = time
	}

	alpha := 1.0

	var angle float64

	// 2.5秒之内rpm从0到300
	// 1秒之内rpm从300到0
	if time <= self.objData.StartTime {
		angle = 0.0
	} else if time-self.objData.StartTime <= 2500 {
		rpm := float64(time-self.objData.StartTime) * 0.12
		angle = float64(time-self.objData.StartTime) * (rpm * math.Pi / 30000)
	} else if self.objData.EndTime-time <= 1000 {
		rpm := float64(self.objData.EndTime-time) * 0.3
		angle = float64(self.objData.EndTime-self.objData.StartTime-3500)*math.Pi/100 - float64(self.objData.EndTime-time)*(rpm*math.Pi/30000)
	} else {
		angle = float64(time-self.objData.StartTime-2500) * math.Pi / 100
	}

	if time < self.renderStartTime {
		return false
	} else if time < self.objData.StartTime {
		alpha = float64(color[3]) * clampF((float64(time-self.objData.StartTime)+preempt)/preempt, 0, 1)
	} else if time <= self.objData.EndTime {
		alpha = float64(color[3])
	} else if time <= self.objData.EndTime+int64(preempt)/2 {
		alpha = float64(color[3]) * clampF((float64(self.objData.EndTime-time)+preempt/2)/preempt*2, 0, 1)
	}

	batch.SetTranslation(self.objData.StartPos)

	batch.SetColor(1, 1, 1, alpha)

	if time <= self.objData.EndTime+int64(preempt)/2 {
		// 绘制Spinner转圈
		if (render.SpinnerBackground.Height != 1 || render.SpinnerBackground.Width != 1) || render.SkinVersion < 2.0 {
			// 旧样式的spinner
			spinnerCircleScale := float64(render.SpinnerCircle.Height/render.SpinnerCircle2x) / (DEFAULT_SKIN_SIZE * 2) * PLAYFIELD_HEIGHT
			batch.DrawUnitSR(*render.SpinnerCircle, bmath.Vector2d{float64(render.SpinnerCircle.Width) / float64(render.SpinnerCircle.Height) * spinnerCircleScale, spinnerCircleScale}, angle)
		} else {
			// 新样式的spinner
			// 设置转满的时间
			cleartime := float64(self.objData.EndTime-self.objData.StartTime) * 0.75
			clearmult := clampF(float64(time-self.objData.StartTime)/cleartime, 0, 1)
			clearmult = -clearmult * (clearmult - 2)
			finishScale := 0.8 + 0.2*clearmult
			spinnerTopScale := float64(render.SpinnerTop.Height/render.SpinnerTop2x) / (DEFAULT_SKIN_SIZE * 2) * PLAYFIELD_HEIGHT
			batch.DrawUnitSR(*render.SpinnerTop, bmath.Vector2d{float64(render.SpinnerTop.Width) / float64(render.SpinnerTop.Height) * spinnerTopScale * finishScale, spinnerTopScale * finishScale}, angle)
			spinnerMiddleScale := float64(render.SpinnerMiddle.Height/render.SpinnerMiddle2x) / (DEFAULT_SKIN_SIZE * 2) * PLAYFIELD_HEIGHT
			batch.DrawUnitSR(*render.SpinnerMiddle, bmath.Vector2d{float64(render.SpinnerMiddle.Width) / float64(render.SpinnerMiddle.Height) * spinnerMiddleScale * finishScale, spinnerMiddleScale * finishScale}, angle)
			spinnerBottomScale := float64(render.SpinnerBottom.Height/render.SpinnerBottom2x) / (DEFAULT_SKIN_SIZE * 2) * PLAYFIELD_HEIGHT
			batch.DrawUnitSR(*render.SpinnerBottom, bmath.Vector2d{float64(render.SpinnerBottom.Width) / float64(render.SpinnerBottom.Height) * spinnerBottomScale * finishScale, spinnerBottomScale * finishScale}, angle)
		}
	}

	batch.SetSubScale(1, 1)

	// 绘制Clear
	spincleartime := float64(self.objData.EndTime-self.objData.StartTime)*0.75 + float64(self.objData.StartTime)
	spinclearalpha := 1.0

	if float64(time) <= spincleartime {
		spinclearalpha = 0.0
	} else if float64(time) <= spincleartime+preempt/2 {
		spinclearalpha = float64(color[3]) * clampF(2*(float64(time)-spincleartime)/preempt, 0, 1)
	} else if time <= self.objData.EndTime {
		spinclearalpha = float64(color[3])
	} else if time <= self.objData.EndTime+int64(preempt)/2 {
		spinclearalpha = float64(color[3]) * clampF((float64(self.objData.EndTime-time)+preempt/2)/preempt*2, 0, 1)
	} else {
		spinclearalpha = 0.0
	}

	batch.SetColor(1, 1, 1, spinclearalpha)
	widthratio := float64(render.SpinnerClear.Width) / float64(render.Circle.Width)
	heightratio := float64(render.SpinnerClear.Height) / float64(render.Circle.Height)
	batch.SetNumberScale(widthratio, heightratio)
	batch.DrawUnitN(*render.SpinnerClear, bmath.Vector2d{self.objData.StartPos.X, self.objData.StartPos.Y - 0.25*PLAYFIELD_HEIGHT})

	if time > self.objData.EndTime+int64(preempt)/2 {
		return true
	}
	return false
}

func (self *Spinner) SetDifficulty(preempt, fadeIn float64) {

}

func (self *Spinner) DrawApproach(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch) {
	// 记录第一次渲染转盘的时间，第一次渲染时，转盘正好撑满整个屏幕，并开始淡入。
	// 转盘时间开始，淡入完成，随后转盘逐渐变小
	// 转盘时间结束，开始淡出
	if self.renderStartTime == REPLAY_END_TIME {
		self.renderStartTime = time
	}

	alpha := 1.0
	// 计算AR
	arr := clampF(float64(self.objData.EndTime-time)/float64(self.objData.EndTime-self.objData.StartTime), 0, 1) * PLAYFIELD_HEIGHT / 2

	if time < self.renderStartTime {
		alpha = 0
	} else if time < self.objData.StartTime {
		alpha = float64(color[3]) * clampF((float64(time-self.objData.StartTime)+preempt)/preempt, 0, 1)
	} else if time <= self.objData.EndTime {
		alpha = float64(color[3])
	} else if time <= self.objData.EndTime+int64(preempt)/2 {
		alpha = float64(color[3]) * clampF((float64(self.objData.EndTime-time)+preempt/2)/preempt*2, 0, 1)
	}

	batch.SetTranslation(self.objData.StartPos)

	if time <= self.objData.EndTime+int64(preempt)/2 {
		batch.SetColor(1, 1, 1, alpha)
		batch.DrawUnitS(*render.SpinnerApproachCircle, bmath.Vector2d{arr, arr})
	}
}

func (self *Spinner) GetObjectNumber() int64 {
	return self.objData.ObjectNumber
}
