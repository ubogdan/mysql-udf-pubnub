package main

type (
	pool struct {
		size int
		conn chan interface{}
	}
	poolInitFn func() (interface{}, error)
)

func (p *pool) InitPool(size int, initfn poolInitFn) error {
	// Create a buffered channel allowing size senders
	p.conn = make(chan interface{}, size)
	for x := 0; x < size; x++ {
		conn, err := initfn()
		if err != nil {
			return err
		}

		// If the init function succeeded, add the connection to the channel
		p.conn <- conn
	}
	p.size = size
	return nil
}

func (p *pool) GetConnection() interface{} {
	return <-p.conn
}

func (p *pool) ReleaseConnection(conn interface{}) {
	p.conn <- conn
}
