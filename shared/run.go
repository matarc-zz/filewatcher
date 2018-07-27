package shared

import (
	"encoding/json"
	"os"

	"github.com/matarc/filewatcher/log"
)

type Runnable interface {
	Run() error
	Stop()
	Init()
}

func LoadConfig(cfgPath string, r Runnable) {
	file, err := os.Open(cfgPath)
	if err != nil {
		log.Errorf("Can't open '%s', using default configuration instead", cfgPath)
	} else {
		err = json.NewDecoder(file).Decode(r)
		if err != nil {
			log.Errorf("Can't decode '%s' as a json file, using default configuration instead", cfgPath)
		}
	}
	r.Init()
}
