package fxpratt_test

import (
	"bufio"
	"errors"
	"fmt"
	"fxlex"
	"fxpratt"
	"math"
	"os"
	"strings"
	"testing"
)

type progExamp struct {
	input string
	isBad bool
	val   float64
}

var genProgs = []progExamp{
	{"1 * 2 + 3", false, 5.0},
	{"1 + 2 * 3", false, 7.0},
	{"3 * (4 + 5)", false, 27.0},
	{"2 ** 2 ** 2 ", false, 16.0},
	{"2 ** 2 ** 2 ** 2", false, 65536.0},
	{"3 / (4 + 6)", false, 0.3},
	{"-(2)", false, -2.0},
	{"--(2)", false, 2.0},
	{"3 * +5", false, 15.0},
	{"3 / -(4 + 6)", false, -0.3},
	{"(3 * 1 / 10 + 12) % 30", false, 12.3},
	{"1 > -5", false, 1.0},
	{"-5 > 1", false, 0.0},
	{"1 >= -5", false, 1.0},
	{"-5 >= 1", false, 0.0},
	{"1 < -5", false, 0.0},
	{"-5 < 1", false, 1.0},
	{"1 <= -5", false, 0.0},
	{"-5 <= 1", false, 1.0},
	{"20 % 5", false, 0.0},
	{"20 % 3", false, 2.0},
	{"True | (4 >= 5)", false, 1.0},
	{"False | (4 >= 5)", false, 0.0},
	{"False & (4 >= 5)", false, 0.0},
	{"True & (4 < 5)", false, 1.0},
	{"!(4 >= 5)", false, 1.0},
	{"!(4 < 5)", false, 0.0},
	{"True ^ (4 >= 5)", false, 1.0},
	{"False ^ (4 >= 5)", false, 0.0},
	//bad expr
	{"", true, -1},
	{"3 *", true, -1},
	{"* 3", true, -1},
	{"3 * 4 5", true, -1},
	{"3 * 4 + 5)", true, -1},
	{"3 * (4 + 5", true, -1},
	{"3 * (4 + 5", true, -1},
	{"3 * 4+) 5", true, -1},
	{"2 ** * 2 ** (2 ** 2", true, -1},
	{"*", true, -1},
	{"()", true, -1},
	{"(", true, -1},
	{"-", true, -1},
}

const Eps = 1e-9

func almostEqual(f, g float64) bool {
	return math.Abs(float64(f-g)) <= Eps
}

func TestGen(t *testing.T) {
	var expr *fxpratt.Expr
	for _, v := range genProgs {
		if testing.Verbose() {
			fmt.Fprintf(os.Stderr, "--> %s\n", v.input)
		}
		reader := bufio.NewReader(strings.NewReader(v.input))
		l, err := fxlex.NewLexer(reader, "test")
		if err != nil {
			t.Fatal(err)
		}
		p := fxpratt.NewParser(l)
		if err, expr = p.Parse(); err != nil && !v.isBad {
			errs := fmt.Sprintf("%s: %s", err, v.input)
			t.Fatal(errs)
		}

		val := expr.Eval()
		if v.isBad && err == nil {
			errs := fmt.Sprintf("%s should fail evals to %f", v.input, val)
			t.Fatal(errors.New(errs))
		} else if !v.isBad && !almostEqual(val, v.val) {
			errs := fmt.Sprintf("%s  is %f should be %f", v.input, val, v.val)
			t.Fatal(errs)
		}
	}
}
