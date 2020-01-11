package priority

import (
	qq "github.com/eapache/queue"
)

type inf struct {
	input  chan interface{}
	output chan interface{}
	q      *qq.Queue
}

func newInf() (chan<- interface{}, <-chan interface{}) {
	i := &inf{
		input:  make(chan interface{}),
		output: make(chan interface{}),
		q:      qq.New(),
	}
	go i.start()
	return i.input, i.output
}

func (i *inf) start() {
	var next interface{}
	var out chan interface{}
	in := i.input
	for in != nil || out != nil {
		select {
		case val, open := <-in:
			if open {
				i.q.Add(val)
			} else {
				in = nil
			}
		case out <- next:
			i.q.Remove()
		}
		if i.q.Length() > 0 {
			out = i.output
			next = i.q.Peek()
		} else {
			out = nil
			next = nil
		}
	}
	close(i.output)
}
