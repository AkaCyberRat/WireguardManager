package manager

type Result[T any] struct {
	Ok    T
	Error error
}

type ExecutableTask[TContext any] interface {
	Execute(ctx TContext)
}

type TaskWaiter[TOut any] interface {
	WaitAsync() <-chan TOut
	WaitSync() TOut
}

type GenericTask[TParams any, TOut any, TContext any] struct {
	params   TParams
	resultCh chan Result[TOut]
	execFunc func(TParams, chan<- Result[TOut], TContext)
}

func NewGenericTask[TParams any, TOut any, TContext any](params TParams, execFunc func(TParams, chan<- Result[TOut], TContext)) GenericTask[TParams, TOut, TContext] {
	resultCh := make(chan Result[TOut], 1)

	return GenericTask[TParams, TOut, TContext]{params: params, resultCh: resultCh, execFunc: execFunc}
}

func (t *GenericTask[TParams, TOut, TContext]) execute(ctx TContext) {
	defer close(t.resultCh)

	t.execFunc(t.params, t.resultCh, ctx)
}

func (t *GenericTask[TParams, TOut, TContext]) WaitAsync() <-chan Result[TOut] {
	return t.resultCh
}

func (t *GenericTask[TParams, TOut, TContext]) WaitSync() Result[TOut] {
	return <-t.resultCh
}
