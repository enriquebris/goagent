package handler

import (
	"fmt"
	"sort"

	"github.com/enriquebris/goagent/message"
	"github.com/enriquebris/goagent/cmd"
)

// getHelpForCMD returns a help string for the given command
func getHelpForCMD(command cmd.CMD, tagHelp string) string {
	ret := ""
	addHelpCommand := true

	if len(command.SubCommands) == 0 {
		command.SubCommands = []cmd.CMD{command}
		addHelpCommand = false
	}

	// sort subcommands alphabetically (sort by first subcommand)
	sort.Sort(cmdSlice(command.SubCommands))

	// get the longest command
	maxCommandLen := 0
	for i := 0; i < len(command.SubCommands); i++ {
		// build an combined string with all patterns
		tmpPattern := getPattern(command.SubCommands[i].Pattern)
		if maxCommandLen < len(tmpPattern) {
			maxCommandLen = len(tmpPattern)
		}
	}

	spaces4params := message.GetSpaces(maxCommandLen)
	// add all commands
	for i := 0; i < len(command.SubCommands); i++ {
		// build an combined string with all patterns
		tmpPattern := getPattern(command.SubCommands[i].Pattern)

		ret += fmt.Sprintf(
			"\n\t%v%v\t\t%v",
			tmpPattern, // command
			message.GetSpaces(maxCommandLen-len(tmpPattern)), // extra blank spaces to match all columns
			command.SubCommands[i].Description,               // command description
		)

		optionalTmp := "optional"
		typeTmp := ""
		// parameters
		for c := 0; c < len(command.SubCommands[i].Params); c++ {
			// optional / required
			if command.SubCommands[i].Params[c].Required {
				optionalTmp = "required"
			}
			// type
			typeTmp = command.SubCommands[i].Params[c].Type
			if typeTmp == "" {
				typeTmp = "string"
			}

			ret += fmt.Sprintf("\n%v\t\t\t %v (%v - %v) ==> %v", spaces4params, command.SubCommands[i].Params[c].ID, typeTmp, optionalTmp, command.SubCommands[i].Params[c].Description)
		}
	}

	if addHelpCommand {
		ret = fmt.Sprintf(
			"\n\t%v%v\t\t%v",
			tagHelp, // command
			message.GetSpaces(maxCommandLen-len(tagHelp)),                                            // extra blank spaces to match all columns
			fmt.Sprintf("Add %v after any command to list its subcommands and description", tagHelp), // command description
		) + ret
	}

	return ret
}

// getPattern returns an combined string with all slice's values
func getPattern(patterns []string) string {
	ret := ""

	for i := 0; i < len(patterns); i++ {
		ret += fmt.Sprintf("%v / ", patterns[i])
	}

	if len(ret) > 0 {
		ret = ret[:len(ret)-2]
	}

	return ret
}

// removeSliceElement returns the slice after remove the given element
func removeSliceElement(sl []string, element string) []string {
	pos := -1
	for i := 0; i < len(sl); i++ {
		if sl[i] == element {
			pos = i
			break
		}
	}

	if pos >= 0 {
		return append(sl[:pos], sl[pos+1:]...)
	}

	return sl
}

// *********************************************************************************************
// ** cmdSlice ==> Sorts alphabetically a string slice
// *********************************************************************************************

type cmdSlice []cmd.CMD

func (st cmdSlice) Len() int {
	return len(st)
}
func (st cmdSlice) Swap(i, j int) {
	st[i], st[j] = st[j], st[i]
}
func (st cmdSlice) Less(i, j int) bool {
	return st[i].Pattern[0] < st[j].Pattern[0]
}
