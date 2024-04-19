package log

// TestLogSource type
type TestLogSource struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// TestLog type
type TestLog struct {
	Time    string        `json:"time"`
	Level   string        `json:"level"`
	Source  TestLogSource `json:"source"`
	Message string        `json:"msg"`
}
