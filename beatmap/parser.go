package beatmap

import (
	"bufio"
	"danser/beatmap/objects"
	"danser/settings"
	"os"
	"sort"
	"strconv"
	"strings"
)

func parseGeneral(line []string, beatMap *BeatMap) bool {

	switch line[0] {
	case "Mode":
		if line[1] != "0" {
			return true
		}
		break
	case "StackLeniency":
		beatMap.StackLeniency, _ = strconv.ParseFloat(line[1], 64)
		break
	case "AudioFilename":
		beatMap.Audio += line[1]
		break
	case "SampleSet":
		switch line[1] {
		case "Normal", "All":
			beatMap.Timings.BaseSet = 1
			break
		case "Soft":
			beatMap.Timings.BaseSet = 2
			break
		case "Drum":
			beatMap.Timings.BaseSet = 3
			break
		}
		beatMap.Timings.LastSet = beatMap.Timings.BaseSet
		break
	}

	return false
}

func parseMetadata(line []string, beatMap *BeatMap) {
	switch line[0] {
	case "Title":
		beatMap.Name = line[1]
		break
	case "TitleUnicode":
		beatMap.NameUnicode = line[1]
		break
	case "Artist":
		beatMap.Artist = line[1]
		break
	case "ArtistUnicode":
		beatMap.ArtistUnicode = line[1]
		break
	case "Creator":
		beatMap.Creator = line[1]
		break
	case "Version":
		beatMap.Difficulty = line[1]
		break
	case "Source":
		beatMap.Source = line[1]
		break
	case "Tags":
		beatMap.Tags = line[1]
		break
	}
}

func parseDifficulty(line []string, beatMap *BeatMap) {
	if line[0] == "SliderMultiplier" {
		beatMap.SliderMultiplier, _ = strconv.ParseFloat(line[1], 64)
		beatMap.Timings.SliderMult = float64(beatMap.SliderMultiplier)
	}
	if line[0] == "ApproachRate" {
		beatMap.ApproachRate, _ = strconv.ParseFloat(line[1], 64)
	}
	// 加入OD
	if line[0] == "OverallDifficulty" {
		beatMap.OverallDifficulty, _ = strconv.ParseFloat(line[1], 64)
	}
	if line[0] == "CircleSize" {
		beatMap.CircleSize, _ = strconv.ParseFloat(line[1], 64)
	}
	if line[0] == "SliderTickRate" {
		beatMap.Timings.TickRate, _ = strconv.ParseFloat(line[1], 64)
	}

}

func parseEvents(line []string, beatMap *BeatMap) {
	if line[0] == "0" {
		beatMap.Bg += strings.Replace(line[2], "\"", "", -1)
	}
	if line[0] == "2" {
		beatMap.PausesText += line[1] + "," + line[2]
		beatMap.Pauses = append(beatMap.Pauses, objects.NewPause(line))
	}
}

func parseHitObjects(line []string, beatMap *BeatMap, number int64) int64 {
	obj, newnumber := objects.GetObject(line, number)

	if obj != nil {
		if o, ok := obj.(*objects.Slider); ok {
			o.SetTiming(beatMap.Timings)
		}
		if o, ok := obj.(*objects.Circle); ok {
			o.SetTiming(beatMap.Timings)
		}
		if o, ok := obj.(*objects.Spinner); ok {
			o.SetTiming(beatMap.Timings)
		}
		beatMap.HitObjects = append(beatMap.HitObjects, obj)
	}
	return newnumber
}

func parseHitObjectsbyPath(line []string, beatMap *BeatMap, number int64, isHR bool) int64 {
	obj, newnumber := objects.GetObjectbyPath(line, number, isHR)

	if obj != nil {
		if o, ok := obj.(*objects.Slider); ok {
			o.SetTiming(beatMap.Timings)
		}
		if o, ok := obj.(*objects.Circle); ok {
			o.SetTiming(beatMap.Timings)
		}
		if o, ok := obj.(*objects.Spinner); ok {
			o.SetTiming(beatMap.Timings)
		}
		beatMap.HitObjects = append(beatMap.HitObjects, obj)
	}
	return newnumber
}

func tokenize(line, delimiter string) []string {
	if strings.HasPrefix(line, "//") || !strings.Contains(line, delimiter) {
		return nil
	}
	divided := strings.Split(line, delimiter)
	for i, a := range divided {
		divided[i] = strings.TrimSpace(a)
	}
	return divided
}

func tokenizeintergal(line, delimiter string) []string {
	if strings.HasPrefix(line, "//") || !strings.Contains(line, delimiter) {
		return nil
	}
	divided := strings.Split(line, delimiter)
	for i, a := range divided {
		divided[i] = strings.TrimSpace(a)
	}
	// 防止属性值整体因为含有分割符而被分割
	if len(divided) > 2 {
		divided[1] = strings.Join(divided[1:], "")
	}
	return divided
}

func getSection(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") {
		return strings.TrimRight(strings.TrimLeft(line, "["), "]")
	}
	return ""
}

func ParseBeatMap(file *os.File) *BeatMap {
	scanner := bufio.NewScanner(file)
	beatMap := NewBeatMap()
	var currentSection string
	counter := 0
	counter1 := 0

	for scanner.Scan() {
		line := scanner.Text()

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "General":
			if arr := tokenize(line, ":"); len(arr) > 1 {
				if err := parseGeneral(arr, beatMap); err {
					return nil
				}
			}
			break
		case "Metadata":
			if arr := tokenizeintergal(line, ":"); len(arr) > 1 {
				parseMetadata(arr, beatMap)
			}
			break
		case "Difficulty":
			if arr := tokenize(line, ":"); len(arr) > 1 {
				parseDifficulty(arr, beatMap)
			}
			break
		case "Events":
			if arr := tokenize(line, ","); len(arr) > 1 {
				if arr[0] == "2" {
					if counter1 > 0 {
						beatMap.PausesText += ","
					}
					counter1++
				}

				parseEvents(arr, beatMap)
			}
			break
		case "TimingPoints":
			if arr := tokenize(line, ","); len(arr) > 1 {
				if counter > 0 {
					beatMap.TimingPoints += "|"
				}
				counter++

				beatMap.TimingPoints += line
			}
			break
		}
	}

	beatMap.LoadTimingPoints()

	file.Seek(0, 0)

	if beatMap.Name+beatMap.Artist+beatMap.Creator == "" || beatMap.TimingPoints == "" {
		return nil
	}

	return beatMap
}

func ParseObjects(beatMap *BeatMap) {

	file, err := os.Open(settings.General.OsuSongsDir + string(os.PathSeparator) + beatMap.Dir + string(os.PathSeparator) + beatMap.File)
	defer file.Close()

	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)

	var currentSection string
	// object的编号
	var objectNumber int64
	for scanner.Scan() {
		line := scanner.Text()

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "HitObjects":
			if arr := tokenize(line, ","); arr != nil {
				objectNumber = parseHitObjects(arr, beatMap, objectNumber)
			}
			break
		}
	}

	sort.Slice(beatMap.HitObjects, func(i, j int) bool {
		return beatMap.HitObjects[i].GetBasicData().StartTime < beatMap.HitObjects[j].GetBasicData().StartTime
	})

	// 计算总的序号
	num := 0

	for _, o := range beatMap.HitObjects {
		_, ok := o.(*objects.Pause)

		if !ok {
			o.GetBasicData().ObjectNumber = int64(num)
			num++
		}

	}

	calculateStackLeniency(beatMap)
}

func ParseObjectsByPath(beatMap *BeatMap, filename string, isHR bool, isEZ bool) {

	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)

	var currentSection string
	// object的编号
	var objectNumber int64
	for scanner.Scan() {
		line := scanner.Text()

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "HitObjects":
			if arr := tokenize(line, ","); arr != nil {
				objectNumber = parseHitObjectsbyPath(arr, beatMap, objectNumber, isHR)
			}
			break
		}
	}

	sort.Slice(beatMap.HitObjects, func(i, j int) bool {
		return beatMap.HitObjects[i].GetBasicData().StartTime < beatMap.HitObjects[j].GetBasicData().StartTime
	})

	num := 0

	for _, o := range beatMap.HitObjects {
		_, ok := o.(*objects.Pause)

		if !ok {
			o.GetBasicData().Number = int64(num)
			num++
		}

	}

	calculateStackLeniencywithMods(beatMap, isHR, isEZ)
}
