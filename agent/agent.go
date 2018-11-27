package agent

import (
	"github.com/enriquebris/goagent/cmd"
	botio "github.com/enriquebris/goagent/io"
	"github.com/op/go-logging"
)

type Agent struct {
	cmdManager *cmd.CMDManager
	input      botio.Input
	outputs    []botio.Output
	inputChann chan botio.InputEntry
	log        *logging.Logger
}

func NewAgent(log *logging.Logger) *Agent {
	ret := &Agent{}
	ret.initialize(log)

	return ret
}

func (st *Agent) initialize(log *logging.Logger) {
	st.cmdManager = cmd.NewCMDManager(nil)
	st.inputChann = make(chan botio.InputEntry, 500)
	st.outputs = make([]botio.Output, 0)
}

// SetInput sets the Input
func (st *Agent) SetInput(input botio.Input) {
	st.input = input
}

// AddOutput adds a io.Output
func (st *Agent) AddOutput(output botio.Output) {
	st.outputs = append(st.outputs, output)
}

// AddCMD adds a new command
func (st *Agent) AddCMD(cmd cmd.CMD) error {
	return st.cmdManager.AddCommand(cmd)
}

func (st *Agent) Listen() {
	go st.input.Listen(st.inputChann)

	keepWorking := true
	for keepWorking {
		select {
		case entry, ok := <-st.inputChann:
			if !ok {
				// the channel is closed: exit
				keepWorking = false
				break
			}

			if err := st.cmdManager.Process(entry, st.outputs); err != nil {
				st.log.Errorf("st.cmdManager.Process: '%v'", err.Error())
			}
		}
	}

	st.log.Notice("Agent.Listener is done")
}
