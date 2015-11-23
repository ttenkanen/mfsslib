package mfsslib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"github.com/bitly/go-simplejson"
)

var MfPaths []MfPath

func MfFindFolders(session, root, mfpath string, depth int) {

	if mfpath == "/" {
		MfPaths = append(MfPaths, MfPath{Path: mfpath, Filename: "", Descriptor: root, Parent: root})
	}

	// -----------------------------
	// GET INFO ABOUT THIS (root) FOLDER

	dInfoUrl, err := url.Parse("https://www.mediafire.com/api/1.4/folder/get_info.php")
	if err != nil {
		fmt.Println(err)
		return
	}

	dInfoUrlParams := dInfoUrl.Query()
	dInfoUrlParams.Set("session_token", session)
	dInfoUrlParams.Add("response_format", "json")
	dInfoUrlParams.Add("content_type", "folders")
	dInfoUrlParams.Add("filter", "all")
	dInfoUrlParams.Add("folder_key", root)
	dInfoUrl.RawQuery = dInfoUrlParams.Encode()

	dInfoResponse, _ := http.Get(dInfoUrl.String())
	defer dInfoResponse.Body.Close()

	dInfoJson, _ := ioutil.ReadAll(dInfoResponse.Body)

	dInfoSimpleJson, _ := simplejson.NewJson(dInfoJson)
	nrdstring, _ := dInfoSimpleJson.Get("response").Get("folder_info").Get("folder_count").String()
	nrd, _ := strconv.Atoi(nrdstring) // <-- nr of folders!!
	nrfstring, _ := dInfoSimpleJson.Get("response").Get("folder_info").Get("file_count").String()
	nrf, _ := strconv.Atoi(nrfstring) // <-- nr of files!!

	// -----------------------------
	// FIND FILES IN THIS (root) FOLDER

	dFilesUrl, err := url.Parse("https://www.mediafire.com/api/1.4/folder/get_content.php")
	if err != nil {
		fmt.Println(err)
		return
	}

	dFilesUrlParams := dFilesUrl.Query()
	dFilesUrlParams.Set("session_token", session)
	dFilesUrlParams.Add("response_format", "json")
	dFilesUrlParams.Add("content_type", "files")
	dFilesUrlParams.Add("filter", "all")
	dFilesUrlParams.Add("folder_key", root)
	dFilesUrl.RawQuery = dFilesUrlParams.Encode()

	dFilesResponse, _ := http.Get(dFilesUrl.String())
	defer dFilesResponse.Body.Close()

	dFilesJson, _ := ioutil.ReadAll(dFilesResponse.Body)
	dFilesSimpleJson, _ := simplejson.NewJson(dFilesJson)

	// LOOP THROUGH FILES
	for f := 0; f < nrf; f++ {
		fname, _ := dFilesSimpleJson.Get("response").Get("folder_content").Get("files").GetIndex(f).Get("filename").String()
		fdesc, _ := dFilesSimpleJson.Get("response").Get("folder_content").Get("files").GetIndex(f).Get("quickkey").String()
		thispath := mfpath + fname
		MfPaths = append(MfPaths, MfPath{Path: thispath, Filename: fname, Descriptor: fdesc, Parent: root})
	}

	// -----------------------------
	// FIND SUBFOLDERS
	dContentUrl, err := url.Parse("https://www.mediafire.com/api/1.4/folder/get_content.php")
	if err != nil {
		fmt.Println(err)
		return
	}

	dContentUrlParams := dContentUrl.Query()
	dContentUrlParams.Set("session_token", session)
	dContentUrlParams.Add("response_format", "json")
	dContentUrlParams.Add("content_type", "folders")
	dContentUrlParams.Add("filter", "all")
	dContentUrlParams.Add("folder_key", root)
	dContentUrl.RawQuery = dContentUrlParams.Encode()

	dContentResponse, _ := http.Get(dContentUrl.String())
	defer dContentResponse.Body.Close()

	dContentJson, _ := ioutil.ReadAll(dContentResponse.Body)
	dContentSimpleJson, _ := simplejson.NewJson(dContentJson)

	// LOOP THROUGH SUBFOLDERS
	for j := 0; j < nrd; j++ {
		dkey, _  := dContentSimpleJson.Get("response").Get("folder_content").Get("folders").GetIndex(j).Get("folderkey").String()
		dname, _ := dContentSimpleJson.Get("response").Get("folder_content").Get("folders").GetIndex(j).Get("name").String()
		dnrd, _  := dContentSimpleJson.Get("response").Get("folder_content").Get("folders").GetIndex(j).Get("folder_count").String()
		dnrf, _  := dContentSimpleJson.Get("response").Get("folder_content").Get("folders").GetIndex(j).Get("file_count").String()
		newpath  := mfpath + dname + "/"
		MfPaths = append(MfPaths, MfPath{Path: newpath, Filename: dname, Descriptor: dkey, Parent: root})

		// Check if we have a need to go deeper - if the folder is empty, no need to recurse into it
		dnrdi, _ := strconv.Atoi(dnrd)
		dnrfi, _ := strconv.Atoi(dnrf)
		if dnrfi+dnrdi != 0 {
			// RECURSE INTO THE SUBFOLDERS
			MfFindFolders(session, dkey, newpath, depth+1)
		}
	}
}
