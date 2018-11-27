package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	botio "github.com/enriquebris/goagent/io"
)

const (
	CMDTypeRegex = "regex"
	CMDTypeWord  = "word"

	CMDHandlerTypeDefault         = "handler.default"
	CMDHandlerTypeError           = "handler.error"
	CMDHandlerTypeParams          = "handler.params"
	CMDHandlerTypeRestrictions    = "handler.restrictions"
	CMDHandlerTypeParamsWrongType = "handler.params.wrong.type"
	CMDHandlerTypeParamsMissing   = "handler.params.missing"
	CMDHandlerTypeParamsExtra     = "handler.params.extra"

	CMDExtraDataParametersIncorrectType = "extra.data.parameters.incorrect.type"
	CMDExtraDataParametersMissing       = "extra.data.parameters.missing"
	CMDExtraDataParametersExtra         = "extra.data.parameters.extra"

	// CMD restriction, includes only the listed elements
	CMDRestrictionConceptInclude = "restriction.include"
	// CMD restriction, excludes the listed elements, allows all others
	CMDRestrictionConceptExclude = "restriction.exclude"
)

// CMDHandler is the function handler
//
// Parameters:
//
// cmd CMD							==> CMD that matched
// pattern string					==> exact CMD pattern that matched
// cmdContent string				==> Content to parse (original content - cmd)
// metadata	botio.Metadata			==> Extra data related to the CMD
// handlerType string				==> Indicates which handler type will be used and gives some extra data (parameters: ok, missing, extra)
// inputMetadata botio.Metadata		==> Metadata from the input (exactly as it comes from the origin)
// generalMetadata botio.Metadata	==> General metadata (some values transformed into general fields: user, where, ...)
// outputs botio.Output				==> Output interface
type CMDHandler func(cmd CMD, pattern string, cmdContent string, metadata botio.Metadata, handlerType string, inputMetadata botio.Metadata, generalMetadata botio.Metadata, outputs []botio.Output)

// ***********************************************************************************************
// **  CMD  **************************************************************************************
// ***********************************************************************************************

// CMD command
type CMD struct {
	PatternType            string
	Pattern                []string
	Description            string
	Required               bool
	compiledRegex          []*regexp.Regexp
	Handler                CMDHandler
	HandlerError           CMDHandler
	HandlerRestrictions    CMDHandler
	HandlerParams          CMDHandler
	HandlerParamsWrongType CMDHandler
	HandlerParamsMissing   CMDHandler
	HandlerParamsExtra     CMDHandler
	SubCommands            []CMD
	Params                 []CMDParam
	params2map             bool
	mpParams               map[string]CMDParam
	GeneralRestrictions    []Restriction
}

type Restriction struct {
	ID      string
	Concept string
	Field   string
	Data    []interface{}
}

// AddSubCMD adds a sub command
func (st *CMD) AddSubCMD(sub CMD) {
	st.SubCommands = append(st.SubCommands, sub)
}

// addCompiledRegex adds a new compiled regex (that matches a pattern)
func (st *CMD) addCompiledRegex(regx *regexp.Regexp) {
	if st.compiledRegex == nil {
		st.compiledRegex = make([]*regexp.Regexp, 0)
	}

	st.compiledRegex = append(st.compiledRegex, regx)
}

// GetParamByID returns a CMDParam given its ID
func (st *CMD) GetParamByID(id string) (CMDParam, error) {
	if !st.params2map {
		st.convertParamsToMap()
	}

	if param, ok := st.mpParams[id]; ok {
		return param, nil
	}

	return CMDParam{}, fmt.Errorf("No param '%v'", id)
}

// GetParamByIndex returns a CMDParam given its index
func (st *CMD) GetParamByIndex(index int) (CMDParam, error) {
	if index >= len(st.Params) {
		return CMDParam{}, fmt.Errorf("Index out of bounds: %v", index)
	}

	return st.Params[index], nil
}

// convertParamsToMap converts Params ([]CMDParam) to a map[string]CMDParam
func (st *CMD) convertParamsToMap() {
	st.mpParams = make(map[string]CMDParam)
	for i := 0; i < len(st.Params); i++ {
		st.mpParams[st.Params[i].ID] = st.Params[i]
	}

	st.params2map = true
}

// CanExecute returns true whether the CMD can be executed. It verifies that the CMD meets all restrictions.
func (st *CMD) CanExecute(inputMetadata botio.Metadata, generalMetadata botio.Metadata) (bool, string) {
	// check general restrictions
	if isOK, message := st.canExecuteByRestrictions(st.GeneralRestrictions, generalMetadata); !isOK {
		return isOK, message
	}

	return true, ""
}

// canExecuteByRestrictions returns true whether the cmd can be executed because it meets all restrictions
func (st *CMD) canExecuteByRestrictions(restrictions []Restriction, metadata botio.Metadata) (bool, string) {
	for i := 0; i < len(restrictions); i++ {
		fieldValue, fieldExists := metadata[restrictions[i].Field]
		if !fieldExists {
			return false, fmt.Sprintf("Missing metadata field: '%v'. Restriction '%v' cannot be checked.", restrictions[i].Field, restrictions[i].ID)
		}

		switch restrictions[i].Concept {
		// metadata value must be equal to one of the provided values (restrictions[i].Data)
		case CMDRestrictionConceptInclude:
			includeOK := true
			for c := 0; c < len(restrictions[i].Data); c++ {
				if restrictions[i].Data[c] == fieldValue {
					includeOK = true
					break
				} else {
					includeOK = false
				}
			}

			if !includeOK {
				return false, fmt.Sprintf("Metadata['%v'] does not meet restriction '%v' with concept '%v'.\nCurrent value: '%v'", restrictions[i].Field, restrictions[i].ID, restrictions[i].Concept, fieldValue)
			}

		// metadata value must not be equal to any of the provided values (restrictions[i].Data)
		case CMDRestrictionConceptExclude:
			excludeOK := true
			for c := 0; c < len(restrictions[i].Data); c++ {
				if restrictions[i].Data[c] == fieldValue {
					excludeOK = false
					break
				}
			}

			if !excludeOK {
				return false, fmt.Sprintf("Metadata['%v'] does not meet restriction '%v' with concept '%v'.\nCurrent value: '%v'", restrictions[i].Field, restrictions[i].ID, restrictions[i].Concept, fieldValue)
			}
		}
	}

	return true, ""
}

// ***********************************************************************************************
// **  CMDParam  *********************************************************************************
// ***********************************************************************************************

const (
	CMDParamTypeInt    = "int"
	CMDParamTypeString = "string"
)

type CMDParam struct {
	ID          string
	Description string
	Type        string
	Required    bool
	Value       string
}

// isExpectedType verifies whether the param's type is the expected.
func (st *CMDParam) isExpectedType() bool {
	switch st.Type {
	case "":
		return true

	case CMDParamTypeString:
		// any type could be converted to string
		return true

	case CMDParamTypeInt:
		_, err := st.GetIntValue()
		return err == nil
	}

	return false
}

// GetIntValue returns the int value
func (st *CMDParam) GetIntValue() (int, error) {
	return strconv.Atoi(st.Value)
}

// ***********************************************************************************************
// **  CMDError  *********************************************************************************
// ***********************************************************************************************

type CMDError struct {
	errorMessage string
	errorType    string
}

const (
	CMDErrorTypeNoCommand = "noCommand"
)

func NewCMDError(errorType string, errorMessage string) *CMDError {
	return &CMDError{
		errorType:    errorType,
		errorMessage: errorMessage,
	}
}

func (st *CMDError) Error() string {
	return st.errorMessage
}

func (st *CMDError) GetType() string {
	return st.errorType
}
