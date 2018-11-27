package example

import (
	"github.com/enriquebris/goagent/cmd"
	"github.com/enriquebris/goagent/handler"
	botio "github.com/enriquebris/goagent/io"
	"github.com/enriquebris/goagent/message"
)

func GetMainCMD(commonHandler *handler.Common) cmd.CMD {
	return cmd.CMD{
		PatternType:  cmd.CMDTypeWord,
		Pattern:      []string{"agent", "agent,", "@agent", "@agent,"},
		Description:  "Agent XYZ",
		Handler:      defaultHandler,
		HandlerError: commonHandler.GetErrorHandler([]string{"error"}),
		SubCommands: []cmd.CMD{
			{
				PatternType: cmd.CMDTypeWord,
				Pattern: []string{"version"},
				Description: "Version",
				Handler: versionHandler,
				HandlerError: commonHandler.GetErrorHandler([]string{"error"}),
			},
		},
	}
}

func defaultHandler(cmd cmd.CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output) {
	message.SendMessageToOutput(
		"Hi!",
		inputMetadata,
		nil,
		outputs,
	)
}

func versionHandler(cmd cmd.CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output) {
	message.SendMessageToOutput(
		"Agent XYZ version 1",
		inputMetadata,
		botio.Metadata{
			"Tags": []string{"version"},
		},
		outputs,
	)
}
