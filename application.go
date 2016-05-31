package main

import (
    "io/ioutil"
    "log"
    "net/http"
    "bytes"
    "encoding/json"
    "os"
    "fmt"
    "strings"
)

/* NOTE: Taking advantage of other projects?
 * go-plivo may be a better way to handle in the future
 */

// TODO: roll id+token and find way to obfuscate
const auth_id string = "MANMM4ZDRJMMMXZGUZNZ"
const auth_token string = "YTJlMDg5ZDg0OWIxZTA1OGU4NDNiMzQxOWQyY2Iw"
const source_phone string = "18056684235"

type plivoMessage struct {
    Src  string `json:"src,omitempty"`
    Dst  string `json:"dst,omitempty"`
    Text string `json:"text,omitempty"`
}

func main() {
    port := os.Getenv("PORT")
        if port == "" {
            port = "5000"
        }

        f, _ := os.Create("/var/log/golang/golang-server.log")
        defer f.Close()
        log.SetOutput(f)

        const indexPage = "public/index.html"
        const confirmPage = "public/confirmation.html"
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            if r.Method == "POST" {
                if buf, err := ioutil.ReadAll(r.Body); err == nil {
                    log.Printf("Received message: %s\n", string(buf))
                }
            } else {
                log.Printf("Serving %s to %s...\n", indexPage, r.RemoteAddr)
                http.ServeFile(w, r, indexPage)
            }
        })

        const plivoBaseUrl string = "https://api.plivo.com/v1/Account/"
        var plivoApiUrl string = plivoBaseUrl + auth_id + "/Message/"
        http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request){
            err := r.ParseForm()
            if err != nil {
                // Handle error here via logging and then return            
            }
            numbers := r.PostFormValue("numbers")
            message := r.PostFormValue("message")
           
            replacer := strings.NewReplacer(",", "<")
            numbers = replacer.Replace(numbers)
            msg := plivoMessage{source_phone, numbers, message}
            jsonToPlivo, err := json.Marshal(msg)

            //Send a post request
            req, err := http.NewRequest("POST", plivoApiUrl, bytes.NewBuffer(jsonToPlivo))
            if err != nil {
                log.Printf("Could not form POST request: ")
                log.Fatal(err)
            }
            req.SetBasicAuth(auth_id, auth_token)
            req.Header.Add("Content-Type", "application/json")
            client := &http.Client{}
            resp, err := client.Do(req)
            if err != nil {
                panic(err)
            }
            defer resp.Body.Close()

            fmt.Println("response Status:", resp.Status)
            fmt.Println("response Headers:", resp.Header)
            body, _ := ioutil.ReadAll(resp.Body)
            fmt.Println("response Body:", string(body))

            http.ServeFile(w, r, confirmPage)
        })

        http.HandleFunc("/scheduled", func(w http.ResponseWriter, r *http.Request){
            if r.Method == "POST" {
            log.Printf("Received task %s scheduled at %s\n", r.Header.Get("X-Aws-Sqsd-Taskname"), r.Header.Get("X-Aws-Sqsd-Scheduled-At"))
            }
        })

        log.Printf("Listening on port %s\n\n", port)
        http.ListenAndServe(":"+port, nil)
}
