package mfsslib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"github.com/bitly/go-simplejson"
)

func MfGetDownloadlink(session, quick string) string {

	dLinkUrl, err := url.Parse("http://www.mediafire.com/api/1.4/file/get_links.php")
	if err != nil {
		fmt.Println(err)
		os.Exit(101)
	}

	dLinkParams := dLinkUrl.Query()
	dLinkParams.Set("session_token", session)
	dLinkParams.Add("response_format", "json")
	dLinkParams.Add("quick_key", quick)
	dLinkParams.Add("link_type", "direct_download")
	dLinkUrl.RawQuery = dLinkParams.Encode()
	
	//fmt.Println("url: " + dLinkUrl.String())

	dLinkResponse, _ := http.Get(dLinkUrl.String())
	defer dLinkResponse.Body.Close()

	dLinkJson, _ := ioutil.ReadAll(dLinkResponse.Body)

	dLinkSimpleJson, _ := simplejson.NewJson(dLinkJson)
	dUrl, _ := dLinkSimpleJson.Get("response").Get("links").GetIndex(0).Get("direct_download").String()
	return dUrl

}
