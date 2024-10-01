package service

const HelloServiceName = "my-hello-service"

const NullSyncOp = "nullSyncOp"
const NullAsyncOp = "nullAsyncOp"

// Echo operation
const EchoOperationName = "echo"

type EchoInput struct {
	Message string
}

type EchoOutput EchoInput

// Hello operation
const HelloOperationName = "say-hello"
const HelloOperation2Name = "say-hello-2"

type Language string

const (
	EN Language = "en"
	FR Language = "fr"
	DE Language = "de"
	ES Language = "es"
	TR Language = "tr"
)

type HelloInput struct {
	Name     string
	Language Language
}

type HelloOutput struct {
	Message string
}
