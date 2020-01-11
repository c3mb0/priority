package priority

type queue struct {
	out    chan<- interface{}
	in     <-chan interface{}
	feed   <-chan interface{}
	closer func()
}

func newQueue(out chan<- interface{}, in <-chan interface{}, closer func()) *queue {
	return &queue{
		out:    out,
		in:     in,
		closer: closer,
	}
}

func (q *queue) start(feed <-chan interface{}) {
	if feed == nil {
		q.closer = nil
	}
	q.feed = feed
	go q.manage()
}

func (q *queue) manage() {
	for {
		select {
		case val, open := <-q.in:
			if !open && q.feed == nil {
				close(q.out)
				return
			}
			q.out <- val
		default:
			select {
			case val, open := <-q.in:
				if !open && q.feed == nil {
					close(q.out)
					return
				}
				q.out <- val
			case val, open := <-q.feed:
				if !open {
					q.feed = nil
					q.closer()
					continue
				}
				q.out <- val
			}
		}
	}
}
