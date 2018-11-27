package main

import (
	"github.com/enriquebris/goagent/agent"
	"github.com/enriquebris/goagent/handler"
	"github.com/enriquebris/goagent/io/implementations/flowdock"
	"github.com/enriquebris/goagent/main/cmds/example"
)

func main() {
	flowdockIOInput := flowdock.NewFlowdockInput("replaceByYourToken", "https://stream.flowdock.com/flows?filter=")
	flowdockIOOutput := flowdock.NewFlowdockOutput("replaceByYourToken", "replaceByYourOrganization", agentName)

	initializeLogger("error.log")

	// common handler helper
	commonHandler := handler.NewCommon(log)

	myAgent := agent.NewAgent(log)
	// set the input / output
	myAgent.SetInput(flowdockIOInput)
	myAgent.AddOutput(flowdockIOOutput)
	// add the main command / skill
	myAgent.AddCMD(example.GetMainCMD(commonHandler))

	// start listening
	myAgent.Listen()
}
