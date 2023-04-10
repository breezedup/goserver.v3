package task

import (
	"github.com/breezedup/goserver.v3/core/basic"
	"github.com/breezedup/goserver.v3/core/utils"
)

type taskExeCommand struct {
	t *Task
}

func (ttc *taskExeCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	defer utils.DumpStackIfPanic("taskExeCommand")
	ttc.t.afterQueCnt = o.GetPendingCommandCnt()
	return ttc.t.run(o)
}

func SendTaskExe(o *basic.Object, t *Task) bool {
	t.beforeQueCnt = o.GetPendingCommandCnt()
	return o.SendCommand(&taskExeCommand{t: t}, true)
}
