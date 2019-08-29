package pubnub

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	origin = "ps.pndsn.com"

	//Sdk Identification Param appended to each request
	sdkIdentificationParamKey = "pnsdk"
	sdkIdentificationParamVal = "PubNub-Go/3.16.1"
)

var sdkIdentificationParam = fmt.Sprintf("%s=%s", sdkIdentificationParamKey, url.QueryEscape(sdkIdentificationParamVal))

// NewPubnub initializes pubnub struct with the user provided values.
// And then initiates the origin by appending the protocol based upon the sslOn argument.
// Then it uses the customuuid or generates the uuid.
//
// It accepts the following parameters:
// publishKey is the user specific Publish Key. Mandatory.
// subscribeKey is the user specific Subscribe Key. Mandatory.
// secretKey is the user specific Secret Key. Accepts empty string if not used.
// cipherKey stores the user specific Cipher Key. Accepts empty string if not used.
// sslOn is true if enabled, else is false.
// customUuid is the unique identifier, it can be a custom value or sent as empty for automatic generation.
//
// returns the pointer to Pubnub instance.
func New(publishKey string, subscribeKey string, secretKey string, cipherKey string, sslOn bool, customUuid string) *Pubnub {

	pubnub := &Pubnub{
		origin:                "http://" + origin,
		publishKey:            publishKey,
		subscribeKey:          subscribeKey,
		secretKey:             secretKey,
		cipherKey:             cipherKey,
		uuid:                  "",
		subscribedChannels:    "",
		newSubscribedChannels: "",
		connectRetry:          3,
		connectTimeout:        10,
		nonSubscribeTimeout:   5,
		presenceChannels:      make(map[string]chan []byte),
		subscribeChannels:     make(map[string]chan []byte),
	}

	if sslOn {
		pubnub.origin = "https://" + origin
	}

	return pubnub
}

// GetClient Get a client for transactional requests
func (pub *Pubnub) GetClient() *http.Client {
	pub.Lock()
	defer pub.Unlock()

	if pub.client == nil {
		transport := &http.Transport{
			// MaxIdleConns: 30,
			Dial: (&net.Dialer{
				// Covers establishing a new TCP connection
				Timeout: time.Duration(pub.connectTimeout) * time.Second,
			}).Dial,
		}

		client := &http.Client{
			Transport: transport,
			// Covers the entire exchange from Dial to reading the body
			Timeout: time.Duration(pub.nonSubscribeTimeout) * time.Second,
		}
		pub.client = client
	}

	return pub.client
}

func (pub *Pubnub) Publish(channel string, message string, auth string, storeInHistory bool) (*Response, error) {
	return pub.sendPublish(channel, message, auth, storeInHistory, true, -1)
}

func (pub *Pubnub) sendPublish(channel string, message string, auth string, storeInHistory, replicate bool, ttl int) (*Response, error) {

	signature := "0"
	if pub.secretKey != "" {
		signature = getHmacSha256(pub.secretKey, fmt.Sprintf("%s/%s/%s/%s/%s", pub.publishKey, pub.subscribeKey, pub.secretKey, channel, message))
	}

	publishURL := fmt.Sprintf("/publish/%s/%s/%s/%s/0/%s",
		pub.publishKey, pub.subscribeKey, signature,
		url.QueryEscape(channel),
		encodeJSONAsPathComponent(message))
	requestURL := publishURL

	// Sdk
	publishURL += "?" + sdkIdentificationParam

	// Send auth-key
	if auth != "" {
		publishURL = fmt.Sprintf("%s&auth=%s", publishURL, auth)
	}

	// Skip history
	if storeInHistory == false {
		publishURL = fmt.Sprintf("%s&store=0", publishURL)
	}

	if !replicate {
		publishURL = fmt.Sprintf("%s&norep=true", publishURL)
	}

	if ttl >= 0 {
		publishURL = fmt.Sprintf("%s&ttl=%d", publishURL, ttl)
	}

	publishURL = pub.checkSecretKeyAndAddSignature(publishURL, requestURL)

	value, responseCode, err := pub.httpRequest(publishURL, false)
	if err != nil {
		return nil, fmt.Errorf("Publish Error Internal: %s", err)
	}

	// Response code
	if responseCode != 200 {
		var response *Response
		if e := json.Unmarshal([]byte(value), &response); e != nil {
			return &Response{
				Status:  400,
				Error:   false,
				Message: e.Error(),
			}, nil
		}
		return response, nil
	}

	return &Response{
		Status:  200,
		Error:   false,
		Message: string(value),
	}, nil
}

// Grant auth access rights
func (pub *Pubnub) Grant(channel string, auth string, read_perm bool, write_perm bool, ttl int) (*GrantResponse, error) {
	return pub._auth(channel, auth, read_perm, write_perm, ttl)
}

// Revoke auth access rights
func (pub *Pubnub) Revoke(channel string, auth string, ttl int) (*GrantResponse, error) {
	return pub._auth(channel, auth, false, false, ttl)
}

// Pubnub's auth call
func (pub *Pubnub) _auth(channel string, auth string, read_perm bool, write_perm bool, ttl int) (*GrantResponse, error) {

	read_str := "0"
	if read_perm {
		read_str = "1"
	}
	write_str := "0"
	if write_perm {
		write_str = "1"
	}
	params := ""
	if auth != "" {
		params = "auth=" + auth
		params += "&channel=" + channel
	} else {
		params += "channel=" + channel
	}
	params += "&r=" + read_str
	params += "&timestamp=" + fmt.Sprintf("%d", time.Now().Unix())

	if ttl > -1 {
		params += "&ttl=" + strconv.Itoa(ttl)
	}
	params += "&w=" + write_str

	// Sign request
	signature := getHmacSha256(
		pub.secretKey,
		pub.subscribeKey+"\n"+pub.publishKey+"\ngrant\n"+params,
	)

	params += "&signature=" + signature

	value, responseCode, err := pub.httpRequest("/v1/auth/grant/sub-key/"+pub.subscribeKey+"?"+params, false)
	if (responseCode != 200) || (err != nil) {
		if err != nil {
			return nil, fmt.Errorf("PAM Error Internal: %s", err)
		}
		return nil, fmt.Errorf("PAM Error Message: %s", value)
	}

	var response *GrantResponse
	err = json.Unmarshal(value, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// -------------------- Private functions -----------------------------------
func (pub *Pubnub) httpRequest(requestURL string, isSubscribe bool) ([]byte, int, error) {

	retryCount := 0
retryRequest:
	req, err := http.NewRequest("GET", pub.origin+requestURL, nil)
	// User Agent
	req.Header.Set("User-Agent", fmt.Sprintf("ua_string=(%s) %s",
		sdkIdentificationParamKey,
		sdkIdentificationParamVal,
	))

	response, err := pub.GetClient().Do(req)
	if err != nil {
		// Check for net.Timeout errors
		if e, ok := err.(*url.Error); ok {
			if nerr, ok := e.Err.(net.Error); ok && nerr.Timeout() && retryCount < pub.connectRetry {
				retryCount += 1
				goto retryRequest
			}
		}
		return nil, 0, err
	}
	defer response.Body.Close()

	// readBody
	bodyContents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, response.StatusCode, err
	}

	return bodyContents, response.StatusCode, nil

}

func (pub *Pubnub) setOrGetTransport(isSubscribe bool) http.RoundTripper {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(netw, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(netw, addr, time.Duration(pub.connectTimeout)*time.Second)

			if c != nil {
				if isSubscribe {
					deadline := time.Now().Add(time.Duration(pub.subscribeTimeout) * time.Second)
					c.SetDeadline(deadline)
					pub.subscribeConn = c
				} else {
					deadline := time.Now().Add(time.Duration(pub.nonSubscribeTimeout) * time.Second)
					c.SetDeadline(deadline)
					pub.nonSubscribeConn = c
				}
			} else {
				err = fmt.Errorf("Error in initializing connection: %s", err.Error())
			}

			if err != nil {
				return nil, err
			}

			return c, nil
		}}

	return transport
}

func (pub *Pubnub) checkSecretKeyAndAddSignature(opURL, requestURL string) string {
	if len(pub.secretKey) > 0 {
		opURL = fmt.Sprintf("%s&timestamp=%d", opURL, time.Now().Unix())

		var reqURL *url.URL
		reqURL, urlErr := url.Parse(opURL)
		if urlErr != nil {
			return opURL
		}
		rawQuery := reqURL.RawQuery

		//sort query
		query, _ := url.ParseQuery(rawQuery)
		signature := getHmacSha256(
			pub.secretKey,
			pub.subscribeKey+"\n"+pub.publishKey+"\n"+requestURL+"\n"+query.Encode(),
		)
		opURL = fmt.Sprintf("%s&signature=%s", opURL, signature)
		return opURL
	}
	return opURL
}

// encodeJSONAsPathComponent properly encodes serialized JSON
// for placement within a URI path
func encodeJSONAsPathComponent(jsonBytes string) string {
	u := &url.URL{Path: jsonBytes}
	return strings.TrimLeft(u.String(), "./")
}

func getHmacSha256(secretKey string, input string) string {
	hmacSha256 := hmac.New(sha256.New, []byte(secretKey))
	hmacSha256.Write([]byte(input))
	rawSig := base64.StdEncoding.EncodeToString(hmacSha256.Sum(nil))
	return strings.Replace(strings.Replace(rawSig, "+", "-", -1), "/", "_", -1)
}
