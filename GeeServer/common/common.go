package common

import (
	"fmt"
)

type NodeEntry struct {
	Key string `json:"key"`
	IP  string `json:"ip"`
}

func GetKetList(list []NodeEntry) ([]string, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("no node to add")
	}
	result := make([]string, len(list))
	for _, v := range list {
		result = append(result, v.Key)
	}
	return result, nil
}
