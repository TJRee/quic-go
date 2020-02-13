package quic

type sendQueue struct {
	queue     chan *packedPacket
	closeChan chan struct{} // closed when Close() is called
	closed    chan struct{} // closed when the run loop returns
	conn      connection
}

func newSendQueue(conn connection) *sendQueue {
	s := &sendQueue{
		conn:      conn,
		closed:    make(chan struct{}),
		closeChan: make(chan struct{}),
		queue:     make(chan *packedPacket, 1),
	}
	return s
}

func (h *sendQueue) Send(p *packedPacket) {
	h.queue <- p
}

func (h *sendQueue) Run() error {
	defer close(h.closed)
	var shouldClose bool
	for {
		if shouldClose && len(h.queue) == 0 {
			return nil
		}
		select {
		case <-h.closeChan:
			h.closeChan = nil // prevent this case from being selected again
			// make sure that all queued packets are actually sent out
			shouldClose = true
		case p := <-h.queue:
			if err := h.conn.Write(p.raw); err != nil {
				return err
			}
			p.buffer.Release()
		}
	}
}

func (h *sendQueue) Close() {
	close(h.closeChan)
	// wait until the run loop returned
	<-h.closed
}
