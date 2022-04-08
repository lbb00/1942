package tasks

import "context"

type Task struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewTask(ctx context.Context) *Task {
	task := &Task{}
	ctx2, cancel := context.WithCancel(ctx)
	task.Ctx = ctx2
	task.Cancel = cancel

	return task
}
