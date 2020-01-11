package priority

type link chan interface{}

type PriorityQueue struct {
	main      <-chan interface{}
	ins       map[int]chan<- interface{}
	defaultIn chan<- interface{}
}

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
		pq.defaultIn = pq.ins[0]
	} else {
		queues[lastIndex].start(nil)
		pq.defaultIn = pq.ins[lastIndex]
	}
	queues = nil
	return pq
}

func (pq *PriorityQueue) Close() {
	close(pq.defaultIn)
}

func (pq *PriorityQueue) Write(priority int) chan<- interface{} {
	in, ok := pq.ins[priority]
	if !ok {
		in = pq.defaultIn
	}
	return in
}

func (pq *PriorityQueue) Read() <-chan interface{} {
	return pq.main
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
