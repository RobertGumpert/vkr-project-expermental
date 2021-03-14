package task




type ITask interface {
	SetType(Type)
	SetKey(string)
	SetExecutionStatus(bool)
	SetRunnableStatus(bool)
	SetDeferStatus(bool)
	SetResult(interface{})
	SetCustomFields(interface{})
	//
	GetType() Type
	GetKey() string
	GetExecutionStatus() bool
	GetDeferStatus() bool
	GetRunnableStatus() bool
	GetResult() interface{}
	GetCustomFields() interface{}
}