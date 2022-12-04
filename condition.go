package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Op string

const (
	OpEq Op = "=="
	OpGe Op = ">="
	OpGt Op = ">"
	OpLe Op = "<="
	OpLt Op = "<"
	OpNe Op = "!="
)

type Condition struct {
	op    Op
	null  bool
	value float64
}

func ParseCondition(s string) (*Condition, error) {
	pts := strings.Split(s, " ")
	if len(pts) != 2 {
		return nil, errors.New("invalid syntax")
	}

	c := &Condition{}
	op, err := parseOp(strings.TrimSpace(pts[0]))
	if err != nil {
		return nil, err
	}
	c.op = op

	v := strings.TrimSpace(pts[1])
	if v == "null" {
		c.null = true
		if op != OpEq && op != OpNe {
			m := "null supports only == and != operations"
			return nil, errors.New(m)
		}
	} else {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		c.value = f
	}

	return c, nil
}

func (c *Condition) Eval(val float64, null bool) bool {
	switch c.op {
	case OpEq:
		return val == c.value && null == c.null
	case OpGe:
		return !null && val >= c.value
	case OpGt:
		return !null && val > c.value
	case OpLe:
		return !null && val <= c.value
	case OpLt:
		return !null && val < c.value
	case OpNe:
		return val != c.value || null != c.null
	default:
		panic("unsupported operation")
	}
}

func (c *Condition) String() string {
	if c.null {
		return fmt.Sprintf("%s null", c.op)
	} else {
		return fmt.Sprintf("%s %f", c.op, c.value)
	}
}
func parseOp(s string) (Op, error) {
	switch s {
	case string(OpEq):
		return OpEq, nil
	case string(OpGe):
		return OpGe, nil
	case string(OpGt):
		return OpGt, nil
	case string(OpLe):
		return OpLe, nil
	case string(OpLt):
		return OpLt, nil
	case string(OpNe):
		return OpNe, nil
	default:
		return "", fmt.Errorf("unsupported operation: %s", s)
	}
}
