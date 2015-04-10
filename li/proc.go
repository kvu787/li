package li

type specialForm func(args []interface{}, env map[string]interface{}) interface{}

type variadicProc struct {
	param string
	body  func(env map[string]interface{}) interface{}
}

type proc struct {
	params []string
	body   func(env map[string]interface{}) interface{}
}
