package objects

import (
	"danser/bmath"
	. "danser/osuconst"
	"strconv"
)

type Pause struct {
	objData *BasicData
}

func NewPause(data []string) *Pause {
	pause := &Pause{}
	pause.objData = &BasicData{}
	pause.objData.StartTime, _ = strconv.ParseInt(data[1], 10, 64)
	pause.objData.EndTime, _ = strconv.ParseInt(data[2], 10, 64)
	pause.objData.StartPos = bmath.NewVec2d(PLAYFIELD_WIDTH/2, PLAYFIELD_HEIGHT/2)
	pause.objData.EndPos = pause.objData.StartPos
	pause.objData.Number = -1
	return pause
}

func (self Pause) GetBasicData() *BasicData {
	return self.objData
}

func (self *Pause) SetDifficulty(preempt, fadeIn float64) {

}

func (self *Pause) Update(time int64) bool {
	return time >= self.objData.EndTime
}

func (self *Pause) GetPosition() bmath.Vector2d {
	return self.objData.StartPos
}

func (self *Pause) GetObjectNumber() int64 {
	return -1
}
