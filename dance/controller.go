package dance

import (
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/dance/movers"
	"danser/dance/schedulers"
	"danser/hitjudge"
	"danser/render"
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

	SetIsHD(isHD bool)
	GetIsHD() bool
	SetIsHR(isHR bool)
	GetIsHR() bool

	SetHitResult(result []hitjudge.ObjectResult)
	GetHitResult() []hitjudge.ObjectResult
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
	isHD        bool
	isHR        bool
	hitresult   []hitjudge.ObjectResult
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

func (controller *ReplayController) SetIsHD(isHD bool) {
	controller.isHD = isHD
}

func (controller *ReplayController) GetIsHD() bool {
	return controller.isHD
}

func (controller *ReplayController) SetIsHR(isHR bool) {
	controller.isHR = isHR
}

func (controller *ReplayController) GetIsHR() bool {
	return controller.isHR
}

func (controller *ReplayController) SetHitResult(result []hitjudge.ObjectResult) {
	controller.hitresult = result
}

func (controller *ReplayController) GetHitResult() []hitjudge.ObjectResult{
	return controller.hitresult
}
