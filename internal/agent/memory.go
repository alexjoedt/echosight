package agent

import "encoding/json"

type RAMResult struct {
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

func (m *RAMResult) Bytes() []byte {
	data, _ := json.MarshalIndent(m, "", "\t")
	return data
}
