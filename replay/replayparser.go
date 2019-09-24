package replay

import (
	"github.com/Mempler/rplpa"
	"io/ioutil"
)

func ExtractReplay(name string) *rplpa.Replay {
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	replay, err := rplpa.ParseReplay(buf)
	if err != nil {
		panic(err)
	}
	return replay
}