package msteams

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	botio "github.com/enriquebris/goagent/io"
	"github.com/gorilla/mux"
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
}

func (st *Input) initialize(port string, certFilePath string, keyFilePath string) error {
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

func (st *Input) initAPI(port string, certFilePath string, keyFilePath string) {
	router := mux.NewRouter()

	// self-signed SSL cert: https://stackoverflow.com/questions/63588254/how-to-set-up-an-https-server-with-a-self-signed-certificate-in-golang

	// endpoints
	// ping
	router.HandleFunc("/api/v1/msteams/ping", st.endpointGETPing).Methods("GET")
	// outgoing (to receive MSTeams outgoing messages)
	router.HandleFunc("/api/v1/msteams/outgoing", st.endpointPOSTOutgoing).Methods("POST")

	fmt.Println("msteams api is ready")
	//fmt.Println(http.ListenAndServe(port, router))
	// https://stackoverflow.com/questions/63588254/how-to-set-up-an-https-server-with-a-self-signed-certificate-in-golang
	fmt.Println(http.ListenAndServeTLS(port, certFilePath, keyFilePath, router))
}

func (st *Input) endpointGETPing(w http.ResponseWriter, req *http.Request) {
	outputJSON(w, http.StatusOK, OutgoingResponse{Type: "message", Text: "pong"})
}

func (st *Input) endpointPOSTOutgoing(w http.ResponseWriter, req *http.Request) {
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
