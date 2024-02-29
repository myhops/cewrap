package cewrap

type Emitter struct {

}

func New() *Emitter {
	return nil
}

// Emits event with the fields set:
// 	- ID
//	- Time
// 	- Source
// 	- Type
// 	- Subject
// 	- Data
func (e *Emitter) Emit() error {
	return nil
}

