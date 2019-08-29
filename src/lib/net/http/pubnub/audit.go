package pubnub

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"log"
)

func (pub *Pubnub) Audit(channel string, authkey string) (*AuditResponse, error) {

	params := ""

	authkey = strings.TrimSpace(authkey)
	if authkey != "" {
		params += "auth=" + authkey + "&"
	}

	channel = strings.TrimSpace(channel)
	if channel != "" {
		params += "channel=" + url.QueryEscape(channel) + "&"
	}

	params += "timestamp=" + fmt.Sprintf("%d", time.Now().Unix())

	// Sign request
	signature := getHmacSha256(
		pub.secretKey,
		pub.subscribeKey+"\n"+pub.publishKey+"\naudit\n"+params,
	)

	request := "/v1/auth/audit/sub-key/" + pub.subscribeKey + "?" + params + "&signature=" + signature
	value, _, err := pub.httpRequest(request, false)

	if err != nil {
		return nil, err
	}
	log.Printf("Audit %s", value)
	var response *AuditResponse
	err = json.Unmarshal([]byte(value), &response)
	if err != nil {
		return nil, err
	}

	return response, nil

}

//
func (a *AuditResponse) GetMaxTTL(channel string) int {
	maxTTL := 0
	_, found := a.Payload.Channels[channel]
	if found {
		for _, auth := range a.Payload.Channels[channel].Auths {
			if auth.Ttl > maxTTL {
				maxTTL = auth.Ttl
			}
		}
	}
	return maxTTL
}
