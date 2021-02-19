package msteams

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	botio "github.com/enriquebris/goagent/io"
)

const (
	Origin                 = "msteams"
	maxListenQueueCapacity = 500
	messagePattern         = "<at>([a-zA-Z]+)</at>(.+)\n"
	timeoutSeconds         = 5
	timeoutMessage         = "response could not be delivered in time"
)

type MSTeamsOutgoingData struct {
	Text            string
	Data            Outgoing
	ResponseChannel *chan string
}

func NewInput(certFilePath string, keyFilePath string) (*Input, error) {
	ret := &Input{}
	if err := ret.initialize(":443", certFilePath, keyFilePath); err != nil {
		return nil, err
	}

	return ret, nil
}

type Input struct {
	// channel to send data from endpoints to Listen(...) function
	dataToListen chan MSTeamsOutgoingData
	// message regexp
	msgRegexp *regexp.Regexp

	minTLSVersion uint16
	headers       map[string]string
}

func (st *Input) initialize(port string, certFilePath string, keyFilePath string) error {
	// TLS 1.0 by default
	st.minTLSVersion = tls.VersionTLS10
	// headers
	st.headers = make(map[string]string)

	// channel to send input messages to agent
	st.dataToListen = make(chan MSTeamsOutgoingData, maxListenQueueCapacity)

	var err error
	// message regex
	st.msgRegexp, err = regexp.Compile(messagePattern)
	if err != nil {
		return err
	}

	// API listeners
	go st.initAPI(port, certFilePath, keyFilePath)

	return nil
}

func (st *Input) SetMinTLSVersion(version uint16) {
	st.minTLSVersion = version
}

func (st *Input) AddHeader(key, value string) {
	st.headers[key] = value
}

func (st *Input) addAllHeaders(w http.ResponseWriter) {
	for k, v := range st.headers {
		w.Header().Add(k, v)
	}
}

func (st *Input) initAPI(port string, certFilePath string, keyFilePath string) {
	// generate a `Certificate` struct
	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		fmt.Println(err)
	}

	// create a custom server with `TLSConfig`
	s := &http.Server{
		Addr:    port,
		Handler: nil, // use `http.DefaultServeMux`
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   st.minTLSVersion, // min version: TLS 1.2
		},
	}

	// ping
	http.HandleFunc("/api/v1/msteams/ping", st.endpointGETPing)
	// api/v1/msteams/outgoing
	http.HandleFunc("/api/v1/msteams/outgoing", st.endpointPOSTOutgoing)

	// run server
	log.Fatal(s.ListenAndServeTLS("", ""))
}

func (st *Input) endpointGETPing(w http.ResponseWriter, req *http.Request) {
	// add headers
	st.addAllHeaders(w)

	if req.Method == "GET" {
		outputJSON(w, http.StatusOK, OutgoingResponse{Type: "message", Text: "pong"})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)

}

func (st *Input) endpointPOSTOutgoing(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// add headers
	st.addAllHeaders(w)

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("[ERROR] outgoing data could not be read:  %v\n", err.Error())
		return
	}

	var msgData Outgoing
	if err := json.Unmarshal(b, &msgData); err != nil {
		fmt.Printf("[ERROR] outgoing data could not be parsed\nbody: %v\nerror: %v\n", string(b), err.Error())
	}

	// 1 - get the message
	msgText := st.getPlainMessage(msgData)
	fmt.Printf("plain: %v\ntext:%v\n", msgText, msgData.Text)
	// 2 - send message to st.dataToListen
	respChan := make(chan string, 2)

	data := MSTeamsOutgoingData{
		Text: msgText,
		Data: msgData,
		// channel to receive the message (from MSOutput) to post in MSTeams
		ResponseChannel: &respChan,
	}

	// sends the msteams' outgoing data to our input listener (Listen method)
	select {
	case st.dataToListen <- data:
	default:
		fmt.Printf("[ERROR] outgoing data could not be sent to MSTeamsInput.Listen\nbody: %v\n", b)
		return
	}

	// Ath this point the message/cmd was sent to our listener (Listen method), it is going to be processed and the
	// response will come from our msteams' output.
	//
	// Why? We want to send the response back to msteams using the same http.ResponseWriter (w).
	// In case we go out this function's scope, an automatic response will be sent (maybe via a defer or timeout) and
	// we don't want that.
	//
	// wait for the output's message (non-blocking channel operation)
	select {
	case respMessage := <-respChan:
		// response in time
		outputJSON(w, http.StatusOK, OutgoingResponse{Type: "message", Text: respMessage})

	case <-time.After(timeoutSeconds * time.Second):
		// timeout

		// close response channel
		close(respChan)
		// output timeout message
		outputJSON(w, http.StatusOK, OutgoingResponse{Type: "message", Text: timeoutMessage})
	}

}

// getPlainMessage returns plain text message
func (st *Input) getPlainMessage(data Outgoing) string {
	rg := st.msgRegexp.FindStringSubmatch(data.Text)
	fmt.Println(rg)
	if len(rg) > 2 {
		return strings.TrimSpace(rg[1]) + " " + strings.TrimSpace(rg[2])
	}

	return ""
}

func (st *Input) Listen(chInputEntry chan botio.InputEntry) error {
	for {
		select {
		case msg := <-st.dataToListen:
			// transform entry into botio.InputEntry
			inputEntry := botio.InputEntry{
				Origin: Origin,
				Query:  msg.Text,
				// Entry to map[string]interface{}
				InputMetadata: map[string]interface{}{
					responseChannelField: msg.ResponseChannel,
				},
				GeneralMetadata: map[string]interface{}{},
			}

			chInputEntry <- inputEntry
		}
	}

	return nil
}
