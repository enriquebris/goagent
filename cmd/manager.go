package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	botio "github.com/enriquebris/goagent/io"
)

type CMDManager struct {
	commands  []CMD
	errorChan chan error
}

func NewCMDManager(errorChan chan error) *CMDManager {
	ret := &CMDManager{}
	ret.initialize(errorChan)

	return ret
}

func (st *CMDManager) initialize(errorChan chan error) {
	st.commands = make([]CMD, 0)
	st.errorChan = errorChan
}

func (st *CMDManager) AddCommand(cmd CMD) error {
	// compile && save the regex pattern
	if cmd.PatternType == CMDTypeRegex {
		// compile the regex patterns
		for i := 0; i < len(cmd.Pattern); i++ {
			compiledPattern, err := regexp.Compile(cmd.Pattern[i])
			if err != nil {
				return err
			}

			// save the compiled regex pattern
			cmd.addCompiledRegex(compiledPattern)
		}
	}

	if cmd.PatternType == CMDTypeWord {
		// lowercase patterns
		for i := 0; i < len(cmd.Pattern); i++ {
			cmd.Pattern[i] = strings.ToLower(cmd.Pattern[i])
		}
	}

	// save the command
	st.commands = append(st.commands, cmd)

	return nil
}

func (st *CMDManager) Process(entry botio.InputEntry, outputs []botio.Output) error {

	cmd, pattern, handlerType, cmdContent, extraData, err := st.matchCMDs(st.commands, entry.Query, entry.InputMetadata, entry.GeneralMetadata)
	if err != nil {
		if st.errorChan != nil {
			st.errorChan <- err
		}
	} else {
		switch handlerType {
		case CMDHandlerTypeDefault:
			if cmd.Handler != nil {
				cmd.Handler(cmd, pattern, cmdContent, extraData, handlerType, entry.InputMetadata, entry.GeneralMetadata, outputs)
			}

		case CMDHandlerTypeError:
			if cmd.HandlerError != nil {
				cmd.HandlerError(cmd, pattern, cmdContent, extraData, handlerType, entry.InputMetadata, entry.GeneralMetadata, outputs)
			}

		case CMDHandlerTypeRestrictions:
			if cmd.HandlerRestrictions != nil {
				cmd.HandlerRestrictions(cmd, pattern, cmdContent, extraData, handlerType, entry.InputMetadata, entry.GeneralMetadata, outputs)
			}

		case CMDHandlerTypeParams:
			if cmd.HandlerParams != nil {
				cmd.HandlerParams(cmd, pattern, cmdContent, extraData, handlerType, entry.InputMetadata, entry.GeneralMetadata, outputs)
			}

		case CMDHandlerTypeParamsWrongType:
			if cmd.HandlerParamsWrongType != nil {
				cmd.HandlerParamsWrongType(cmd, pattern, cmdContent, extraData, handlerType, entry.InputMetadata, entry.GeneralMetadata, outputs)
			}

		case CMDHandlerTypeParamsMissing:
			if cmd.HandlerParamsMissing != nil {
				cmd.HandlerParamsMissing(cmd, pattern, cmdContent, extraData, handlerType, entry.InputMetadata, entry.GeneralMetadata, outputs)
			}

		case CMDHandlerTypeParamsExtra:
			if cmd.HandlerParamsExtra != nil {
				cmd.HandlerParamsExtra(cmd, pattern, cmdContent, extraData, handlerType, entry.InputMetadata, entry.GeneralMetadata, outputs)
			}

		default:
			log.Printf("Unknown CMDHandlerType: %v", handlerType)
		}
	}

	return nil
}

// matchCMDs finds for the best CMD (command) match.
// Returns CMD, bool, err
// CMD					==> best match command
// string				==> pattern that matches
// string				==> CMD handler type true, which handler type should be used
// string				==> content
// map[string][string]	==> extra information related to the cmd
// err		==> error
func (st *CMDManager) matchCMDs(cmds []CMD, content string, inputMetadata botio.Metadata, generalMetadata botio.Metadata) (CMD, string, string, string, map[string]interface{}, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return CMD{}, "", CMDHandlerTypeError, content, nil, NewCMDError(CMDErrorTypeNoCommand, fmt.Sprintf("no commands for '%v'", content))
	}

	// split the content into words separated by blank spaces
	words := strings.Split(content, " ")

	var (
		cmd2ret      CMD
		patternMatch string
	)

	for _, cmd := range cmds {
		match := false
		extraContent := ""

		switch cmd.PatternType {
		// regex
		case CMDTypeRegex:
			for i := 0; i < len(cmd.compiledRegex); i++ {
				match = cmd.compiledRegex[i].MatchString(content)

				if match {
					// saved the pattern that matches
					patternMatch = cmd.Pattern[i]
					break
					// TODO ::: get the extracontent from the regex
				}
			}

		// first word
		case CMDTypeWord:
			for i := 0; i < len(cmd.Pattern); i++ {
				match = strings.ToLower(words[0]) == cmd.Pattern[i]
				if match {
					// saved the pattern that matches
					patternMatch = cmd.Pattern[i]
					// extra content to keep parsing
					extraContent = content[len(words[0]):]
					break
				}
			}
		}

		if match {
			cmd2ret = cmd

			// check cmd restrictions
			if canExecute, errMessage := cmd2ret.CanExecute(inputMetadata, generalMetadata); !canExecute {
				return cmd2ret, patternMatch, CMDHandlerTypeRestrictions, content, map[string]interface{}{"message": errMessage}, nil
			}

			// get the extra content to parse (extra content == content - cmd)
			extraContent = strings.TrimSpace(extraContent)

			checkForParams := true

			// check for subCommands
			if len(cmd.SubCommands) > 0 {
				// do not check for params
				checkForParams = false
				extraCmd2ret, patternMatch2, useHandler, content2, extraData, err2 := st.matchCMDs(cmd.SubCommands, extraContent, inputMetadata, generalMetadata)
				if err2 == nil {
					return extraCmd2ret, patternMatch2, useHandler, content2, extraData, nil
				}
			}

			// check for params
			if checkForParams && len(cmd.Params) > 0 {
				totalParameters := 0

				for i := 0; i < len(cmd.Params); i++ {
					if i < len(words)-1 {
						totalParameters++

						cmd.Params[i].Value = words[i+1]
						// parameter type checking
						if !cmd.Params[i].isExpectedType() {
							// pass back the wrong param data
							return cmd2ret, patternMatch, CMDHandlerTypeParamsWrongType, content, map[string]interface{}{
								CMDExtraDataParametersIncorrectType: []CMDParam{cmd.Params[i]},
							}, nil
						}
					} else {
						// parameter required checking
						if cmd.Params[i].Required {
							// return insufficient parameters handler
							// pass back the missing param data
							return cmd2ret, patternMatch, CMDHandlerTypeParamsMissing, content, map[string]interface{}{
								CMDExtraDataParametersMissing: []CMDParam{cmd.Params[i]},
							}, nil
						}
					}
				}

				// more parameters than expected
				if len(words)-1 > len(cmd.Params) {
					// pass back the extra params
					extraParams := make([]CMDParam, 0)
					for c := len(cmd.Params) + 1; c < len(words); c++ {
						extraParams = append(extraParams, CMDParam{
							Value: words[c],
						})
					}

					// return extra parameters handler
					return cmd2ret, patternMatch, CMDHandlerTypeParamsExtra, content, map[string]interface{}{
						CMDExtraDataParametersExtra: extraParams,
					}, nil
				}

				// return CMDHandlerTypeParams ONLY if there are at least one parameter
				if totalParameters > 0 {
					return cmd2ret, patternMatch, CMDHandlerTypeParams, content, nil, nil
				}
			}

			// return CMD if there is no more content to parse
			if extraContent == "" {
				return cmd2ret, patternMatch, CMDHandlerTypeDefault, content, nil, nil
			}

			// use errorHandler instead of handler
			return cmd2ret, patternMatch, CMDHandlerTypeError, extraContent, nil, nil
		}

	}

	return CMD{}, "", CMDHandlerTypeError, content, nil, NewCMDError(CMDErrorTypeNoCommand, fmt.Sprintf("no commands for '%v'", content))
}
