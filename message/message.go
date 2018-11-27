package message

import (
	"fmt"

	botio "github.com/enriquebris/goagent/io"
)

const (
	dontUnderstandMessage = "Hmmm, I don't understand what do you mean by '%v'"
)

// sendMessageToOutput sends a message to the given outputs
func SendMessageToOutput(message string, inputMetadata botio.Metadata, outputMetadata botio.Metadata, outputs []botio.Output) {
	for i := 0; i < len(outputs); i++ {
		outputs[i].Send(botio.OutputMessageTypeDefault, message, inputMetadata, outputMetadata)
	}
}

// SendFramedMessageToOutput sends a framed message to the given putputs
func SendFramedMessageToOutput(message string, inputMetadata botio.Metadata, outputMetadata botio.Metadata, outputs []botio.Output) {
	for i := 0; i < len(outputs); i++ {
		outputs[i].Send(botio.OutputMessageTypeFramed, message, inputMetadata, outputMetadata)
	}
}

// SendDontUnderstandMessage sends a "do not understand" message to the given outputs
func SendDontUnderstandMessage(cmdContent string, tags []string, inputMetadata botio.Metadata, outputs []botio.Output) {
	SendMessageToOutput(
		fmt.Sprintf(dontUnderstandMessage, cmdContent),
		inputMetadata,
		botio.Metadata{
			"Tags": tags,
		},
		outputs,
	)
}
