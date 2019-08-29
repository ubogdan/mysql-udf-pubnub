package main

/*
#cgo CFLAGS: -I/usr/include/mysql -DMYSQL_DYNAMIC_PLUGIN -DMYSQL_ABI_CHECK
#include <stdio.h>
#include <mysql.h>
#include <string.h>

static int is_arg_string(UDF_ARGS *args,int arg_num) {
	if (args->arg_count > arg_num &&
		args->arg_type[arg_num] == STRING_RESULT) {
		return 1;
	}
	return 0;
}

static char* get_string_val(UDF_ARGS *args, int arg_num) {
	if (args->arg_count > arg_num) {
		return args->args[arg_num];
	}
}

static long long get_int_val(UDF_ARGS *args, int arg_num) {
	long long int_val;
	if (args->arg_count > arg_num) {
		int_val = *((long long*) args->args[arg_num]);
	}
	return int_val;
}

*/
import "C"

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

var w *worker

const (
	pubKey = "pub-c-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	subKey = "sub-c-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	secKey = "sec-c-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
)

type (
	publishMessage struct {
		Channel string // Channel
		Store   bool   // Store in history
		Online  bool   // Send only if active grants on chan
		Message []byte // Json message
	}

	grantMessage struct {
		Channel              string // Channel
		Auth                 string // Auth key
		Read, Write, Manager bool   // Rights
		Ttl                  int    // TTL
	}
)

func main() {}

//export pubnub_grant_init
func pubnub_grant_init(
	initid *C.UDF_INIT,
	args *C.UDF_ARGS,
	message *C.char,
) C.my_bool {

	if args.arg_count != 4 {
		C.strcpy(message, C.CString("pubnub_grant(channel string, auth string, rights string, ttl int ). \n"))
		return 1
	}

	if C.is_arg_string(args, 0) == 0 {
		C.strcpy(message, C.CString("channel param is not string\n"))
		return 1
	}

	if C.is_arg_string(args, 1) == 0 {
		C.strcpy(message, C.CString("auth param is not string\n"))
		return 1
	}

	if C.is_arg_string(args, 2) == 0 {
		C.strcpy(message, C.CString("rights param is not string\n"))
		return 1
	}

	return 0
}

//export pubnub_grant
func pubnub_grant(
	initid *C.UDF_INIT,
	args *C.UDF_ARGS,
	result *C.char,
	length *C.ulong,
	is_null *C.char,
	error *C.char,
) C.longlong {

	chann, auth, rights, ttlString :=
		C.GoString(C.get_string_val(args, 0)),
		C.GoString(C.get_string_val(args, 1)),
		C.GoString(C.get_string_val(args, 2)),
		C.GoString(C.get_string_val(args, 3))

	channel, v := validate(chann)
	if !v {
		log.Printf("Invalid channel name %q for grant %q !", chann, channel)
		return 1
	}

	ttl, err := strconv.Atoi(ttlString)
	if err != nil {
		// Default TTL value (from Pubnub doc)
		ttl = 1440
	}

	w.Grant(chann, auth, rights, ttl)

	return 0
}

//export pubnub_publish_init
func pubnub_publish_init(
	initid *C.UDF_INIT,
	args *C.UDF_ARGS,
	message *C.char,
) C.my_bool {

	if args.arg_count < 2 {
		C.strcpy(message, C.CString("pubnub_publish(channel string, message string, [flags string]). \n"))
		return 1
	}

	if C.is_arg_string(args, 0) == 0 {
		C.strcpy(message, C.CString("channel param is not string\n"))
		return 1
	}

	if C.is_arg_string(args, 1) == 0 {
		C.strcpy(message, C.CString("message param is not string\n"))
		return 1
	}

	return 0
}

//export pubnub_publish
func pubnub_publish(
	initid *C.UDF_INIT,
	args *C.UDF_ARGS,
	result *C.char,
	length *C.ulong,
	is_null *C.char,
	error *C.char,
) C.longlong {

	chann, message, flags :=
		C.GoString(C.get_string_val(args, 0)),
		C.GoString(C.get_string_val(args, 1)),
		""

	if args.arg_count > 2 {
		flags = C.GoString(C.get_string_val(args, 2))
	}

	payload := []byte(message)
	var js map[string]interface{}
	if err := json.Unmarshal(payload, &js); err != nil {
		log.Printf("Failed to decode json %q : %s", payload, err)
		return 1
	}

	// Avoid putting in queue messages with invalid payload
	channel, v := validate(chann)
	if !v {
		log.Printf("Invalid channel name %q for publish %q!", chann, channel)
		return 1
	}

	w.Publish(channel, payload, flags)
	return 0

}

func validate(channel string) (string, bool) {
	l := len(channel)
	if l < 1 || l > 92 {
		return channel, false
	}
	channel = strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '_' || r == '-':
			return r
		}
		return -1
	}, channel)

	return channel, len(channel) == l
}
