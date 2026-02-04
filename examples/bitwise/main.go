package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"unicode/utf8"

	. "jnsn.in/katti"
)

type item struct {
	repr  string
	value int64
	op    string
}

func printExpression(items []item) {
	var maxReprWidth int
	var maxBinWidth int

	for _, t := range items {
		if w := len(t.repr); w > maxReprWidth {
			maxReprWidth = w
		}
		if w := len(strconv.FormatInt(t.value, 2)); w > maxBinWidth {
			maxBinWidth = w
		}
	}

	for _, t := range items {
		formatOp := " "

		if t.op != "" {
			formatOp = t.op
		}

		format := fmt.Sprintf("\t %v %%%ds %%0%db\n", formatOp, maxReprWidth, maxBinWidth)
		fmt.Printf(format, t.repr, t.value)
	}
}

func resolveNumberItem(raw string) (item, error) {
	it := item{}
	it.repr = raw

	r, size := utf8.DecodeRuneInString(raw)

	if r == '~' || r == '-' {
		raw = raw[size:]
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		return it, err
	}

	v := int64(n)

	switch r {
	case '-':
		it.value = -v
	case '~':
		it.value = ^v
	default:
		it.value = v
	}

	return it, nil
}

func numberListAction(result *MatchResult) error {
	var items []item

	for _, n := range result.Match {
		if item, err := resolveNumberItem(n); err != nil {
			return err
		} else {
			items = append(items, item)
		}
	}

	printExpression(items)
	return nil
}

func evalExpr(a, b item) item {
	av := a.value
	bv := b.value

	switch b.op {
	case "+":
		a.value = av + bv
	case "-":
		a.value = av - bv
	case "*":
		a.value = av * bv
	case "/":
		a.value = av / bv

	case "&":
		a.value = av & bv
	case "|":
		a.value = av | bv
	case "^":
		a.value = av ^ bv

	case "<<":
		a.value = av << bv
	case ">>":
		a.value = av >> bv
	case "<<<":
		s := uint(bv % 64)
		a.value = (av << s) | (av >> (64 - s))
	}

	a.repr = fmt.Sprintf("%d", a.value)

	return a
}

func expressionAction(result *MatchResult) (err error) {
	items := []item{}

	if len(result.Match) == 0 {
		return err
	}

	isNum := true

	for i, tok := range result.Match {
		if isNum {
			num, err := resolveNumberItem(tok)

			if err != nil {
				return err
			}

			if i == 0 {
				items = append(items, num)
			} else {
				last := &items[len(items)-1]
				last.value = num.value
				last.repr = num.repr
			}
		} else {
			items = append(items, item{op: tok})
		}

		isNum = !isNum
	}

	a := items[0]

	for _, b := range items[1:] {
		a = evalExpr(a, b)
	}

	a.op = "="
	items = append(items, a)

	printExpression(items)

	return nil
}

func Eval(expr string) (*MatchResult, error) {
	DIGIT := CharIn('0', '9')

	WS := Char(' ')

	BINARYOP := Alternation(
		Literal("<<<"),
		Literal("<<"),
		Literal(">>"),
		Char('&', '|', '^', '+', '-', '*', '/'),
	)

	UNARYOP := Char('-', '~')

	UNSIGNEDINT := Sequence(DIGIT, Repeat(DIGIT, true))

	UNARYEXPR := Sequence(UNARYOP, UNSIGNEDINT)

	SKIPWS := Skip(Optional(WS))

	OPERAND := Join(
		Alternation(
			UNSIGNEDINT,
			UNARYEXPR,
		),
	)

	NUMBERLIST := Action(
		SepBy(
			OPERAND,
			SKIPWS,
			false,
		),
		numberListAction,
	)

	EXPRESSION := Action(
		Sequence(
			OPERAND,
			Repeat(
				Sequence(
					SKIPWS,
					BINARYOP,
					SKIPWS,
					OPERAND,
				),
				false,
			),
		),
		expressionAction,
	)

	expressions := Sequence(
		Alternation(
			EXPRESSION,
			NUMBERLIST,
		),
		Optional(Char('\n')),
	)

	return Parse(expressions, expr)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')

		if _, err := Eval(text); err != nil {
			fmt.Println(err)
		}
	}
}
