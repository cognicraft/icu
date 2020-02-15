package icu

import "fmt"

type Tag string

const (
	TagEn = "en"
)

type MessageFormat string

func (m MessageFormat) String() string {
	return string(m)
}

func (m MessageFormat) Translate(tag Tag, ps ...Parameter) (string, error) {
	return Translate(tag, m, ps...)
}

func ParametersFrom(v interface{}) Parameters {
	var res []Parameter
	switch v := v.(type) {
	case map[string]string:
		for key, value := range v {
			res = append(res, P(key, value))
		}
	case map[string]interface{}:
		for key, value := range v {
			res = append(res, P(key, fmt.Sprint(value)))
		}
	}
	return res
}

type Parameters []Parameter

func (ps Parameters) Len() int           { return len(ps) }
func (ps Parameters) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
func (ps Parameters) Less(i, j int) bool { return ps[i].Name < ps[j].Name }

func P(name string, value interface{}) Parameter {
	return Parameter{
		Name:  name,
		Value: value,
	}
}

type Parameter struct {
	Name  string
	Value interface{}
}

func Translate(tag Tag, msg MessageFormat, ps ...Parameter) (string, error) {
	node, err := parse(string(msg))
	if err != nil {
		return "", err
	}
	return node.translate(newContext(tag, ps...)), nil
}
