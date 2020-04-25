package resultcache

import (
	"danser/hitjudge"
	"danser/settings"
	"encoding/json"
	"github.com/Mempler/rplpa"
	"io/ioutil"
	"os"
)

func CacheResult(oresult []hitjudge.ObjectResult, tresult []hitjudge.TotalResult, replay *rplpa.Replay) {
	filename := settings.VSplayer.ReplayandCache.CacheDir + replay.ReplayMD5
	oerr := ioutil.WriteFile(filename+".ooc", []byte(marshalObjectResult(oresult)), 0666)
	if oerr != nil {
		panic(oerr)
	}
	terr := ioutil.WriteFile(filename+".otc", []byte(marshalTotalResult(tresult)), 0666)
	if terr != nil {
		panic(terr)
	}
}

func GetResult(replay *rplpa.Replay) ([]hitjudge.ObjectResult, []hitjudge.TotalResult) {
	filename := settings.VSplayer.ReplayandCache.CacheDir + replay.ReplayMD5
	oread, _ := ioutil.ReadFile(filename + ".ooc")
	tread, _ := ioutil.ReadFile(filename + ".otc")
	return unmarshalObjectResult(oread), unmarshalTotalResult(tread)
}

func IsCacheExists(replay *rplpa.Replay) bool {
	filename := settings.VSplayer.ReplayandCache.CacheDir + replay.ReplayMD5
	_, err1 := os.Stat(filename + ".ooc")
	_, err2 := os.Stat(filename + ".otc")
	return !os.IsNotExist(err1) && !os.IsNotExist(err2)
}

func marshalObjectResult(oresult []hitjudge.ObjectResult) string {
	data, err := json.MarshalIndent(oresult, "", "     ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func marshalTotalResult(tresult []hitjudge.TotalResult) string {
	data, err := json.MarshalIndent(tresult, "", "     ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func unmarshalObjectResult(r []byte) []hitjudge.ObjectResult {
	var oresult []hitjudge.ObjectResult
	if err := json.Unmarshal(r, &oresult); err != nil {
		panic(err)
	}
	return oresult
}

func unmarshalTotalResult(r []byte) []hitjudge.TotalResult {
	var tresult []hitjudge.TotalResult
	if err := json.Unmarshal(r, &tresult); err != nil {
		panic(err)
	}
	return tresult
}
