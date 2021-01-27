package msteams

import (
	"fmt"

	botio "github.com/enriquebris/goagent/io"
)

type Output struct{}

func NewOutput() *Output {
	return &Output{}
}

func (st *Output) Send(messageType string, message string, inputData botio.Metadata, outputData botio.Metadata) error {
	// get the response
	respChan, ok := inputData[responseChannelField].(*chan string)
	if !ok {
		fmt.Printf("[ERROR] can't get *chan string to output message: %v\n", message)
		return fmt.Errorf("can't get *chan string to output message: %v\n", message)
	}

	// send response back to input (non-blocking channel operation)
	select {
	case *respChan <- message:
		// message could be delivered back to sender (input)

	default:
		// message could not be delivered back to sender, timeout

		fmt.Printf("[ERROR] response message could not be delivered back to input: %v\n", message)
		// TODO ::: publish the message in a different way
	}

	return nil
}
