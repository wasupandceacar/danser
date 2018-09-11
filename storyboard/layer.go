package storyboard

import (
	"sort"
	"sync"
	"github.com/wieku/danser/render"
)

type StoryboardLayer struct {
	spriteQueue     []Object
	spriteProcessed []Object
	drawArray       []Object
	visibleObjects  int
	mutex           *sync.Mutex
}

func NewStoryboardLayer() *StoryboardLayer {
	return &StoryboardLayer{mutex: &sync.Mutex{}}
}

func (layer *StoryboardLayer) Add(object Object) {
	layer.spriteQueue = append(layer.spriteQueue, object)
}

func (layer *StoryboardLayer) FinishLoading() {
	sort.Slice(layer.spriteQueue, func(i, j int) bool {
		return layer.spriteQueue[i].GetStartTime() < layer.spriteQueue[j].GetStartTime()
	})
	layer.drawArray = make([]Object, len(layer.spriteQueue))
}

func (layer *StoryboardLayer) Update(time int64) {
	added := false
	toRemove := 0

	for i := 0; i < len(layer.spriteQueue); i++ {
		c := layer.spriteQueue[i]
		if c.GetStartTime() > time {
			break
		}

		layer.spriteProcessed = append(layer.spriteProcessed, c)
		added = true
		toRemove++
	}

	if added {
		sort.Slice(layer.spriteProcessed, func(i, j int) bool {
			return layer.spriteProcessed[i].GetZIndex() < layer.spriteProcessed[j].GetZIndex()
		})
	}

	if toRemove > 0 {
		layer.spriteQueue = layer.spriteQueue[toRemove:]
	}

	layer.mutex.Lock()

	for i := 0; i < len(layer.spriteProcessed); i++ {
		c := layer.spriteProcessed[i]
		c.Update(time)

		if time >= c.GetEndTime() {
			layer.spriteProcessed = append(layer.spriteProcessed[:i], layer.spriteProcessed[i+1:]...)
			i--
		}
	}

	layer.visibleObjects = len(layer.spriteProcessed)
	copy(layer.drawArray, layer.spriteProcessed)

	layer.mutex.Unlock()
}

func (layer *StoryboardLayer) Draw(time int64, batch *render.SpriteBatch) {
	layer.mutex.Lock()

	for i := 0; i < layer.visibleObjects; i++ {
		layer.drawArray[i].Draw(time, batch)
	}

	layer.mutex.Unlock()
}