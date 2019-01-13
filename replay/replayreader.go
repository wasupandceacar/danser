package replay

import (
	"danser/settings"
	"io/ioutil"
	"strings"
)

func GetOsrFiles() (files []string, err error) {
	dir, err := ioutil.ReadDir(settings.General.ReplayDir)
	if err != nil {
		return nil, err
	}

	for _, fi := range dir {
		ok := strings.HasSuffix(fi.Name(), ".osr")
		if ok {
			files = append(files, settings.General.ReplayDir+fi.Name())
		}
	}

	return files,nil
}
