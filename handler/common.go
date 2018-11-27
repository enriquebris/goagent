package handler

import (
	"fmt"

	"github.com/enriquebris/goagent/cmd"
	botio "github.com/enriquebris/goagent/io"
	"github.com/enriquebris/goagent/message"
	"github.com/op/go-logging"
)

const (
	defaultCommandHelp = "help"
	defaultTagError    = "error"
	defaultTagHelp     = "help"
)

type Common struct {
	log         *logging.Logger
	commandHelp string
	tagError    string
	tagHelp     string
}

func NewCommon(log *logging.Logger) *Common {
	ret := &Common{}
	ret.initialize(log)

	return ret
}

func (st *Common) initialize(log *logging.Logger) {
	st.log = log
	st.commandHelp = defaultCommandHelp
	st.tagError = defaultTagError
	st.tagHelp = defaultTagHelp
}

func (st *Common) SetCommandHelp(commandHelp string) {
	st.commandHelp = commandHelp
}

func (st *Common) SetTagHelp(tagHelp string) {
	st.tagHelp = tagHelp
}

func (st *Common) SetTagError(tagError string) {
	st.tagError = tagError
}

// GetErrorHandler returns a error handler
func (st *Common) GetErrorHandler(tags []string) cmd.CMDHandler {
	return func(command cmd.CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output) {
		if cmdContent == st.commandHelp {
			// remove the tagError from the tag list
			tagsHelp := removeSliceElement(tags, st.tagError)
			// add the help tag
			tagsHelp = append(tagsHelp, st.tagHelp)

			help := getHelpForCMD(command, st.tagHelp)

			message.SendMessageToOutput(
				fmt.Sprintf("`%v %v:`\n%v", command.Pattern[0], st.tagHelp, help),
				inputMetadata,
				botio.Metadata{
					"Tags": tagsHelp,
				},
				outputs,
			)
		} else {
			message.SendDontUnderstandMessage(cmdContent, tags, inputMetadata, outputs)
		}
	}
}

// GetParametersIncorrectTypeHandler returns a function to handle incorrect param's types
func (st *Common) GetParametersIncorrectTypeHandler(tags []string) cmd.CMDHandler {
	return func(command cmd.CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output) {
		tmp, _ := metadata[cmd.CMDExtraDataParametersIncorrectType]
		extraMessage := ""
		if paramsData, ok := tmp.([]cmd.CMDParam); ok {
			extraMessage = ": "
			for i := 0; i < len(paramsData); i++ {
				extraMessage += fmt.Sprintf("\nparameter: '%v'\nvalue: %v\nexpected type : %v", paramsData[i].ID, paramsData[i].Value, paramsData[i].Type)
			}
		} else {
			st.log.Error("Unexpected data for metadata[CMDExtraDataParametersIncorrectType]")
		}

		message.SendMessageToOutput(
			fmt.Sprintf("Incorrect parameter's type%v", extraMessage),
			inputMetadata,
			botio.Metadata{
				"Tags": tags,
			},
			outputs,
		)
	}
}

// GetParametersMissing returns a function to handle missing params
func (st *Common) GetParametersMissing(tags []string) cmd.CMDHandler {
	return func(command cmd.CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output) {
		tmp, _ := metadata[cmd.CMDExtraDataParametersMissing]
		extraMessage := ""
		if paramsData, ok := tmp.([]cmd.CMDParam); ok {
			extraMessage = ": "
			for i := 0; i < len(paramsData); i++ {
				extraMessage += fmt.Sprintf("\nparameter: '%v'\nexpected type : %v", paramsData[i].ID, paramsData[i].Type)
			}
		} else {
			st.log.Error("Unexpected data for metadata[CMDExtraDataParametersMissing]")
		}

		message.SendMessageToOutput(
			fmt.Sprintf("Missing parameters%v", extraMessage),
			inputMetadata,
			botio.Metadata{
				"Tags": tags,
			},
			outputs,
		)
	}
}

// GetParametersExtra returns a function to handle extra params
func (st *Common) GetParametersExtra(tags []string) cmd.CMDHandler {
	return func(command cmd.CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output) {
		tmp, _ := metadata[cmd.CMDExtraDataParametersExtra]
		extraMessage := ""
		if paramsData, ok := tmp.([]cmd.CMDParam); ok {
			extraMessage = ": "
			for i := 0; i < len(paramsData); i++ {
				extraMessage += fmt.Sprintf("\nvalue: %v", paramsData[i].Value)
			}
		} else {
			st.log.Error("Unexpected data for metadata[CMDExtraDataParametersExtra]")
		}

		message.SendMessageToOutput(
			fmt.Sprintf("There are some extra parameters, I don't know what to do with them%v", extraMessage),
			inputMetadata,
			botio.Metadata{
				"Tags": tags,
			},
			outputs,
		)
	}
}

func (st *Common) GetRestrictionsHandler(tags []string) cmd.CMDHandler {
	return func(command cmd.CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output) {
		st.log.Error(metadata["message"])

		message.SendMessageToOutput(
			"Sorry, this action cannot be executed because a restriction",
			inputMetadata,
			botio.Metadata{
				"Tags": tags,
			},
			outputs,
		)
	}
}
