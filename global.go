package distributed_lock

import (
	"github.com/chuckchann/distributed-lock/entry"
	"io"
	"log"
)

//suggest set your project name as GlobalPrefix
func SetGlobalPrefix(p string) {
	entry.GlobalPrefix = p
}

func SetLoggerOutput(w io.Writer) {
	log.SetOutput(w)
}
