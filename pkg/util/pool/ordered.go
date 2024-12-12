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

// Task represents a function that takes no arguments
// and returns a value and an error.
type Task[O any] func() (O, error)

// OrderedGo executes tasks concurrently and returns results in order.
// Each result maintains the same position as its corresponding task.
func OrderedGo[O any](tasks []Task[O]) []Result[O] {
	if len(tasks) == 0 {
		return nil
	}

	ch := make(chan Result[O], len(tasks))
	wg := &sync.WaitGroup{}

	for i, t := range tasks {
		wg.Add(1)
		go func(id int, t Task[O]) {
			defer wg.Done()
			output, err := t()
			ch <- Result[O]{id, output, err}
		}(i, t)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make([]Result[O], len(tasks))
	for r := range ch {
		results[r.id] = r
	}

	return results
}

// OrderedGoFunc applies a function to each input concurrently and returns
// a slice of Result in order.
func OrderedGoFunc[I, O any](items []I, fn func(I) (O, error)) []Result[O] {
	if len(items) == 0 {
		return nil
	}

	tasks := make([]Task[O], len(items))
	for i, v := range items {
		tasks[i] = func() (O, error) {
			return fn(v)
		}
	}

	return OrderedGo(tasks)
}

// OrderedRun applies a function to each input concurrently and returns a slice of output in order,
// stopping at the first error encountered.
func OrderedRun[I, O any](items []I, fn func(I) (O, error)) ([]O, error) {
	if len(items) == 0 {
		return nil, nil
	}

	results := OrderedGoFunc(items, fn)
	outputs := make([]O, len(items))
	for i, r := range results {
		if r.Err != nil {
			return nil, r.Err
		}
		outputs[i] = r.Output
	}
	return outputs, nil
}
