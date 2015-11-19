package mfsslib

import (
	"fmt"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"errors"
	"encoding/hex"
	"strings"
	"crypto/md5"
)

func SsGetToken (ssapikey string) (string, error) {
	// sendspace api key	4R6XNCX3EA
	// http://api.sendspace.com/rest/?method=auth.createtoken&api_key=12DPC5Q11N&api_version=1.0&response_format=xml&app_version=0.1
	
	urlToCall, err := url.Parse("http://api.sendspace.com/rest/")
	if err != nil {
		log.Fatal(err)
	}
	
	urlParams := urlToCall.Query()
	urlParams.Set("method", "auth.createtoken")
	urlParams.Add("api_key", ssapikey)
	urlParams.Add("api_version", "1.0")
	
	urlToCall.RawQuery = urlParams.Encode()
		
	response, err := http.Get(urlToCall.String())
	defer response.Body.Close()
	
	if err != nil {
        return "", errors.New("Getsstoken: network error")
    } else {
		
		var q Query
		contents, _ := ioutil.ReadAll(response.Body)
		xml.Unmarshal(contents, &q.Res)
		
		if q.Res.Token != "" {
			return q.Res.Token, nil
		} else {
			fmt.Println(string(contents))
			return "", errors.New("Getsstoken: no token")
		}
    }
}


func SsLogin (token, username, password string) (string, error) {
	// sendspace api key	4R6XNCX3EA
	// http://api.sendspace.com/rest/?method=auth.login&token=57md654jwfl6l25idskzh8x3b5zwp46k&user_name=testuser&tokened_password=2cb501e4j86ef8ad17f6b26b90ee5764
	// lowercase(md5(token+lowercase(md5(password)))) 
	
	hasher := md5.New()
	b := []byte(password)
	hasher.Write(b)
	hashed_password := strings.ToLower(hex.EncodeToString(hasher.Sum(nil)))
	
	//fmt.Printf("%s\n", hashed_password)
	
	tokenizer := md5.New()
	d := []byte(token + hashed_password)
	tokenizer.Write(d)
	tokened_password := strings.ToLower(hex.EncodeToString(tokenizer.Sum(nil)))
	
	//fmt.Println(tokened_password)	
	
	urlToCall, err := url.Parse("http://api.sendspace.com/rest/")
	if err != nil {
		log.Fatal(err)
	}
	
	urlParams := urlToCall.Query()
	urlParams.Set("method", "auth.login")
	urlParams.Add("token", token)
	urlParams.Add("user_name", username)
	urlParams.Add("tokened_password", tokened_password)
	
	urlToCall.RawQuery = urlParams.Encode()
	
	//fmt.Println(urlToCall.String())
		
	response, err := http.Get(urlToCall.String())
	defer response.Body.Close()
	
	if err != nil {
        return "", errors.New("Sslogin: network error")
    } else {
		var q Query
		contents, _ := ioutil.ReadAll(response.Body)
		xml.Unmarshal(contents, &q.Res)
		
		if q.Res.Sessionkey != "" {
			return q.Res.Sessionkey, nil
		} else {
			fmt.Println(string(contents))
			return "", errors.New("Sslogin: login failed")
		}
    }
}

