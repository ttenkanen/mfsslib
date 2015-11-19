package mfsslib

import (
	"encoding/hex"
	"crypto/sha1"
	"log"
	"net/http"
	"net/url"
	"github.com/bitly/go-simplejson"
	"errors"
)

func MfGetSessionToken (mfappid, mfapikey, mfemail, mfpasswd string) (string, error) {
	
	mfsignat := mfemail + mfpasswd + mfappid + mfapikey
	hasher := sha1.New()
	hasher.Write([]byte(mfsignat))
	mfs := hex.EncodeToString(hasher.Sum(nil))
		
	urlToCall, err := url.Parse("https://www.mediafire.com/api/1.4/user/get_session_token.php")
	if err != nil {
		log.Fatal(err)
	}
	
	urlParams := urlToCall.Query()
	urlParams.Set("application_id", mfappid)
	urlParams.Add("signature", mfs)
	urlParams.Add("email", mfemail)
	urlParams.Add("password", mfpasswd)
	urlParams.Add("token_version", "1")
	urlParams.Add("response_format", "json")
	
	urlToCall.RawQuery = urlParams.Encode()
		
	response, err := http.Get(urlToCall.String())
	defer response.Body.Close()
	
    if err != nil {
        return "", errors.New("Getmftoken: network error")
    } else {
		js, _ := simplejson.NewFromReader(response.Body)
		mfsession, err := js.Get("response").Get("session_token").String()
		if err == nil {
			return mfsession, nil
		} else {
			message, _ := js.Get("response").Get("message").String()
			return "", errors.New("Getmftoken: " + message)
		}
    }	
}