package transact

import (
	"github.com/breezedup/goserver.v3/core/basic"
)

type transactYieldCommand struct {
	tnode *TransNode
}

func (trc *transactYieldCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	trc.tnode.checkExeOver()
	return nil
}

func SendTranscatYield(tnode *TransNode) bool {
	return tnode.ownerObj.SendCommand(&transactYieldCommand{tnode: tnode}, true)
}
