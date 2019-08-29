package main

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := map[string]bool{
		"": false, // Empty channel name
		"pbx_ad0fdc9118b741c89cfa1963571c4b64":                                                                true,
		"private_chat_086a3b5a6a7be779eacc5c63c3b83db4":                                                       true,
		"sms_cf1d42cf783bd46473f5817e1f99a53b":                                                                true,
		"voicemail_3_021bbc7ee20b71134d53e20206bd6feb":                                                        true,
		"pbx_ad0fdc9118b741c89cfa1963571c4b6489cfa1963d0fdc911ad0fdc9118b741c89cfa1963571c8b74cfa1963571c4b6": false, // len > 92
		"voicemail_!_021bbc7ee20b71134d53e20206bd6feb":                                                        false, //
		"pbx_ad0fdc9118b741c89cfa1963571c4b64#test":                                                           false,
		"¥bx_ad0fdc9118b741c89cfa1963571c4b64":                                                                false,
	}

	for channel, result := range tests {
		chann, valid := validate(channel)
		if valid != result {
			t.Errorf("Unexpected validation result %s : %s ", channel, chann)
		}
	}
}
