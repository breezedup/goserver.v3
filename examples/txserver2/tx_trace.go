package main

import (
	"github.com/breezedup/goserver.v3/core"
	"github.com/breezedup/goserver.v3/core/logger"
	"github.com/breezedup/goserver.v3/core/netlib"
	"github.com/breezedup/goserver.v3/core/transact"
	"github.com/breezedup/goserver.v3/examples/protocol"
	"github.com/breezedup/goserver.v3/srvlib"
)

type traceTransHandler struct {
}

func init() {
	transact.RegisteHandler(protocol.TxTrace, &traceTransHandler{})
	srvlib.ServerSessionMgrSington.AddListener(&MyServerSessionRegisteListener{})
}

func (this *traceTransHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnExcute ")
	tnp := &transact.TransNodeParam{
		Tt:     protocol.TxTrace,
		Ot:     transact.TransOwnerType(2),
		Oid:    201,
		AreaID: 1,
		Tct:    transact.TransactCommitPolicy_TwoPhase,
	}
	p := new(int)
	*p = -2
	userData := protocol.StructA{X: 10, Y: -1, Z: 65535, P: p, Desc: "welcome!"}
	tNode.StartChildTrans(tnp, userData, transact.DefaultTransactTimeout)
	return transact.TransExeResult_Success
}

func (this *traceTransHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *traceTransHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *traceTransHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

type MyServerSessionRegisteListener struct {
}

func (mssrl *MyServerSessionRegisteListener) OnRegiste(*netlib.Session) {
	logger.Logger.Trace("MyServerSessionRegisteListener.OnRegiste")
	tnp := &transact.TransNodeParam{
		Tt:     protocol.TxTrace,
		Ot:     transact.TransOwnerType(2),
		Oid:    202,
		AreaID: 1,
	}

	tNode := transact.DTCModule.StartTrans(tnp, nil, transact.DefaultTransactTimeout)
	if tNode != nil {
		tNode.Go(core.CoreObject())
	}
}

func (mssrl *MyServerSessionRegisteListener) OnUnregiste(*netlib.Session) {
	logger.Logger.Trace("MyServerSessionRegisteListener.OnUnregiste")
}
