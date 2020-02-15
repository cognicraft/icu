package icu

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type stack struct {
	nodes []node
}

func (s *stack) push(t node) {
	s.nodes = append(s.nodes, t)
}

func (s *stack) pop() node {
	x := s.nodes[len(s.nodes)-1]
	s.nodes = s.nodes[:len(s.nodes)-1]
	return x
}

func (s *stack) top() node {
	if len(s.nodes) > 0 {
		return s.nodes[len(s.nodes)-1]
	}
	return nil
}

func (s *stack) empty() bool {
	return len(s.nodes) == 0
}

type tree struct {
	root nodeMessage
}

type node interface {
	translate(ctx *context) string
}

func newContext(tag Tag, ps ...Parameter) *context {
	idx := strings.Index(string(tag), "-")
	t := tag
	if idx > 0 {
		t = tag[:idx]
	}
	ctx := &context{
		tag:    t,
		values: map[string]interface{}{},
	}
	for _, p := range ps {
		ctx.values[p.Name] = p.Value
	}
	return ctx
}

type context struct {
	tag    Tag
	values map[string]interface{}
}

type nodeSelector string

func (n nodeSelector) translate(ctx *context) string {
	return ""
}

type nodeStartAction struct{}

func (n nodeStartAction) translate(ctx *context) string {
	return ""
}

type nodeStartMessage struct{}

func (n nodeStartMessage) translate(ctx *context) string {
	return ""
}

type nodeMessage []node

func (n nodeMessage) translate(ctx *context) string {
	buf := bytes.Buffer{}
	for _, c := range n {
		buf.WriteString(c.translate(ctx))
	}
	return buf.String()
}

type nodeText string

func (n nodeText) translate(ctx *context) string {
	return string(n)
}

type nodeHash struct{}

func (n nodeHash) translate(ctx *context) string {
	var vals []string
	for k, v := range ctx.values {
		if !strings.HasPrefix(k, "$") {
			vals = append(vals, fmt.Sprint(v))
		}
	}
	if len(vals) == 1 {
		return vals[0]
	}
	return "#"
}

type nodeQuotedText string

func (n nodeQuotedText) translate(ctx *context) string {
	return string(n)
}

type nodeFormatPlaceholder struct {
	key string
}

func (n nodeFormatPlaceholder) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

type nodeFormatNumber struct {
	key   string
	style string
}

func (n nodeFormatNumber) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	if n.style == "" {
		n.style = "%v"
	}
	return fmt.Sprintf(n.style, v)
}

type nodeFormatDate struct {
	key   string
	style string
}

func (n nodeFormatDate) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}

	format, ok := ctx.values["$date-format"]
	if !ok {
		format = time.RFC3339
	}
	date := v.(time.Time)
	return fmt.Sprintf("%v", date.Format(format.(string)))
}

type nodeFormatTime struct {
	key   string
	style string
}

func (n nodeFormatTime) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

type nodeFormatOrdinal struct {
	key   string
	style string
}

func (n nodeFormatOrdinal) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

type nodeFormatDuration struct {
	key   string
	style string
}

func (n nodeFormatDuration) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

type nodeFormatSpellout struct {
	key   string
	style string
}

func (n nodeFormatSpellout) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

const (
	zero  = "zero"
	one   = "one"
	two   = "two"
	few   = "few"
	many  = "many"
	other = "other"
)

func cardinalToCategory(tag Tag, n int) string {
	result := ""
	switch tag {
	case "de", "es", "bg", "pt", "en", "it":
		if n == 1 {
			result = one
		} else {
			result = other
		}
	case "zh":
		result = other
	}

	return result
}

func ordinalToCategory(tag Tag, n int) string {
	result := ""
	switch tag {
	case "es", "bg", "pt", "zh", "de":
		result = other
	case "it":
		if n == 8 || n == 11 || n == 80 || n == 800 {
			result = many
		} else {
			result = other
		}
	case "en":
		t1 := n % 10
		t2 := n % 100
		if t1 == 1 && t2 != 11 {
			result = one
		} else if t1 == 2 && t2 != 12 {
			result = two
		} else if t1 == 3 && t2 != 13 {
			result = few
		} else {
			result = other
		}
	}

	return result
}

type nodeFormatPlural struct {
	key    string
	offset int
	cases  map[string]nodeMessage
}

func (n nodeFormatPlural) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}

	sv := fmt.Sprintf("=%v", v)
	c, ok := n.cases[sv]

	intV, _ := v.(int)
	nv := intV - n.offset

	if !ok {
		cat := cardinalToCategory(ctx.tag, nv)
		c, ok = n.cases[cat]
		if !ok {
			c, ok = n.cases[other]
			if !ok {
				return ""
			}
		}
	}

	ctx.values[n.key] = nv
	return c.translate(ctx)
}

type nodeFormatSelectOrdinal struct {
	key    string
	offset int
	cases  map[string]nodeMessage
}

func (n nodeFormatSelectOrdinal) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}

	sv := fmt.Sprintf("=%v", v)
	c, ok := n.cases[sv]

	intV, _ := v.(int)
	nv := intV - n.offset

	if !ok {
		cat := ordinalToCategory(ctx.tag, nv)
		c, ok = n.cases[cat]
		if !ok {
			c, ok = n.cases[other]
			if !ok {
				return ""
			}
		}
	}

	ctx.values[n.key] = nv
	return c.translate(ctx)
}

type nodeFormatSelect struct {
	key   string
	cases map[string]nodeMessage
}

func (n nodeFormatSelect) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	sv := fmt.Sprintf("%v", v)
	c, ok := n.cases[sv]
	if !ok {
		c, ok = n.cases[other]
		if !ok {
			return ""
		}
	}
	return c.translate(ctx)
}

type nodeFormatCustom struct {
	key    string
	custom string
	args   []string
}

func (n nodeFormatCustom) translate(ctx *context) string {
	v, ok := ctx.values[n.key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s(%v,%v)", n.custom, v, n.args)
}

func parse(input string) (nodeMessage, error) {
	stack := &stack{}
	lex := newLexer(input)
	for {
		t := lex.nextToken()
		switch t.cat {
		case tokenEOF:
			msg := nodeMessage{}
			for !stack.empty() {
				msg = append(nodeMessage{stack.pop()}, msg...)
			}
			return msg, nil
		case tokenStartAction:
			stack.push(nodeStartAction{})
		case tokenStartMessage:
			stack.push(nodeStartMessage{})
		case tokenText:
			stack.push(nodeText(t.val))
		case tokenQuotedText:
			stack.push(nodeQuotedText(t.val))
		case tokenHash:
			stack.push(nodeHash{})
		case tokenIdentifier:
			switch last := stack.top().(type) {
			case nodeStartAction:
				stack.push(nodeFormatPlaceholder{
					key: t.val,
				})
			case nodeFormatPlaceholder:
				stack.pop()
				typ := t.val
				switch typ {
				case "number":
					stack.push(nodeFormatNumber{
						key: last.key,
					})
				case "date":
					stack.push(nodeFormatDate{
						key: last.key,
					})
				case "time":
					stack.push(nodeFormatTime{
						key: last.key,
					})
				case "ordinal":
					stack.push(nodeFormatOrdinal{
						key: last.key,
					})
				case "duration":
					stack.push(nodeFormatDuration{
						key: last.key,
					})
				case "spellout":
					stack.push(nodeFormatSpellout{
						key: last.key,
					})
				case "plural":
					stack.push(nodeFormatPlural{
						key:    last.key,
						offset: 0,
						cases:  map[string]nodeMessage{},
					})
				case "selectordinal":
					stack.push(nodeFormatSelectOrdinal{
						key:    last.key,
						offset: 0,
						cases:  map[string]nodeMessage{},
					})
				case "select":
					stack.push(nodeFormatSelect{
						key:   last.key,
						cases: map[string]nodeMessage{},
					})
				default:
					stack.push(nodeFormatCustom{
						key:    last.key,
						custom: typ,
					})
				}
			case nodeFormatNumber:
				stack.pop()
				last.style = t.val
				stack.push(last)
			case nodeFormatDate:
				stack.pop()
				last.style = t.val
				stack.push(last)
			case nodeFormatTime:
				stack.pop()
				last.style = t.val
				stack.push(last)
			case nodeFormatOrdinal:
				stack.pop()
				last.style = t.val
				stack.push(last)
			case nodeFormatDuration:
				stack.pop()
				last.style = t.val
				stack.push(last)
			case nodeFormatSpellout:
				stack.pop()
				last.style = t.val
				stack.push(last)
			case nodeFormatSelectOrdinal:
				stack.pop()
				// NOTE: Not sure if this is a thing
				if t.val == "offset" {
					n := lex.nextToken()
					last.offset, _ = strconv.Atoi(n.val[1:])
					stack.push(last)
				} else {
					last.cases[t.val] = nodeMessage{}
					stack.push(last)
					stack.push(nodeSelector(t.val))
					stack.push(nodeStartMessage{})

				consume00:
					for {
						n := lex.nextToken()
						if n.cat == tokenStartMessage {
							break consume00
						}
					}
				}
			case nodeFormatPlural:
				stack.pop()
				if t.val == "offset" {
					n := lex.nextToken()
					last.offset, _ = strconv.Atoi(n.val[1:])
					stack.push(last)
				} else {
					last.cases[t.val] = nodeMessage{}
					stack.push(last)
					stack.push(nodeSelector(t.val))
					stack.push(nodeStartMessage{})

				consume0:
					for {
						n := lex.nextToken()
						if n.cat == tokenStartMessage {
							break consume0
						}
					}
				}
			case nodeFormatSelect:
				stack.pop()
				last.cases[t.val] = nodeMessage{}
				stack.push(last)
				stack.push(nodeSelector(t.val))
				stack.push(nodeStartMessage{})
			consume:
				for {
					n := lex.nextToken()
					if n.cat == tokenStartMessage {
						break consume
					}
				}
			case nodeFormatCustom:
				stack.pop()
				last.args = append(last.args, t.val)
				stack.push(last)
			}
		case tokenEndMessage:
			msg := nodeMessage{}
		pop1:
			for {
				switch n := stack.pop().(type) {
				case nodeStartMessage:
					sel := stack.pop().(nodeSelector)
					switch comp := stack.pop().(type) {
					case nodeFormatPlural:
						comp.cases[string(sel)] = msg
						stack.push(comp)
					case nodeFormatSelectOrdinal:
						comp.cases[string(sel)] = msg
						stack.push(comp)
					case nodeFormatSelect:
						comp.cases[string(sel)] = msg
						stack.push(comp)
					}
					break pop1
				default:
					msg = append(nodeMessage{n}, msg...)
				}
			}
		case tokenEndAction:
			msg := nodeMessage{}
		pop2:
			for {
				switch n := stack.pop().(type) {
				case nodeStartAction:
					stack.push(msg)
					break pop2
				default:
					msg = append(nodeMessage{n}, msg...)
				}
			}
		default:
			continue
		}
	}
}
