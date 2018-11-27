package io

// MergeMetadata merges two Metadata structs. Values from param2 (for duplicated keys) will overwrite values from param1.
func MergeMetadata(param1 Metadata, param2 Metadata) Metadata {
	ret := Metadata{}
	// iterate over param1
	for key, value := range param1 {
		ret[key] = value
		if v, ok := param2[key]; ok {
			ret[key] = v
		}
	}

	// iterate over param2
	for key, value := range param2 {
		if _, ok := param1[key]; !ok {
			ret[key] = value
		}
	}

	return ret
}
