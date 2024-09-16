package pool

import (
	"sync"
)

// Result represents the result of a task with its ID, output, and error.
type Result[O any] struct {
	id     int
	Output O
	Err    error
}

type Task[O any] func() (O, error)

// OrderedGo executes tasks concurrently and returns results in the same order as provided.
func OrderedGo[O any](tasks []Task[O]) []Result[O] {
	ch := make(chan Result[O], len(tasks))
	wg := &sync.WaitGroup{}

	for i, f := range tasks {
		wg.Add(1)
		go func(id int, t func() (O, error)) {
			defer wg.Done()
			data, err := t()
			ch <- Result[O]{id, data, err}
		}(i, f)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	res := make([]Result[O], len(tasks))
	for r := range ch {
		res[r.id] = r
	}

	return res
}

// OrderedGoFunc executes functions concurrently and returns results in the same order as provided.
// Unlike [OrderedGo], it takes a slice of input and a function to apply on each input.
func OrderedGoFunc[I, O any](s []I, f func(I) (O, error)) []Result[O] {
	tasks := make([]Task[O], 0, len(s))
	for _, v := range s {
		tasks = append(tasks, func() (O, error) {
			return f(v)
		})
	}

	return OrderedGo(tasks)
}

// OrderedRun is a convenience function that run [OrderedGoFunc] and return the output slice directly instead of a slice of [Result]. If any task returns an error, it will return the error immediately.
func OrderedRun[I, O any](s []I, f func(I) (O, error)) ([]O, error) {
	results := OrderedGoFunc(s, f)
	outputs := make([]O, 0, len(s))
	for _, r := range results {
		if r.Err != nil {
			return nil, r.Err
		}
		outputs = append(outputs, r.Output)
	}

	return outputs, nil
}
