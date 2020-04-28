package resultcache

import (
	"danser/build"
	"danser/hitjudge"
	"danser/settings"
	"encoding/json"
	"github.com/Mempler/rplpa"
	"io/ioutil"
	"log"
)

type Cache struct {
	ObjectResults []hitjudge.ObjectResult
	TotalResults  []hitjudge.TotalResult
	Version       int
}

func CacheResult(objectResults []hitjudge.ObjectResult, totalResults []hitjudge.TotalResult, rep *rplpa.Replay) {
	err := ioutil.WriteFile(settings.VSplayer.ReplayandCache.CacheDir+rep.ReplayMD5+".oac", marshalCache(Cache{ObjectResults: objectResults, TotalResults: totalResults, Version: build.CACHE_VERSION}), 0666)
	if err != nil {
		panic(err)
	}
}

func GetResult(rep *rplpa.Replay) ([]hitjudge.ObjectResult, []hitjudge.TotalResult, bool) {
	bytes, err := ioutil.ReadFile(settings.VSplayer.ReplayandCache.CacheDir + rep.ReplayMD5 + ".oac")
	if err != nil {
		log.Printf("Could not find ot could not access cache file for %v's replay, MD5 = %v", rep.Username, rep.ReplayMD5)
		return nil, nil, false
	}
	cache := unmarshalCache(bytes)
	if cache.Version != build.CACHE_VERSION {
		log.Printf("Detected unmatched CACHE_VERSION in %v.oac (%v), overwriting...", rep.ReplayMD5, rep.Username)
		return nil, nil, false
	}
	return cache.ObjectResults, cache.TotalResults, true
}

func marshalCache(cache Cache) []byte {
	data, err := json.Marshal(cache)
	if err != nil {
		panic(err)
	}
	return data
}

func unmarshalCache(r []byte) Cache {
	var cache Cache
	if err := json.Unmarshal(r, &cache); err != nil {
		panic(err)
	}
	return cache
}
