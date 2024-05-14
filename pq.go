package priority

type PriorityQueue struct {
	lowest, highest int
	main            <-chan interface{}
	ins             map[int]chan<- interface{}
	closed          bool
}

const (
	_ = -iota
	Lowest
	Highest
)

func NewPriorityQueue(numQueues int, ascendingPriority, blocking bool) *PriorityQueue {
	main := make(chan interface{})
	pq := &PriorityQueue{
		ins:  make(map[int]chan<- interface{}, numQueues),
		main: main,
	}
	queues := make([]*queue, numQueues)
	lastIndex := numQueues - 1
	for i := range queues {
		j := i
		if ascendingPriority {
			j = lastIndex - i
		}
		in, out, closer := generate(blocking)
		pq.ins[j] = in
		if i == 0 {
			queues[j] = newQueue(main, out, closer)
			continue
		}
		link := make(chan interface{})
		queues[j] = newQueue(link, out, closer)
		if ascendingPriority {
			queues[j+1].start(link)
		} else {
			queues[j-1].start(link)
		}
	}
	if ascendingPriority {
		queues[0].start(nil)
		pq.lowest = 0
		pq.highest = lastIndex
	} else {
		queues[lastIndex].start(nil)
		pq.lowest = lastIndex
		pq.highest = 0
	}
	queues = nil
	return pq
}

func (pq *PriorityQueue) Close() {
	pq.closed = true
	close(pq.ins[pq.lowest])
}

func (pq *PriorityQueue) Write(priority int) chan<- interface{} {
	switch priority {
	case Lowest:
		priority = pq.lowest
	case Highest:
		priority = pq.highest
	}
	in, ok := pq.ins[priority]
	if !ok {
		in = pq.ins[pq.lowest]
	}
	return in
}

func (pq *PriorityQueue) WriteValue(priority int, msg interface{}) bool {
	if pq.closed {
		return false
	}
	pq.Write(priority) <- msg
	return true
}

func (pq *PriorityQueue) ReadChan() <-chan interface{} {
	return pq.main
}

func (pq *PriorityQueue) ReadValue() (interface{}, bool) {
	val, ok := <-pq.main
	return val, ok
}

func generate(blocking bool) (in chan<- interface{}, out <-chan interface{}, closer func()) {
	if blocking {
		in, out = newChan()
	} else {
		in, out = newInf()
	}
	closer = func() { close(in) }
	return
}

func newChan() (chan<- interface{}, <-chan interface{}) {
	c := make(chan interface{})
	return c, c
}
