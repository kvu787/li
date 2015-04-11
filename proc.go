package main

type specialForm func(args []interface{}, env map[string]interface{}) (interface{}, error)

type variadicProc struct {
	param string
	body  func(env map[string]interface{}) (interface{}, error)
}

type proc struct {
	params []string
	body   func(env map[string]interface{}) (interface{}, error)
}
