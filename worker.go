package main

// Go imports
import (
	"container/list"
	"log"
	"strings"
	"sync"
	"time"

	"lib/net/http/pubnub"
)

type worker struct {
	connPool *pool      // Pool of PubnubAgents
	queue    *list.List // Messages to Publish
	qlock    sync.Mutex // Publish lock
}

func init() {
	w = &worker{
		connPool: &pool{},
		queue:    list.New(),
	}

	// Initialize Pubnub Agent pool
	w.connPool.InitPool(30,
		func() (interface{}, error) {
			return pubnub.New(pubKey, subKey, secKey, "", true, ""), nil
		},
	)

	// Start Worker
	go func(w *worker) {
		for {
			select {
			case <-time.After(200 * time.Millisecond):
				w.qlock.Lock()
				for i := 0; i < w.queue.Len(); i++ {
					message := w.queue.Back()
					w.queue.Remove(message)
					// Async send ussing AgentPool
					go w.deliver(message)
				}
				w.qlock.Unlock()
			}
		}
	}(w)

}

func (w *worker) Publish(channel string, message []byte, flags string) {
	w.qlock.Lock()
	defer w.qlock.Unlock()

	history := strings.Contains(flags, "h")

	w.queue.PushFront(
		&publishMessage{
			Channel: channel,
			Store:   history,
			Message: message,
		},
	)
}

func (worker *worker) Grant(channel, auth string, rights string, ttl int) {
	worker.qlock.Lock()
	defer worker.qlock.Unlock()

	read := strings.Contains(rights, "r")
	write := strings.Contains(rights, "w")
	//manage := strings.Contains(flags, "m")

	worker.queue.PushFront(
		&grantMessage{
			Channel: channel,
			Auth:    auth,
			Read:    read,
			Write:   write,
			Ttl:     ttl,
		},
	)
}

func (w *worker) deliver(message *list.Element) {

	agent := w.connPool.GetConnection().(*pubnub.Pubnub)
	defer w.connPool.ReleaseConnection(agent)

	// Grant Message
	if g, ok := message.Value.(*grantMessage); ok {

	punubGrant:
		_, err := agent.Grant(g.Channel, g.Auth, g.Read, g.Write, g.Ttl)
		if err != nil {
			log.Printf("Grant for %s failed %s !", g.Channel, err)
			// Give a rest to PubNub for retry
			time.Sleep(100 * time.Millisecond)
			goto punubGrant
		}
	}

	// Publish Message
	if publish, ok := message.Value.(*publishMessage); ok {
	pubnubPublish:

		response, err := agent.Publish(publish.Channel, string(publish.Message), "", publish.Store)
		if err != nil {
			log.Printf("Publish for %s failed %s !", publish.Channel, err)
			// Give a rest to PubNub for retry
			time.Sleep(100 * time.Millisecond)
			goto pubnubPublish
		}

		if response.Status == 403 {
			log.Printf("Publish %d %q: %s", response.Status, response.Message, publish.Message)
		}

	}

}
