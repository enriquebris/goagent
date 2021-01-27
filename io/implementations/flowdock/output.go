package flowdock

import (
	"fmt"

	"github.com/enriquebris/goflowdock"
	botio "github.com/enriquebris/goagent/io"
)

type FlowdockOutput struct {
	authToken      string
	messageManager *goflowdock.MessageManager
	username       string
}

func NewFlowdockOutput(authToken string, organization string, username string) *FlowdockOutput {
	ret := &FlowdockOutput{}
	ret.initialize(authToken, organization, username)

	return ret
}

func (st *FlowdockOutput) initialize(authToken string, organization string, username string) {
	st.authToken = authToken
	st.messageManager = goflowdock.NewMessageManager(authToken, organization)
	st.username = username
}

func (st *FlowdockOutput) Send(messageType string, message string, inputData botio.Metadata, outputData botio.Metadata) error {
	switch messageType {
	case botio.OutputMessageTypeDefault:
		fmt.Println(st.messageManager.SendMessage(buildMessageData(message, st.username, inputData, outputData)))
	case botio.OutputMessageTypeFramed:
		messageData := buildMessageData(message, st.username, inputData, outputData)
		messageData.Content = fmt.Sprintf("```%v```", messageData.Content)
		st.messageManager.SendMessage(messageData)
	default:
	}

	return nil
}

func buildMessageData(message string, username string, inputData botio.Metadata, outputData botio.Metadata) goflowdock.MessageData {
	ret := goflowdock.MessageData{
		Content:          message,
		ExternalUserName: username,
	}

	mergedMetadata := botio.MergeMetadata(inputData, outputData)

	ret.Flow = fmt.Sprintf("%v", mergedMetadata["Flow"])
	ret.Event = "message"
	if tags, ok := mergedMetadata["Tags"].([]string); ok {
		ret.Tags = tags
	}

	if _, ok := mergedMetadata["ThreadID"]; ok {
		ret.ThreadID = fmt.Sprintf("%v", mergedMetadata["ThreadID"])
	}

	return ret
}
