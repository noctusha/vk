package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Pool struct {
	ch         chan string
	numWorkers int
	mu         sync.Mutex
	workers    map[int]context.CancelFunc
}

var wg sync.WaitGroup

func (p *Pool) worker(id int, ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-p.ch:
			if !ok {
				fmt.Printf("worker %d: no more tasks, exiting\n", id)
				return
			}
			fmt.Printf("worker %d received task %s\n", id, task)
			time.Sleep(50 * time.Millisecond)
		case <-ctx.Done():
			fmt.Printf("worker %d: received stop signal, exiting\n", id)
			return
		}
	}
}

func (p *Pool) AddWorker(baseCtx context.Context) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx, cancel := context.WithCancel(baseCtx)
	if p.workers == nil {
		p.workers = make(map[int]context.CancelFunc)
	}

	wg.Add(1)
	p.numWorkers++
	workerID := p.numWorkers
	p.workers[workerID] = cancel

	go p.worker(workerID, ctx)
	fmt.Printf("Added worker %d, total workers: %d\n", workerID, p.numWorkers)

	return workerID
}

func (p *Pool) RemoveWorker(workerID int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	cancel, exists := p.workers[workerID]
	if exists {
		cancel()
		delete(p.workers, workerID)
		p.numWorkers--
		fmt.Printf("Requested removal of worker %d. Remaining workers: %d\n", workerID, p.numWorkers)
	} else {
		fmt.Printf("Worker %d does not exist\n", workerID)
	}
}

func main() {

	baseCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := Pool{
		ch: make(chan string, 4),
	}

	worker1 := pool.AddWorker(baseCtx)
	worker2 := pool.AddWorker(baseCtx)
	worker3 := pool.AddWorker(baseCtx)
	worker4 := pool.AddWorker(baseCtx)

	for i := 'A'; i < 'Z'; i++ {
		pool.ch <- string(i)
	}

	time.Sleep(200 * time.Millisecond)
	pool.AddWorker(baseCtx)

	time.Sleep(200 * time.Millisecond)
	pool.RemoveWorker(worker4)

	time.Sleep(200 * time.Millisecond)
	_ = pool.AddWorker(baseCtx)

	time.Sleep(200 * time.Millisecond)
	pool.RemoveWorker(worker1)
	pool.RemoveWorker(worker3)
	pool.RemoveWorker(worker2)

	cancel()
	wg.Wait()
	fmt.Println("All workers exited")
}
