package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/breezedup/goserver.v3/core/module"
	"github.com/breezedup/goserver.v3/core/task"
)

var TaskExampleSington = &TaskExample{}

type TaskExample struct {
	id int
}

// in task.Worker goroutine
func (this *TaskExample) Call() interface{} {
	tNow := time.Now()
	fmt.Println("[", this.id, "]TaskExample execute start ")
	time.Sleep(time.Second * time.Duration(rand.Intn(10)))
	fmt.Println("[", this.id, "]TaskExample execute end, take ", time.Now().Sub(tNow))
	return nil
}

// in laucher goroutine
func (this *TaskExample) Done(i interface{}, t *task.Task) {
	fmt.Println("TaskExample execute over")
}

// //////////////////////////////////////////////////////////////////
// / Module Implement [beg]
// //////////////////////////////////////////////////////////////////
func (this *TaskExample) ModuleName() string {
	return "taskexample"
}

func (this *TaskExample) Init() {
	for i := 1; i < 100; i++ {
		th := &TaskExample{id: i}
		t := task.New(nil, th, th)
		if b := t.StartByExecutor(fmt.Sprintf("%v", i)); !b {
			fmt.Println("[", i, "]task lauch failed")
		} else {
			fmt.Println("[", i, "]task lauch success")
		}
	}

	for i := 100; i < 200; i++ {
		th := &TaskExample{id: i}
		t := task.New(nil, th, th)
		w := rand.Intn(100)
		go func(id, n int) {
			if b := t.StartByFixExecutor(fmt.Sprintf("test%v", n)); !b {
				fmt.Println("[", id, "]task lauch failed")
			} else {
				fmt.Println("[", id, "]task lauch success")
			}
		}(i, w)
	}
}

func (this *TaskExample) Update() {
	fmt.Println("TaskExample.Update")
}

func (this *TaskExample) Shutdown() {
	module.UnregisteModule(this)
}

////////////////////////////////////////////////////////////////////
/// Module Implement [end]
////////////////////////////////////////////////////////////////////

func init() {
	module.RegisteModule(TaskExampleSington, time.Second, 0)
}
