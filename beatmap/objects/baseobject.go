package objects

import (
	om "danser/bmath"
	. "danser/osuconst"
	"danser/render"
	"danser/settings"
	"github.com/go-gl/mathgl/mgl32"
	"strconv"
	"strings"
)

type BaseObject interface {
	GetBasicData() *basicData
	Update(time int64) bool
	GetPosition() om.Vector2d
	SetDifficulty(preempt, fadeIn float64)
}

type Renderable interface {
	Draw(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch) bool
	DrawApproach(time int64, preempt float64, fadeIn float64, color mgl32.Vec4, batch *render.SpriteBatch)
	GetObjectNumber() int64
}

type basicData struct {
	StartPos, EndPos   om.Vector2d
	StartTime, EndTime int64
	StackOffset        om.Vector2d
	StackIndex         int64
	// 一个combo内的object序号
	Number             int64
	// 总的obejct序号
	ObjectNumber	   int64
	SliderPoint        bool

	sampleSet    int
	additionSet  int
	customIndex  int
	customVolume float64
}

func commonParse(data []string, number int64) *basicData {
	x, _ := strconv.ParseFloat(data[0], 64)
	y, _ := strconv.ParseFloat(data[1], 64)
	if settings.VSplayer.Mods.EnableHR {
		y = PLAYFIELD_HEIGHT - y
	}
	time, _ := strconv.ParseInt(data[2], 10, 64)
	return &basicData{StartPos: om.NewVec2d(x, y), StartTime: time, Number: number}
}

func (bData *basicData) parseExtras(data []string, extraIndex int) {
	if extraIndex < len(data) {
		extras := strings.Split(data[extraIndex], ":")
		sampleSet, _ := strconv.ParseInt(extras[0], 10, 64)
		additionSet, _ := strconv.ParseInt(extras[1], 10, 64)
		index, _ := strconv.ParseInt(extras[2], 10, 64)
		if len(extras) > 3 {
			volume, _ := strconv.ParseInt(extras[3], 10, 64)
			bData.customVolume = float64(volume) / 100.0
		} else {
			bData.customVolume = 0
		}


		bData.sampleSet = int(sampleSet)
		bData.additionSet = int(additionSet)
		bData.customIndex = int(index)
	}
}
