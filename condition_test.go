package main

import "testing"

func TestEvalEq(t *testing.T) {
	testEval(t, false, "== 5", 1, false)
	testEval(t, true, "== 5", 5, false)
	testEval(t, false, "== null", 5, false)
	testEval(t, true, "== null", 0, true)
}

func TestEvalGe(t *testing.T) {
	testEval(t, false, ">= 5", 4, false)
	testEval(t, true, ">= 5", 5, false)
	testEval(t, true, ">= 5", 6, false)
	testEval(t, false, ">= 5", 0, true)
}

func TestEvalGt(t *testing.T) {
	testEval(t, false, "> 5", 5, false)
	testEval(t, true, "> 5", 6, false)
	testEval(t, false, "> 5", 0, true)
}

func TestEvalLe(t *testing.T) {
	testEval(t, false, "<= 5", 6, false)
	testEval(t, true, "<= 5", 5, false)
	testEval(t, true, "<= 5", 4, false)
	testEval(t, false, "<= 5", 0, true)
}

func TestEvalLt(t *testing.T) {
	testEval(t, false, "< 5", 5, false)
	testEval(t, true, "< 5", 4, false)
	testEval(t, false, "< 5", 0, true)
}

func TestEvalNe(t *testing.T) {
	testEval(t, false, "!= 5", 5, false)
	testEval(t, true, "!= 5", 4, false)
	testEval(t, true, "!= null", 5, false)
	testEval(t, false, "!= null", 0, true)
}

func testEval(t *testing.T, exp bool, cond string, val float64, null bool) {
	c, err := ParseCondition(cond)
	if err != nil {
		t.Fatal(err)
	}
	if c.Eval(val, null) != exp {
		t.Fatal()
	}
}
