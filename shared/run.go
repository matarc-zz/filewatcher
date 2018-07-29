package shared

import (
	"encoding/json"
	"os"

	"github.com/matarc/filewatcher/log"
)

// Runnable is an interface that represents either a server or a client.
type Runnable interface {
	Run() error
	Stop()
	Init()
}

// LoadConfig loads the json file `cfgPath` and stores it in `r`.
// It initialises `r` after loading it by calling `Init`
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
