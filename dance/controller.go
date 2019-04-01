package dance

import (
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/dance/movers"
	"danser/dance/schedulers"
	"danser/hitjudge"
	. "danser/osuconst"
	"danser/render"
	"danser/render/texture"
	"danser/settings"
	"github.com/Mempler/rplpa"
	"strings"
)

type Controller interface {
	SetBeatMap(beatMap *beatmap.BeatMap)
	InitCursors()
	//Update(time int64, delta float64)

	/////////////////////////////////////////////////////////////////////////////////////////////////////
	// 改成更多参数
	Update(time int64, delta float64, position bmath.Vector2d)
	/////////////////////////////////////////////////////////////////////////////////////////////////////

	GetCursors() []*render.Cursor

	SetPlayername(playername string)
	GetPlayname() string

	SetPresskey(presskey rplpa.KeyPressed)
	GetPresskey() rplpa.KeyPressed

	SetMods(mods int)
	GetMods() int

	SetHitResult(result []hitjudge.ObjectResult)
	GetHitResult() []hitjudge.ObjectResult

	SetTotalResult(result []hitjudge.TotalResult)
	GetTotalResult() []hitjudge.TotalResult

	SetAcc(result float64)
	GetAcc() float64

	SetPP(result float64)
	GetPP() float64

	SetUR(result float64)
	GetUR() float64

	SetRank(result texture.TextureRegion)
	GetRank() texture.TextureRegion

	SetIsShow(isShow bool)
	GetIsShow() bool

	SetDishowTime(time float64)
	GetDishowTime() float64

	SetDishowPos(pos bmath.Vector2d, rate int)
	GetDishowPos() bmath.Vector2d

	AddMissInfo(misstime float64, missjudgetime int64, misspos bmath.Vector2d, rate int)
	GetMissInfo() []missInfo

	IsInMiss(time int64) bool
}

var Mover = movers.NewAngleOffsetMover

func SetMover(name string) {
	name = strings.ToLower(name)

	if name == "bezier" {
		Mover = movers.NewBezierMover
	} else if name == "circular" {
		Mover = movers.NewHalfCircleMover
	} else if name == "linear" {
		Mover = movers.NewLinearMover
	} else if name == "axis" {
		Mover = movers.NewAxisMover
	} else {
		Mover = movers.NewAngleOffsetMover
	}
}

//type GenericController struct {
//	bMap       *beatmap.BeatMap
//	cursors    []*render.Cursor
//	schedulers []schedulers.Scheduler
//}
//
//func NewGenericController() Controller {
//	return &GenericController{}
//}
//
//func (controller *GenericController) SetBeatMap(beatMap *beatmap.BeatMap) {
//	controller.bMap = beatMap
//}
//
//func (controller *GenericController) InitCursors() {
//	controller.cursors = make([]*render.Cursor, settings.TAG)
//	controller.schedulers = make([]schedulers.Scheduler, settings.TAG)
//
//	for i := range controller.cursors {
//		controller.cursors[i] = render.NewCursor()
//		controller.schedulers[i] = schedulers.NewGenericScheduler(Mover)
//	}
//
//	type Queue struct {
//		objs []objects.BaseObject
//	}
//
//	objs := make([]Queue, settings.TAG)
//
//	queue := controller.bMap.GetObjectsCopy()
//
//	if settings.Dance.TAGSliderDance && settings.TAG > 1 {
//		for i := 0; i < len(queue); i++ {
//			queue = schedulers.PreprocessQueue(i, queue, true)
//		}
//	}
//
//	for j, o := range queue {
//		i := j % settings.TAG
//		objs[i].objs = append(objs[i].objs, o)
//	}
//
//	for i := range controller.cursors {
//		controller.schedulers[i].Init(objs[i].objs, controller.cursors[i])
//	}
//
//}
//
///////////////////////////////////////////////////////////////////////////////////////////////////////
//// 更新时间和光标位置
//func (controller *GenericController) Update(time int64, delta float64) {
//	for i := range controller.cursors {
//		controller.schedulers[i].Update(time)
//		controller.cursors[i].Update(delta)
//	}
//}
///////////////////////////////////////////////////////////////////////////////////////////////////////
//
//func (controller *GenericController) GetCursors() []*render.Cursor {
//	return controller.cursors
//}





/////////////////////////////////////////////////////////////////////////////////////////////////////
// 重写replay的controller
/////////////////////////////////////////////////////////////////////////////////////////////////////

type ReplayController struct {
	bMap       *beatmap.BeatMap
	cursors    []*render.Cursor
	schedulers []schedulers.Scheduler
	playername  string
	presskey    rplpa.KeyPressed
	mods		int
	hitresult   []hitjudge.ObjectResult
	totalresult []hitjudge.TotalResult
	acc  		float64
	rank 		texture.TextureRegion
	isShow		bool
	dishowtime	float64
	dishowpos	bmath.Vector2d
	missinfo	[]missInfo
	pp			float64
	ur			float64
}

type missInfo struct {
	MissTime	float64
	MissJudgeTime	int64
	MissPos		bmath.Vector2d
}

func NewReplayController() Controller {
	return &ReplayController{}
}

func (controller *ReplayController) SetBeatMap(beatMap *beatmap.BeatMap) {
	controller.bMap = beatMap
}

func (controller *ReplayController) InitCursors() {
	controller.cursors = make([]*render.Cursor, settings.TAG)
	controller.schedulers = make([]schedulers.Scheduler, settings.TAG)

	for i := range controller.cursors {
		controller.cursors[i] = render.NewCursor()
		controller.schedulers[i] = schedulers.NewReplayScheduler()
	}

	type Queue struct {
		objs []objects.BaseObject
	}

	objs := make([]Queue, settings.TAG)

	queue := controller.bMap.GetObjectsCopy()

	//if settings.Dance.TAGSliderDance && settings.TAG > 1 {
	//	for i := 0; i < len(queue); i++ {
	//		queue = schedulers.PreprocessQueue(i, queue, true)
	//	}
	//}

	for j, o := range queue {
		i := j % settings.TAG
		objs[i].objs = append(objs[i].objs, o)
	}

	for i := range controller.cursors {
		controller.schedulers[i].Init(objs[i].objs, controller.cursors[i])
	}

}

/////////////////////////////////////////////////////////////////////////////////////////////////////
// 更新时间和光标位置
func (controller *ReplayController) Update(time int64, delta float64, position bmath.Vector2d) {
	for i := range controller.cursors {
		controller.schedulers[i].Update(time, position)
		controller.cursors[i].Update(delta)
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////////

func (controller *ReplayController) GetCursors() []*render.Cursor {
	return controller.cursors
}

func (controller *ReplayController) SetPlayername(playername string) {
	controller.playername = playername
}

func (controller *ReplayController) GetPlayname() string {
	return controller.playername
}

func (controller *ReplayController) SetPresskey(presskey rplpa.KeyPressed) {
	controller.presskey = presskey
}

func (controller *ReplayController) GetPresskey() rplpa.KeyPressed {
	return controller.presskey
}

func (controller *ReplayController) SetMods(mods int) {
	controller.mods = mods
}

func (controller *ReplayController) GetMods() int {
	return controller.mods
}

func (controller *ReplayController) SetHitResult(result []hitjudge.ObjectResult) {
	controller.hitresult = result
}

func (controller *ReplayController) GetHitResult() []hitjudge.ObjectResult{
	return controller.hitresult
}

func (controller *ReplayController) SetTotalResult(result []hitjudge.TotalResult) {
	controller.totalresult = result
}

func (controller *ReplayController) GetTotalResult() []hitjudge.TotalResult{
	return controller.totalresult
}

func (controller *ReplayController) SetAcc(result float64) {
	controller.acc = result
}

func (controller *ReplayController) GetAcc() float64{
	return controller.acc
}

func (controller *ReplayController) SetPP(result float64) {
	controller.pp = result
}

func (controller *ReplayController) GetPP() float64{
	return controller.pp
}

func (controller *ReplayController) SetUR(result float64) {
	controller.ur = result
}

func (controller *ReplayController) GetUR() float64{
	return controller.ur
}

func (controller *ReplayController) SetRank(result texture.TextureRegion) {
	controller.rank = result
}

func (controller *ReplayController) GetRank() texture.TextureRegion{
	return controller.rank
}

func (controller *ReplayController) SetIsShow(isShow bool) {
	controller.isShow = isShow
}

func (controller *ReplayController) GetIsShow() bool {
	return controller.isShow
}

func (controller *ReplayController) SetDishowTime(time float64) {
	controller.dishowtime = time
}

func (controller *ReplayController) GetDishowTime() float64{
	return controller.dishowtime
}

func (controller *ReplayController) SetDishowPos(pos bmath.Vector2d, rate int) {
	// 向下一个偏移
	offsetY := settings.VSplayer.Knockout.SameTimeOffset * float64(rate)
	controller.dishowpos = bmath.Vector2d{pos.X, PLAYFIELD_HEIGHT - pos.Y - offsetY}
}

func (controller *ReplayController) GetDishowPos() bmath.Vector2d{
	return controller.dishowpos
}

func (controller *ReplayController) AddMissInfo(misstime float64, missjudgetime int64, misspos bmath.Vector2d, rate int) {
	// 向下一个偏移
	offsetY := settings.VSplayer.Knockout.SameTimeOffset * float64(rate)
	controller.missinfo = append(controller.missinfo, missInfo{misstime, missjudgetime, bmath.Vector2d{misspos.X, PLAYFIELD_HEIGHT - misspos.Y - offsetY}})
}

func (controller *ReplayController) GetMissInfo() []missInfo {
	return controller.missinfo
}

// 检查是否已经录入
func (controller *ReplayController) IsInMiss(time int64) bool {
	for _, missinfo := range controller.missinfo {
		if time == missinfo.MissJudgeTime {
			return false
		}
	}
	return true
}
