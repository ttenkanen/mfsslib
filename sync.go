package mfsslib

import (
	"reflect"
	"mime/multipart"
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strings"
	"fmt"
	"io"
	"io/ioutil"
)


type Query struct {
	Res 		Result 		`xml:"result"`
}


type Result struct {
	Token 		string 		`xml:"token"`
	Sessionkey 	string 		`xml:"session_key"`
	Folders 	[]Folder 	`xml:"folder"`
	Files		[]File 		`xml:"file"`
	Upload		Upload		`xml:"upload"`
}


type Folder struct {
	Name		string 		`xml:"name,attr"`
	Id			string		`xml:"id,attr"`
	Parent		string		`xml:"parent_folder_id,attr"`
}


type File struct {
	Name		string 		`xml:"name,attr"`
	Id			string		`xml:"id,attr"`
	Parent		string		`xml:"parent_folder_id,attr"`
}


type Upload struct {
	Url			string		`xml:"url,attr"`
	Max			string		`xml:"max_file_size,attr"`
	Extra		string		`xml:"extra_info,attr"`
	Uid			string		`xml:"upload_identifier,attr"`
	Prog		string		`xml:"progress_url,attr"`
}


type SsPath struct {
	Path, Filename, Descriptor, Parent string
}


type MfPath struct {
	Path, Filename, Descriptor, Parent, SsD, SsP string
}


func inArray (arr []SsPath, str string) bool {
	for _, s := range arr {
		if s.Path == str {
			return true
		}
	}
	return false
}


func getSsD (arr []MfPath, needle string) string {
	for _, m := range arr {
		if m.Descriptor == needle {
			return m.SsD
		}
	}
	return ""
}


func isFolderPath (path string) bool {
	return strings.HasSuffix(path, "/")
}


func Sync(mfuser, mfpass, ssuser, sspass string, o io.Writer) bool {
	
	useflush := false
	var flusher http.Flusher
	
	if strings.Contains(reflect.TypeOf(o).String(), "http") {
		flusher = o.(http.Flusher)
		useflush = true
	}
	
	// sendspace api key	4R6XNCX3EA
	// mediafire api key	5v1vtcl1k8285gtbpc891jtijjxnit5v1cgopla8
	// mediafire app id		47574
	
	mfapi := "5v1vtcl1k8285gtbpc891jtijjxnit5v1cgopla8"
	mfapp := "47574"
	ssapi := "4R6XNCX3EA"
	
	// create channels for goroutines
	c := make(chan string)
	f := make(chan string)
	numfilestosync := 0
	
	// sign in to Mediafire and get a file path list
	mfsessionkey, err := MfGetSessionToken (mfapp, mfapi, mfuser, mfpass)
	if err != nil {
		fmt.Fprintln(o, err)
		return false
	}
	go func() {
		MfFindFolders(mfsessionkey, "myfiles", "/", 0)
		c <- "Mediafire file list ok"
	}()
	

	// sign in to Sendspace and get a file path list
	sstoken, err := SsGetToken(ssapi)
	if err != nil {
		fmt.Fprintln(o, err)
		return false
	}
	sssessionkey, err := SsLogin(sstoken, ssuser, sspass)
	if err != nil {
		fmt.Fprintln(o, err)
		return false
	}
	go func() {
		SsFindFolders(sssessionkey, "0", "/")
		c <- "Sendspace file list ok"
	}()
	

	// wait for both file path lists before going further
	for i := 0; i < 2; i++ {
        msg := <- c
		fmt.Fprintln(o, msg)
		if useflush {
			flusher.Flush()
		}
	}
	

	// copy Sendspace data to the Mediafire structs, if exists. don't care about the rest
	for _, s := range SsPaths {
		for i, m := range MfPaths {
			if s.Path == m.Path {
				MfPaths[i].SsD = s.Descriptor
				MfPaths[i].SsP = s.Parent
			}
		}	
	}
	

	// create folder tree on Sendspace as it is on Mediafire
	for i, m := range MfPaths {
		// (m.SsD == "") <==> folder does not exist on Sendspace
		if m.SsD == "" && isFolderPath(m.Path) {
			SsMkFolder(m.Filename, getSsD(MfPaths, m.Parent), sssessionkey, i, m.Path)
			fmt.Fprintln(o, m.Path + " -- CREATED")
		} else if isFolderPath(m.Path) {
			fmt.Fprintln(o, m.Path + " -- exists")
		}
		if useflush {
			flusher.Flush()
		}
	}

	
	// create files
	for _, m := range MfPaths {

		if !isFolderPath(m.Path) {

			numfilestosync += 1
			go func(m MfPath) {

				// (m.SsD == "") <==> file does not exist on Sendspace
				if m.SsD == "" {
					fmt.Fprintln(o, m.Path + " -- SYNCING")
					pr, pw := io.Pipe()

					go func() {
						// get a download link & download
						dlink := MfGetDownloadlink(mfsessionkey, m.Descriptor)
						dload, err := http.Get(dlink)
						if err != nil {
							log.Fatal(err)
						}
			
						// copy from download to pipe
						_, err = io.Copy(pw, dload.Body)
						if err != nil {
							log.Fatal(err)
						}

						//fmt.Fprintln(m.Path + " DOWLOADED")
						dload.Body.Close()
						pw.Close()
						c <- m.Filename + " downloaded"
					}()

					go func() {
						// get upload info for the file
						uq := SsGetUploadInfo(sssessionkey)
					
						urlToCall, err := url.Parse(uq.Res.Upload.Url)
						if err != nil {
							log.Fatal(err)
						}
				
						urlParams := urlToCall.Query()
						urlParams.Add("folder_id", getSsD(MfPaths, m.Parent))
						urlParams.Add("userfile", m.Filename)		
						urlToCall.RawQuery = urlParams.Encode()
						// have to do string concat as urlParams.Add encodes % => %25 and extra info fails then
						callUrl := urlToCall.String() + "&extra_info=" + uq.Res.Upload.Extra
		
						// create a writer for the upload				
						bodyBuf := &bytes.Buffer{}
						bodyWriter := multipart.NewWriter(bodyBuf)
						fileWriter, err := bodyWriter.CreateFormFile("userfile", m.Filename)
					    if err != nil {
			       			log.Fatal(err)
					    }
	
					    // iocopy from the pipe
					    _, err = io.Copy(fileWriter, pr)
						if err != nil {
							log.Fatal(err)
						}
				
						contentType := bodyWriter.FormDataContentType()
					    bodyWriter.Close()
						pr.Close()
	
			   			resp, err := http.Post(callUrl, contentType, bodyBuf)
					    defer resp.Body.Close()
						if err != nil {
							log.Fatal(err)
						}
				
					    resp_body, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							log.Fatal(err)
						}
				
						if strings.Contains(string(resp_body), "fail") {
							fmt.Fprintln(o, string(resp_body))
							return
						} else {
							//fmt.Fprintln(m.Path + " UPLOADED")
						}
						c <- m.Filename + " uploaded"
					}()

					// wait for both download & upload routines to complete
					for i := 0; i < 2; i++ {
						_ = <- c
					}
				
					// signal sync channel
					f <- m.Path + " -- DONE"
			
				} else {
					f <- m.Path + " -- exists"
				}
			
			}(m)
		}
	}

	// wait for all file syncs/exists to complete and print	
	for j := 0; j < numfilestosync; j++ {
		msg := <- f
		fmt.Fprintln(o, msg)
		if useflush {
			flusher.Flush()
		}
	}
	
	fmt.Fprintln(o, "Sync complete!")
	return true
}


