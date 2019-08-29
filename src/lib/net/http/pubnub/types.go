package pubnub

import (
	"net"
	"net/http"
	"sync"
)

type (
	Pubnub struct {
		origin       string
		publishKey   string
		subscribeKey string
		secretKey    string
		cipherKey    string
		//isSSL              bool
		uuid               string
		subscribedChannels string
		connectRetry       int
		subscribeTimeout   int64

		subscribeChannels map[string]chan []byte
		//subscribeErrorChannels map[string]chan []byte
		subscribeTransport http.RoundTripper
		subscribeConn      net.Conn

		// Global variable to reuse a commmon transport instance for non subscribe requests
		// Publish/HereNow/DetailedHitsory/Unsubscribe/UnsibscribePresence/Time.
		nonSubscribeTransport http.RoundTripper
		nonSubscribeConn      net.Conn

		client *http.Client

		presenceChannels map[string]chan []byte
		//presenceErrorChannels map[string]chan []byte

		newSubscribedChannels string
		connectTimeout        int64
		nonSubscribeTimeout   int64
		sync.Mutex
	}

	// Base response
	Response struct {
		Status  int    `json:"status"`
		Service string `json:"service"`
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	// Grant response
	GrantResponse struct {
		Response
		Payload struct {
			Subscribe_key string `json:"subscribe_key"`
			Level         string `json:"level"`
			Channel       string `json:"channel"`
			Auths         map[string]struct {
				R int `json:"r"`
				M int `json:"m"`
				W int `json:"w"`
				D int `json:"d"`
			} `json:"auths"`
			Ttl int `json:"ttl"`
		} `json:"payload,omitempty"`
	}

	// Audit response
	AuditResponse struct {
		Response
		Payload struct {
			Channels map[string]struct {
				Auths map[string]struct {
					R   int `json:"r"`
					M   int `json:"m"`
					W   int `json:"w"`
					Ttl int `json:"ttl"`
				} `json:"auths"`
			} `json:"channels"`
			Subscribe_key string `json:"subscribe_key"`
			Level         string `json:"level"`
		} `json:"payload,omitempty"`
	}
)
