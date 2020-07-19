package fxpratt

import (
	"errors"
	"fmt"
	"fxlex"
	"os"
	"strings"
)

/*
 * 	An interpretation of Pratt parsing
 *	https://web.archive.org/web/20151223215421/http://hall.org.ua/halls/wizzard/pdf/Vaughan.Pratt.TDOP.pdf
 *	Proceedings of the 1st Annual ACM SIGACT-SIGPLAN
 *	Symposium on Principles of Programming Languages (1973)
 */

type Parser struct {
	l     *fxlex.Lexer
	depth int
	tag   string
}

func NewParser(l *fxlex.Lexer) *Parser {
	return &Parser{l, 0, ""}
}

//https://web.archive.org/web/20151223215421/http://hall.org.ua/halls/wizzard/pdf/Vaughan.Pratt.TDOP.pdf

func (p *Parser) match(tT int) (tok fxlex.Token, err error, isMatch bool) {
	p.dPrintf("match: %s", tT)
	tok, err = p.l.Peek()
	if err != nil {
		return fxlex.Token{}, err, false
	}
	if tok.GetTokType() != tT {
		return tok, nil, false
	}
	p.l.Lex() //already peeked
	p.pushTrace(tok.String())
	defer p.popTrace(&err)
	return tok, nil, true
}

var precTab = map[rune]int{
	')':          1,
	'|':          10,
	'&':          10,
	'!':          10,
	'^':          10,
	'<':          10,
	'>':          10,
	fxlex.TokGTE: 10,
	fxlex.TokLTE: 10,
	'+':          20,
	'-':          20,
	'*':          30,
	'/':          30,
	'%':          30,
	fxlex.TokPow: 40,
	'(':          50,
}

var leftTab = map[rune]bool{
	fxlex.TokPow: true,
}
var unaryTab = map[rune]bool{
	'+': true,
	'-': true,
	'(': true,
	'!': true,
	'^': true,
}

//no left context, null-denotation: nud
func (p *Parser) Nud(tok fxlex.Token) (expr *Expr, err error) {
	var rExpr *Expr
	var rbp int
	p.dPrintf("Nud:  %d, %s \n", rbp, tok)
	if tok.GetTokType() == fxlex.TokLPar { //special unary, parenthesis
		expr, err = p.Expr(rbp)
		if err != nil {
			return nil, err
		}
		if _, err, isClosed := p.match(fxlex.TokRPar); err != nil {
			return nil, err
		} else if !isClosed {
			return nil, errors.New("unmatched parenthesis")
		}
		return expr, nil
	}
	expr = NewExpr(tok)
	rbp = bindPow(tok)
	rTok := rune(tok.GetTokType())
	if rbp != defRbp { //regular unary operators
		if !unaryTab[rTok] {
			errs := fmt.Sprintf("%s  is not unary", tok.GetType())
			return nil, errors.New(errs)
		}
		rExpr, err = p.Expr(rbp)
		if rExpr == nil {
			return nil, errors.New("unary operator without operand")
		}
		expr.ERight = rExpr
	}
	return expr, nil
}

//left context, left-denotation: led
func (p *Parser) Led(left *Expr, tok fxlex.Token) (expr *Expr, err error) {
	var rbp int
	expr = NewExpr(tok)
	expr.ELeft = left
	rbp = bindPow(tok)
	if isleft := leftTab[rune(tok.GetTokType())]; isleft {
		rbp -= 1
	}
	p.dPrintf("Led: %d, {{%s}} %s \n", rbp, left, tok)
	rExpr, err := p.Expr(rbp)
	if err != nil {
		return nil, err
	}
	if rExpr == nil {
		errs := fmt.Sprintf("missing operand for %s\n", tok.GetType())
		return nil, errors.New(errs)
	}
	expr.ERight = rExpr
	return expr, nil
}

const defRbp = 0

func bindPow(tok fxlex.Token) int {
	if rbp, ok := precTab[rune(tok.GetTokType())]; ok {
		return rbp
	}
	return defRbp
}

func (p *Parser) Expr(rbp int) (expr *Expr, err error) {
	var left *Expr

	s := fmt.Sprintf("Expr: %d", rbp)
	p.pushTrace(s)
	defer p.popTrace(&err)

	tok, err := p.l.Peek()
	if err != nil {
		return expr, err
	}
	p.dPrintf("expr: Nud Lex: %s", tok)

	if tok.GetTokType() == fxlex.RuneEOF {
		return expr, nil
	}
	p.l.Lex() //already peeked
	if left, err = p.Nud(tok); err != nil {
		return nil, err
	}
	expr = left
	for {
		tok, err := p.l.Peek()
		if err != nil {
			return expr, err
		}
		if tok.GetTokType() == fxlex.RuneEOF || tok.GetTokType() == fxlex.TokRPar {
			return expr, nil
		}
		if bindPow(tok) <= rbp {
			p.dPrintf("Not enough binding: %d <= %d, %s\n", bindPow(tok), rbp, tok)
			return left, nil
		}
		p.l.Lex() //already peeked
		p.dPrintf("expr: led Lex: %s", tok)
		if left, err = p.Led(left, tok); err != nil {
			return expr, err
		}
		expr = left
	}
}

//PROG :== EXPR EOF
func (p *Parser) Prog() (e error, expr *Expr) {
	p.pushTrace("Prog")
	defer p.popTrace(&e)
	expr, err := p.Expr(defRbp - 1)
	if err != nil {
		return err, nil
	}
	if expr == nil {
		return errors.New("empty expression"), nil
	}
	t, err, isEof := p.match(fxlex.RuneEOF)
	if err != nil {
		return err, nil
	}
	if !isEof {
		es := fmt.Sprintf("need %s, got %s", "TokEOF", t.GetType())
		return errors.New(es), nil
	}
	return nil, expr
}

func (p *Parser) Parse() (err error, expr *Expr) {
	p.pushTrace("Parse")
	defer p.popTrace(&err)
	if err, expr = p.Prog(); err != nil {
		return err, nil
	}

	return nil, expr
}

const DebugDesc = true

func (p *Parser) pushTrace(tag string) {
	if DebugDesc {
		tabs := strings.Repeat("\t", p.depth)
		fmt.Fprintf(os.Stderr, "->%s%s\n", tabs, tag)
	}
	p.tag = tag
	p.depth++
}

func (p *Parser) dPrintf(format string, a ...interface{}) {
	if DebugDesc {
		tabs := strings.Repeat("\t", p.depth)
		format = fmt.Sprintf("%s%s", tabs, format)
		fmt.Fprintf(os.Stderr, format, a...)
	}
}

func (p *Parser) popTrace(e *error) {
	if e != nil && *e != nil {
		if DebugDesc {
			tabs := strings.Repeat("\t", p.depth)
			fmt.Fprintf(os.Stderr, "<-%s%s:%s\n", tabs, p.tag, *e)
		}
	}
	p.tag = ""
	p.depth--
}
