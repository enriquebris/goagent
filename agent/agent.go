package agent

import (
	"github.com/enriquebris/goagent/cmd"
	botio "github.com/enriquebris/goagent/io"
	"github.com/enriquebris/goworkerpool"
	//goworkerpool "github.com/enriquebris/goworkerpool-v2"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

const (
	// default number of concurrent workers to handle the requests
	defaultMaxConcurrentRequests = 10
)

type Agent struct {
	cmdManager *cmd.CMDManager
	input      botio.Input
	outputs    []botio.Output
	inputChann chan botio.InputEntry
	log        *logging.Logger
	workerpool *goworkerpool.Pool
	isListening bool
}

func NewAgent(log *logging.Logger, maxConcurrentRequests int) *Agent {
	ret := &Agent{}
	ret.initialize(log, maxConcurrentRequests)

	return ret
}

func (st *Agent) initialize(log *logging.Logger, maxConcurrentRequests int) {
	st.cmdManager = cmd.NewCMDManager(nil)
	st.inputChann = make(chan botio.InputEntry, 500)
	st.outputs = make([]botio.Output, 0)
	// goworkerpool
	st.initializeWorkerPool(maxConcurrentRequests)
}

// initializeWorkerPool initializes workerpool to handle concurrent requests
func (st *Agent) initializeWorkerPool(totalWorkers int) {
	if totalWorkers <= 0 {
		totalWorkers = defaultMaxConcurrentRequests
	}
	st.workerpool = goworkerpool.NewPool(totalWorkers, 1000, false)

	// set the main handler function
	st.workerpool.SetWorkerFunc(func (data interface{}) bool {
		// cast the job as a io.InputEntry
		if entry, ok := data.(botio.InputEntry); !ok {
			st.log.Errorf("Enqueued job is not a io.InputEntry: '%v'", data)
		} else {
			// process the entry
			if err := st.cmdManager.Process(entry, st.outputs); err != nil {
				st.log.Errorf("st.cmdManager.Process: '%v'", err.Error())
			}
		}

		return true
	})
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

func (st *Agent) Listen() error {
	// only one listener could be alive at the same time
	if st.isListening {
		return errors.New("Agent is already listening")
	}
	st.isListening = true
	defer func() {st.isListening = false}()

	// spin up workers
	if err := st.workerpool.StartWorkers(); err != nil {
		return err
	}

	// start listening
	go st.input.Listen(st.inputChann)

	keepWorking := true
	for keepWorking {
		select {
		case entry, ok := <-st.inputChann:
			if !ok {
				// the channel is closed: exit
				keepWorking = false
				// process all enqueued input entries and kill the workers
				st.workerpool.LateKillAllWorkers()

				break
			}

			// enqueue the input entry to be processed by a worker
			st.workerpool.AddTask(entry)
		}
	}

	// wait until all workers are down
	st.workerpool.Wait()

	st.log.Notice("Agent.Listener is done")
	return nil
}
