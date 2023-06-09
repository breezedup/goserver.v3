package task

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"fmt"
	"github.com/breezedup/goserver.v3/core"
	"github.com/breezedup/goserver.v3/core/basic"
	"github.com/breezedup/goserver.v3/core/container"
	"github.com/breezedup/goserver.v3/core/container/recycler"
	"github.com/breezedup/goserver.v3/core/logger"
	"github.com/breezedup/goserver.v3/core/profile"
)

var taskMutexLock sync.Mutex
var taskMutexPool = make(map[string]*Task)
var ErrTaskIsRunning = errors.New("mutex task is running")

type Callable interface {
	Call(*basic.Object) interface{}
}

type CompleteNotify interface {
	Done(interface{}, *Task)
}

type CallableWrapper func(o *basic.Object) interface{}

func (cw CallableWrapper) Call(o *basic.Object) interface{} {
	return cw(o)
}

type CompleteNotifyWrapper func(interface{}, *Task)

func (cnw CompleteNotifyWrapper) Done(i interface{}, t *Task) {
	cnw(i, t)
}

type Task struct {
	s            *basic.Object
	c            Callable
	n            CompleteNotify
	r            chan interface{}
	env          *container.SynchronizedMap
	tCreate      time.Time
	tStart       time.Time
	alertTime    time.Duration
	name         string
	refTaskCnt   int32
	beforeQueCnt int    //入队列前，等待中的任务数量
	afterQueCnt  int    //出队列后，等待中的任务数量
	running      int32  //是否正在运行
	mutexFlag    bool   //是否互斥任务
	mutexKey     string //互斥任务key
}

func New(s *basic.Object, c Callable, n CompleteNotify, name ...string) *Task {
	t := &Task{
		s:       s,
		c:       c,
		n:       n,
		r:       make(chan interface{}, 1),
		tCreate: time.Now(),
	}
	if len(name) != 0 {
		t.name = name[0]
	}
	if s == nil {
		t.s = core.CoreObject()
	}
	return t
}

func NewMutexTask(s *basic.Object, c Callable, n CompleteNotify, key, name string) (t *Task, done bool) {
	mutexKey := name + key
	taskMutexLock.Lock()
	if t, ok := taskMutexPool[mutexKey]; ok {
		taskMutexLock.Unlock()
		return t, true
	}

	t = New(s, c, n, name)
	t.mutexKey = mutexKey
	t.mutexFlag = true
	taskMutexPool[mutexKey] = t
	taskMutexLock.Unlock()
	return t, false
}

func (t *Task) AddRefCnt(cnt int32) int32 {
	return atomic.AddInt32(&t.refTaskCnt, cnt)
}

func (t *Task) GetRefCnt() int32 {
	return atomic.LoadInt32(&t.refTaskCnt)
}

func (t *Task) Get() interface{} {
	if t.n != nil {
		panic("Task result by CompleteNotify return")
	}

	return <-t.r
}

func (t *Task) GetWithTimeout(timeout time.Duration) interface{} {
	if timeout == 0 {
		return t.Get()
	} else {
		timer := recycler.GetTimer(timeout)
		defer recycler.GiveTimer(timer)
		select {
		case r, ok := <-t.r:
			if ok {
				return r
			} else {
				return nil
			}
		case <-timer.C:
			return nil
		}
	}
	return nil
}

func (t *Task) GetEnv(k interface{}) interface{} {
	if t.env == nil {
		return nil
	}
	return t.env.Get(k)
}

func (t *Task) PutEnv(k, v interface{}) bool {
	if t.env == nil {
		t.env = container.NewSynchronizedMap()
	}
	if t.env != nil {
		t.env.Set(k, v)
	}

	return true
}

func (t *Task) run(o *basic.Object) (e error) {
	if t.mutexFlag {
		// process mutex task
		if !atomic.CompareAndSwapInt32(&t.running, 0, 1) {
			return ErrTaskIsRunning
		}
	}

	watch := profile.TimeStatisticMgr.WatchStart(fmt.Sprintf("/task/%v/run", t.name), profile.TIME_ELEMENT_TASK)
	defer func() {
		if watch != nil {
			watch.Stop()
		}

		if t.mutexFlag {
			taskMutexLock.Lock()
			delete(taskMutexPool, t.mutexKey)
			taskMutexLock.Unlock()
		}

		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			logger.Logger.Error("Task::run stack--->", string(buf[:n]))
		}
	}()

	t.tStart = time.Now()
	wait := t.tStart.Sub(t.tCreate)
	ret := t.c.Call(o)
	dura := t.GetRunTime()

	if t.r != nil {
		t.r <- ret
	}

	if t.n != nil {
		SendTaskRes(t.s, t)
	}
	if t.alertTime != 0 && t.name != "" {
		cost := t.GetCostTime()
		if cost > t.alertTime {
			logger.Logger.Warn("task [", t.name, "] since createTime(",
				cost, ") since startTime(", dura, "), in quene wait(", wait, ")", " beforeQueCnt(", t.beforeQueCnt, ") afterQueCnt(", t.afterQueCnt, ")")
		}
	}
	return nil
}

func (t *Task) Start() {
	go t.run(nil)
}

func (t *Task) SetAlertTime(alertt time.Duration) {
	t.alertTime = alertt
}

func (t *Task) GetCostTime() time.Duration {
	return time.Now().Sub(t.tCreate)
}

func (t *Task) GetRunTime() time.Duration {
	return time.Now().Sub(t.tStart)
}

func (t *Task) StartByExecutor(name string) bool {
	return sendTaskReqToExecutor(t, name, "")
}

func (t *Task) StartByFixExecutor(name string) bool {
	return sendTaskReqToFixExecutor(t, name, "")
}

func (t *Task) BroadcastToAllExecutor() bool {
	return sendTaskReqToAllExecutor(t)
}

func (t *Task) StartByGroupExecutor(gname string, name string) bool {
	return sendTaskReqToExecutor(t, name, gname)
}

func (t *Task) StartByGroupFixExecutor(name, gname string) bool {
	return sendTaskReqToFixExecutor(t, name, gname)
}
