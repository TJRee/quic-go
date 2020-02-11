package quic

type sendQueue struct {
	queue     chan *packetBuffer
	closeChan chan struct{}
	conn      connection
}

func newSendQueue(conn connection) *sendQueue {
	s := &sendQueue{
		conn:      conn,
		closeChan: make(chan struct{}),
		queue:     make(chan *packetBuffer, 1),
	}
	return s
}

func (h *sendQueue) Send(p *packetBuffer) {
	h.queue <- p
}

func (h *sendQueue) Run() error {
	var p *packetBuffer
	for {
		select {
		case <-h.closeChan:
			return nil
		case p = <-h.queue:
		}
		if err := h.conn.Write(p.Data); err != nil {
			return err
		}
		p.Release()
	}
}

func (h *sendQueue) Close() {
	close(h.closeChan)
}
