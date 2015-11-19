package mfsslib

import (
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"encoding/xml"
)

func SsGetUploadInfo (session string) Query {
	
	urlToCall, err := url.Parse("http://api.sendspace.com/rest/")
	if err != nil {
		log.Fatal(err)
	}
	
	urlParams := urlToCall.Query()
	urlParams.Set("method", "upload.getinfo")
	urlParams.Add("session_key", session)
	urlParams.Add("speed_limit", "0")	
	urlToCall.RawQuery = urlParams.Encode()
	
	response, err := http.Get(urlToCall.String())
	defer response.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
		
	var uq Query
	xml.Unmarshal(contents, &uq.Res)
		
	return uq
}



