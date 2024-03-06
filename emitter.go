package cewrap

type Emitter interface {}

type CEEmitter struct {
	
}

func New() *CEEmitter {
	return nil
}

// Emits event with the fields set:
// 	- ID
//	- Time
// 	- Source
// 	- Type
// 	- Subject
// 	- Data
func (e *CEEmitter) Emit() error {
	return nil
}

