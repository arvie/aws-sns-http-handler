package main

import (
//    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "encoding/base64"
    "crypto/x509"
    "encoding/pem"
    "github.com/wallix/awless/logger"
)

func getMessageBytesToSign(msg map[string]string) []byte {
    res := "Message\n"
    res += msg["Message"] + "\n"
    res += "MessageId\n"
    res += msg["MessageId"] + "\n"

    switch msg["Type"] {
	case "Notification":
	    if (msg["Subject"] != "") {
		res += "Subject\n"
	        res += msg["Subject"] + "\n"
	    }
	    res += "Timestamp\n"
	    res += msg["Timestamp"] + "\n"
	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
            res += "SubscribeURL\n"
            res += msg["SubscribeURL"] + "\n"
            res += "Timestamp\n"
            res += msg["Timestamp"] + "\n"
            res += "Token\n"
            res += msg["Token"] + "\n"
    }

    res += "TopicArn\n"
    res += msg["TopicArn"] + "\n"
    res += "Type\n"
    res += msg["Type"] + "\n"

    logger.Info("signedmessage: ", res)
    return []byte(res)
}

func checkSignatureV1(msg map[string]string) bool {
    url := msg["SigningCertURL"]
    sig, err := base64.StdEncoding.DecodeString(msg["Signature"])
    if err != nil {
	logger.Error("Signature base64.decode returned", err)
	return false
    }
    logger.Info("Sig: ", sig)
    logger.Info("checkSignatureV1: ", url)
    rs, err := http.Get(url)
    if err != nil {
	logger.Error("Cert Loading http.Get() returned", err)
	return false
    }
    logger.Info("PEM: ", rs)
    bodyBytes, _ := ioutil.ReadAll(rs.Body)
    defer rs.Body.Close()
    block, _ := pem.Decode(bodyBytes)
    if block == nil {
	logger.Error("failed to parse certificate PEM")
	return false
    }
    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
	logger.Error("failed to parse certificate: " + err.Error())
	return false
    }
    ok := cert.CheckSignature(x509.SHA1WithRSA, getMessageBytesToSign(msg), sig)
    if ok != nil {
	logger.Error("Signature is not valid: " + err.Error())
	return false
    }
    
    return true
}


func handler(w http.ResponseWriter, r *http.Request) {
    // Read body
    b, err := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
	http.Error(w, err.Error(), 500)
	return
    }
    var jsonb map[string]string
    err = json.Unmarshal(b, &jsonb)
    if err != nil {
	http.Error(w, err.Error(), 500)
	return
    }

    SignatureVersion := jsonb["SignatureVersion"]
    if SignatureVersion != "1" {
	logger.Error("Unsupported signature version.")
	return
    }
    if !checkSignatureV1(jsonb) {
	logger.Error("Signature is invalid.")
	return
    }
    return
    mt := r.Header.Get("X-Amz-Sns-Message-Type")

    
    
    if mt == "SubscriptionConfirmation" {
	logger.Info("Subscription confirmation requested.")
	SubscribeURL := jsonb["SubscribeURL"]
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