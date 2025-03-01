package model

type Request struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
	Type  string
	Group string `json:"Group"`
}

type Response struct {
	Value string `json:"value"`
	Error string `json:"error"`
}
