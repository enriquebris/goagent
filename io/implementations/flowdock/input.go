package flowdock

import (
	"github.com/enriquebris/goflowdock"
	botio "github.com/enriquebris/goagent/io"
	"github.com/fatih/structs"
)

const (
	Origin = "flowdock"
)

type FlowdockInput struct {
	streamURL     string
	streamManager *goflowdock.StreamManager
}

func NewFlowdockInput(authToken string, streamURL string) *FlowdockInput {
	ret := &FlowdockInput{}
	ret.initialize(authToken, streamURL)

	return ret
}

func (st *FlowdockInput) initialize(authToken string, streamURL string) {
	st.streamManager = goflowdock.NewStreamManager(authToken, nil)
	st.streamURL = streamURL
}

func (st *FlowdockInput) Listen(chInputEntry chan botio.InputEntry) error {
	return st.streamManager.Listen(st.streamURL, func(entry goflowdock.Entry) {
		// transform goflowdock.Entry into botio.InputEntry
		inputEntry := botio.InputEntry{
			Origin: Origin,
			Query:  entry.Content,
			// Entry to map[string]interface{}
			InputMetadata:   structs.Map(entry),
			GeneralMetadata: st.buildGeneralMetadata(entry),
		}

		chInputEntry <- inputEntry
	})
}

// buildGeneralMetadata converts the input metadata into the General Metadata
func (st *FlowdockInput) buildGeneralMetadata(entry goflowdock.Entry) botio.Metadata {
	return botio.Metadata{
		botio.GeneralMetadataFieldUserID: entry.User,
		botio.GeneralMetadataFieldWhere:  entry.Flow,
		botio.GeneralMetadataFieldThread: entry.ThreadID,
	}
}
