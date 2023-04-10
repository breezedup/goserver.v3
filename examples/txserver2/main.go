// main
package main

import (
	"github.com/breezedup/goserver.v3/core"
	"github.com/breezedup/goserver.v3/core/module"
)

func main() {
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	waiter := module.Start()
	waiter.Wait()
}
