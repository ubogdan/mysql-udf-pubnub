package pubnub

import (
	"sync"
	"testing"
)

var pubnub = Pubnub{
	origin:              "https://pubsub.pubnub.com",
	publishKey:          "pub-c-82403925-3c5f-466f-a67b-8faef7d9492d",
	subscribeKey:        "sub-c-57712b5e-05ae-11e4-aac6-02ee2ddab7fe",
	secretKey:           "sec-c-NjM3NDgyYjEtYjY2Zi00N2I5LThmMWEtN2I2MjFkZTY1ZTY2",
	connectRetry:        3,
	connectTimeout:      10,
	subscribeTimeout:    310,
	nonSubscribeTimeout: 5,
}

func TestGrant(t *testing.T) {

	request := func(wg *sync.WaitGroup, t *testing.T) {
		defer wg.Done()
		response, err := pubnub.Grant("console", "auth_AAA", true, true, 1700)
		if err != nil {
			t.Errorf("Grant %s", err)
		}
		t.Logf("res %v", response)
	}

	wg := &sync.WaitGroup{}
	for i := 1; i < 10; i++ {
		wg.Add(1)
		go request(wg, t)
	}
	wg.Wait()
	t.Logf("Done")
}

func zTestAst(t *testing.T) {
	response, err := pubnub.SubKey("ttasdsad", "apns", "b5f5bd50-d8c9-4773-973f-a9c710e9dba9")
	if err != nil {
		t.Errorf("SubKey %s", err)
		return
	}
	t.Logf("res %s", response)
}
