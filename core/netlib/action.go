package netlib

import (
	"fmt"
	"reflect"

	"github.com/breezedup/goserver.v3/core/logger"
	"github.com/breezedup/goserver.v3/core/profile"
	"github.com/breezedup/goserver.v3/core/utils"
)

type action struct {
	s       *Session
	p       interface{}
	n       string
	packid  int
	logicNo uint32
	next    *action
}

func (this *action) do() {
	watch := profile.TimeStatisticMgr.WatchStart(fmt.Sprintf("/action/%v", this.n), profile.TIME_ELEMENT_ACTION)
	defer func() {
		FreeAction(this)
		if watch != nil {
			watch.Stop()
		}
		utils.DumpStackIfPanic(fmt.Sprintf("netlib.session.task.do exe error, packet type:%v", reflect.TypeOf(this.p)))
	}()

	h := GetHandler(this.packid)
	if h != nil {
		err := h.Process(this.s, this.packid, this.p)
		if err != nil {
			logger.Logger.Infof("%v process error %v", this.n, err)
		}
	} else {
		logger.Logger.Infof("%v not registe handler", this.n)
	}
}
