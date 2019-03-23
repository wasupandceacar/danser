package objects

import (
	"danser/bmath"
	. "danser/osuconst"
	"danser/render"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"strconv"
)

type Spinner struct {
	objData *basicData
	pos     bmath.Vector2d
	Timings *Timings
	renderStartTime int64
}

func NewSpinner(data []string, number int64) *Spinner {
	spinner := &Spinner{}
	spinner.objData = commonParse(data, number)
	endtime, _ := strconv.ParseInt(data[5], 10, 64)
	spinner.objData.EndTime = int64(endtime)
	spinner.pos = bmath.Vector2d{PLAYFIELD_WIDTH / 2,PLAYFIELD_HEIGHT / 2}
	spinner.renderStartTime = -12345
	return spinner
}

func (self Spinner) GetBasicData() *basicData {
	return self.objData
}

func (self *Spinner) SetTiming(timings *Timings) {
	self.Timings = timings
}

func (self *Spinner) GetPosition() bmath.Vector2d {
	return self.pos
}

func (self *Spinner) Update(time int64) bool {
	return true
}

func (self *Spinner) Draw(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch) bool {
	if self.renderStartTime == -12345 {
		self.renderStartTime = time
	}

	alpha := 1.0

	// 计算角度，设定Spinner为300rpm的转圈
	angle := float64(time - self.renderStartTime) * math.Pi / 100

	if time < self.renderStartTime - int64(preempt) {
		return false
	} else if time < self.renderStartTime {
		alpha = float64(color[3]) / preempt
	}else {
		alpha = float64(color[3])
	}

	batch.SetTranslation(self.objData.StartPos)

	batch.SetColor(1, 1, 1, alpha)
	batch.DrawUnitSR(*render.SpinnerBottom, bmath.Vector2d{float64(render.SpinnerBottom.Width) / 4, float64(render.SpinnerBottom.Height) / 4}, angle)

	batch.SetSubScale(1, 1)

	if time >= self.objData.EndTime+int64(preempt/4) {
		return true
	}
	return false
}

func (self *Spinner) SetDifficulty(preempt, fadeIn float64) {

}

func (self *Spinner) DrawApproach(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch) {
	// 记录第一次渲染转盘的时间，第一次渲染时，转盘正好撑满整个屏幕，随后逐渐变小
	if self.renderStartTime == -12345 {
		self.renderStartTime = time
	}

	alpha := 1.0
	// 计算AR
	fake_preempt := 2 * float64(self.objData.EndTime - self.renderStartTime) / PLAYFIELD_HEIGHT
	arr := float64(self.objData.EndTime - time) / fake_preempt

	// 计算角度，设定Spinner为300rpm的转圈
	angle := float64(time - self.renderStartTime) * math.Pi / 100

	if time < self.renderStartTime - int64(preempt){
		alpha = 0
	} else if time < self.renderStartTime{
		alpha = float64(color[3]) / preempt
	}else {
		alpha = float64(color[3])
	}

	batch.SetTranslation(self.objData.StartPos)

	if time <= self.objData.EndTime {
		batch.SetColor(1, 1, 1, alpha)
		batch.DrawUnitS(*render.SpinnerApproachCircle, bmath.Vector2d{arr, arr})
		// 绘制Spinner转圈
		batch.DrawUnitSR(*render.SpinnerCircle, bmath.Vector2d{float64(render.SpinnerCircle.Width) / 4.75, float64(render.SpinnerCircle.Height) / 4.75}, angle)
		batch.DrawUnitSR(*render.SpinnerMiddle, bmath.Vector2d{float64(render.SpinnerMiddle.Width) / 2, float64(render.SpinnerMiddle.Height) / 2}, angle)
	}
}