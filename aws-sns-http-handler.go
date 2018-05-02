package main

import (
//    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "github.com/wallix/awless/logger"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Read body
    b, err := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
	http.Error(w, err.Error(), 500)
	return
    }
    var jsonb map[string]interface{}
    err = json.Unmarshal(b, &jsonb)
    if err != nil {
	http.Error(w, err.Error(), 500)
	return
    }
    mt := r.Header.Get("X-Amz-Sns-Message-Type")
    
    if mt == "SubscriptionConfirmation" {
	logger.Info("Subscription confirmation requested.")
	if SubscribeURL, ok := jsonb["SubscribeURL"].(string); ok {
	    logger.Info("SubscribeURL: ", SubscribeURL)
	    rs, err := http.Get(SubscribeURL)
	    if err != nil {
		logger.Error("http.Get() returned", err)
		return
	    }
	    logger.Info("Creq: ", rs)
	    bodyBytes, _ := ioutil.ReadAll(rs.Body)
	    bodyString := string(bodyBytes)
	    logger.Info("CBody: ", bodyString)
	    defer rs.Body.Close()
	} else {
	    logger.Error("SubscribeURL is not defined")
	    return
	}
	
    }
    
    // fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
    logger.Info("Topic ARN: ", r.Header.Get("X-Amz-Sns-Topic-Arn"))
    logger.Info("Token: ", jsonb["Token"])
    logger.Info("Request: ", r)
    logger.Info("Body: ", jsonb)
}

func main() {
    logger.Info("Started...")
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}