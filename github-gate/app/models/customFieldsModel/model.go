package customFieldsModel

import "github.com/RobertGumpert/gotasker/itask"

type Model struct {
	TaskType itask.Type
	Fields   interface{}
}

func (c *Model) GetFields() interface{} {
	return c.Fields
}

func (c *Model) GetTaskType() itask.Type {
	return c.TaskType
}
