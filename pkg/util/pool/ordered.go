package pool

import "sync"

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
