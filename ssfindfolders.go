package mfsslib

import (
	"fmt"
	"strings"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"encoding/xml"
)

var SsPaths []SsPath

func SsFindFolders (sssession, ssroot, sspath string) {
	// init						0		"/"
	
	if sspath == "/" {
		SsPaths = append(SsPaths, SsPath{sspath, sspath, ssroot, ssroot})
	}
	
	urlToCall, err := url.Parse("http://api.sendspace.com/rest/")
	if err != nil {
		log.Fatal(err)
	}
	
	urlParams := urlToCall.Query()
	urlParams.Set("method", "folders.getcontents")
	urlParams.Add("session_key", sssession)
	urlParams.Add("folder_id", ssroot)
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
		
	var q Query
	xml.Unmarshal(contents, &q.Res)		// creates File and Folder objects
		
	for _, f := range q.Res.Files {
		thispath := sspath + f.Name
		SsPaths = append(SsPaths, SsPath{thispath, f.Name, f.Id, f.Parent})		
	}

	for _, f := range q.Res.Folders {
		newpath := sspath + f.Name + "/"
		SsPaths = append(SsPaths, SsPath{newpath, f.Name, f.Id, f.Parent})
		SsFindFolders(sssession, f.Id, newpath)
	}
}


func SsMkFolder (filename, parent, sssession string, i int, path string) {
	var parentid string
	
	if parent == "" {
		fmt.Println("ERROR")
		return
	} else {
		parentid = parent
	}
	
	urlToCall, err := url.Parse("http://api.sendspace.com/rest/")
	if err != nil {
		log.Fatal(err)
	}
	
	urlParams := urlToCall.Query()
	urlParams.Set("method", "folders.create")
	urlParams.Add("session_key", sssession)
	urlParams.Add("name", filename)
	urlParams.Add("parent_folder_id", parentid)
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
		
	var q Query
	xml.Unmarshal(contents, &q.Res)

	// copy new data to the struct array
	for _, f := range q.Res.Folders {
		MfPaths[i].SsD = f.Id
		MfPaths[i].SsP = parentid
		
		// insert new folder as parent for all matching paths
		// (will be recursively set to children of the parent, if needed)
		for n, m := range MfPaths {
			if strings.HasPrefix(m.Path, path) {
				if m.SsD == "" {
					MfPaths[n].SsP = f.Id
				}
			}
		}
	}
}