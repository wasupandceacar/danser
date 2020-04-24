package resultcache

import (
	"danser/hitjudge"
	"danser/settings"
	"encoding/json"
	"github.com/Mempler/rplpa"
	"io/ioutil"
)

func SaveResult(oresult []hitjudge.ObjectResult, tresult []hitjudge.TotalResult, replay *rplpa.Replay) {
	filename := settings.VSplayer.ReplayandCache.CacheDir + replay.BeatmapMD5
	oerr := ioutil.WriteFile(filename+".ooc", []byte(getObjectCache(oresult)), 0666)
	if oerr != nil {
		panic(oerr)
	}
	terr := ioutil.WriteFile(filename+".otc", []byte(getTotalCache(tresult)), 0666)
	if terr != nil {
		panic(terr)
	}
}

func ReadResult(replay *rplpa.Replay) ([]hitjudge.ObjectResult, []hitjudge.TotalResult) {
	filename := settings.VSplayer.ReplayandCache.CacheDir + replay.BeatmapMD5
	oread, _ := ioutil.ReadFile(filename + ".ooc")
	tread, _ := ioutil.ReadFile(filename + ".otc")
	return setObjectCache(oread), setTotalCache(tread)
}

func getObjectCache(oresult []hitjudge.ObjectResult) string {
	data, err := json.MarshalIndent(oresult, "", "     ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func getTotalCache(tresult []hitjudge.TotalResult) string {
	data, err := json.MarshalIndent(tresult, "", "     ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func setObjectCache(r []byte) []hitjudge.ObjectResult {
	var oresult []hitjudge.ObjectResult
	if err := json.Unmarshal(r, &oresult); err != nil {
		panic(err)
	}
	return oresult
}

func setTotalCache(r []byte) []hitjudge.TotalResult {
	var tresult []hitjudge.TotalResult
	if err := json.Unmarshal(r, &tresult); err != nil {
		panic(err)
	}
	return tresult
}
