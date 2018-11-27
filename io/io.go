package io

const (
	OutputMessageTypeDefault = "default"
	OutputMessageTypeFramed  = "framed"

	GeneralMetadataFieldUserID   = "userID"
	GeneralMetadataFieldUsername = "username"
	GeneralMetadataFieldWhere    = "where"
	GeneralMetadataFieldThread   = "thread"
)

type Input interface {
	Listen(chan InputEntry) error
}

type Output interface {
	Send(messageType string, message string, inputData Metadata, outputData Metadata) error
}

// InputEntry ==> Entry data from the input
type InputEntry struct {
	Origin          string
	Query           string
	InputMetadata   Metadata
	GeneralMetadata Metadata
}

/*
	General Metadata

	goal: Transform some particular fields into general fields
	fields:
		- UserID  	==> user ID
		- Username	==> username
		- Where		==> from where the entry was generated (chat room, for example)
		- Thread	==> Thread ID
*/

// Metadata ==> InputEntry's metadata
type Metadata map[string]interface{}
