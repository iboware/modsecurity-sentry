package main

type ModsecurityLogEntry struct {
	Transaction Transaction `json:"transaction"`
}

type Transaction struct {
	ClientIP   string    `json:"client_ip"`
	TimeStamp  string    `json:"time_stamp"`
	ServerID   string    `json:"server_id"`
	ClientPort int       `json:"client_port"`
	HostIP     string    `json:"host_ip"`
	HostPort   int       `json:"host_port"`
	UniqueID   string    `json:"unique_id"`
	Request    Request   `json:"request"`
	Response   Response  `json:"response"`
	Producer   Producer  `json:"producer"`
	Messages   []Message `json:"messages"`
}

type Request struct {
	Method      string         `json:"method"`
	HTTPVersion float64        `json:"http_version"`
	URI         string         `json:"uri"`
	Headers     RequestHeaders `json:"headers"`
}
type RequestHeaders struct {
	Host      string `json:"host"`
	Accept    string `json:"accept"`
	UserAgent string `json:"user-agent"`
}

type Response struct {
	Body     string          `json:"body"`
	HTTPCode int             `json:"http_code"`
	Headers  ResponseHeaders `json:"headers"`
}

type ResponseHeaders struct {
	Server                  string `json:"Server"`
	Date                    string `json:"Date"`
	ContentLength           string `json:"Content-Length"`
	ContentType             string `json:"Content-Type"`
	Connection              string `json:"Connection"`
	StrictTransportSecurity string `json:"Strict-Transport-Security"`
}

type Message struct {
	Message string         `json:"message"`
	Details MessageDetails `json:"details"`
}

type MessageDetails struct {
	Match      string        `json:"match"`
	Reference  string        `json:"reference"`
	RuleID     string        `json:"ruleId"`
	File       string        `json:"file"`
	LineNumber string        `json:"lineNumber"`
	Data       string        `json:"data"`
	Severity   string        `json:"severity"`
	Ver        string        `json:"ver"`
	Rev        string        `json:"rev"`
	Tags       []interface{} `json:"tags"`
	Maturity   string        `json:"maturity"`
	Accuracy   string        `json:"accuracy"`
}

type Producer struct {
	Modsecurity    string   `json:"modsecurity"`
	Connector      string   `json:"connector"`
	SecrulesEngine string   `json:"secrules_engine"`
	Components     []string `json:"components"`
}
