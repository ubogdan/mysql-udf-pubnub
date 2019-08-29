package pubnub

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func zTestSendRequest(t *testing.T) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)

	}
	t.Logf("%s", l.Addr().String())

	lib := Pubnub{
		origin:              fmt.Sprintf("http://%s/", l.Addr().String()),
		connectRetry:        3,
		connectTimeout:      10,
		subscribeTimeout:    310,
		nonSubscribeTimeout: 5,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "\n")
		time.Sleep(time.Duration(lib.nonSubscribeTimeout) * time.Second)
		fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])

	})

	go http.Serve(l, nil)

	buff, code, err := lib.httpRequest("user", false)
	if err != nil {
		t.Fatalf("httpRequest %s", err)
		return
	}
	if code != 200 {
		t.Logf("Return code %d", code)
	}
	t.Logf("%s ", buff)
}
