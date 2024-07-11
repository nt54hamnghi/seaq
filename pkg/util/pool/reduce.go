package pool

import (
	"math"
	"runtime"
	"sync"
)

type Job[O any] struct {
	Id     int
	Output O
}

// GetThreadCount calculates the number of threads for processing a given number of tasks.
// If taskCount is less than or equal to 0, it returns 1.
// If taskCount is less than the number of CPUs, it returns taskCount.
// Otherwise, it returns the number of CPUs.
func GetThreadCount(taskCount int) int {
	if taskCount <= 0 {
		return 1
	}
	numCpu := runtime.NumCPU()
	return int(math.Min(float64(taskCount), float64(numCpu)))
}

// BatchReduce concurrently processes input data in batches.
// It divides the input slice into smaller batches,
// processes each batch in parallel using the specified operation, and then collects the results.
//
// Parameters:
//   - nThreads: The number of concurrent threads to use for processing.
//   - in: The input slice containing data of any type to be processed.
//   - op: A function that defines the operation to be applied to each batch of the input slice.
//     This function must take a slice of the input type and return a single value of the output type.
//
// Returns:
//   - A slice containing the results of applying the operation to each batch of the input slice.
//     The order of results in the output slice corresponds to the order of the batches processed.
func BatchReduce[I, O any](
	nThreads int, // number of threads
	in []I, // input slice
	op func([]I) O, // operation to apply to each batch
) []O {
	if len(in) == 0 {
		return []O{}
	}

	ch := make(chan Job[O], nThreads)
	wg := &sync.WaitGroup{}

	splitIntoBatches(nThreads, in, op, ch, wg)

	go func() {
		wg.Wait()
		close(ch)
	}()

	res := make([]O, nThreads)
	for r := range ch {
		res[r.Id] = r.Output
	}
	return res
}

// inspired by: https://stackoverflow.com/a/35179941
func splitIntoBatches[I, O any](
	nThreads int, // number of threads
	in []I, // input slice
	op func([]I) O, // operation to apply to each chunk
	ch chan<- Job[O], // channel to collect results
	wg *sync.WaitGroup,
) {

	// Adds nThreads - 1 to len(in)
	// to ensure that any remainder from the division results in an extra batch being created,
	// effectively rounding up the division to handle all elements.
	batchSize := (len(in) + nThreads - 1) / nThreads

	for i := 0; i < len(in); i += batchSize {
		wg.Add(1)
		end := i + batchSize
		if end > len(in) {
			end = len(in)
		}

		go func(id int, s []I) {
			defer wg.Done()

			ch <- Job[O]{
				Id:     id / batchSize,
				Output: op(s),
			}

		}(i, in[i:end])
	}
}
