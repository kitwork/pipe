package pipe

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type TokenType string

const (
	Num   TokenType = "num"
	Var   TokenType = "var"
	Op    TokenType = "op"
	Paren TokenType = "paren"
	Tern  TokenType = "tern"
)

type Token struct {
	typ   TokenType
	value string
}

var opMap = map[string]string{
	"+": "add", "-": "sub", "*": "mul", "/": "div",
	"==": "eq", "!=": "ne", ">": "gt", "<": "lt", ">=": "gte", "<=": "lte",
	"&&": "and", "||": "or", "!": "not", "??": "nullish",
	"?": "tern",
}

var precedence = map[string]int{
	"!": 6, "*": 5, "/": 5, "+": 4, "-": 4,
	">": 3, "<": 3, ">=": 3, "<=": 3, "==": 3, "!=": 3,
	"&&": 2, "||": 1, "??": 0, // ternary handled separately
}

// tokenize: supports numbers, identifiers (with dots), strings (not quoted here), parens and operators
func tokenize(expr string) []Token {
	var tokens []Token
	i := 0
	n := len(expr)

	for i < n {
		ch := expr[i]

		// spaces
		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}

		// paren
		if ch == '(' || ch == ')' {
			tokens = append(tokens, Token{Paren, string(ch)})
			i++
			continue
		}

		// two-char ops
		if i+1 < n {
			two := expr[i : i+2]
			if two == "==" || two == "!=" || two == ">=" || two == "<=" ||
				two == "&&" || two == "||" || two == "??" {
				tokens = append(tokens, Token{Op, two})
				i += 2
				continue
			}
		}

		// single char ops (include ? and :)
		if strings.ContainsRune("+-*/><!:?", rune(ch)) {
			tokens = append(tokens, Token{Op, string(ch)})
			i++
			continue
		}

		// number (integer or float)
		if unicode.IsDigit(rune(ch)) {
			start := i
			hasDot := false
			for i < n && (unicode.IsDigit(rune(expr[i])) || (!hasDot && expr[i] == '.')) {
				if expr[i] == '.' {
					hasDot = true
				}
				i++
			}
			tokens = append(tokens, Token{Num, expr[start:i]})
			continue
		}

		// identifier/var (allow dot for deep access, and underscores)
		if unicode.IsLetter(rune(ch)) || ch == '_' || ch == '$' {
			start := i
			for i < n && (unicode.IsLetter(rune(expr[i])) || unicode.IsDigit(rune(expr[i])) || expr[i] == '_' || expr[i] == '.') {
				i++
			}
			tokens = append(tokens, Token{Var, expr[start:i]})
			continue
		}

		// unknown: treat as single char token
		tokens = append(tokens, Token{Var, string(ch)})
		i++
	}

	return tokens
}

// parseTernary: find "cond ? a : b" occurrences and collapse to single Tern token.
// We build Tern token value as: tern(<cond_expr>,<true_expr>,<false_expr>) without converting those sub-expr to pipeline yet.
func parseTernary(tokens []Token) []Token {
	var out []Token
	i := 0
	for i < len(tokens) {
		// find '?'
		if tokens[i].typ == Op && tokens[i].value == "?" {
			// cond is last item in out
			if len(out) == 0 {
				// malformed, just append and continue
				out = append(out, tokens[i])
				i++
				continue
			}
			cond := out[len(out)-1]
			out = out[:len(out)-1]

			// find matching ':'
			depth := 0
			j := i + 1
			for ; j < len(tokens); j++ {
				if tokens[j].typ == Op && tokens[j].value == "?" {
					depth++
				} else if tokens[j].typ == Op && tokens[j].value == ":" {
					if depth == 0 {
						break
					}
					depth--
				}
			}
			if j >= len(tokens) {
				// malformed: no matching colon -> append remaining and break
				out = append(out, cond)
				out = append(out, tokens[i:]...)
				break
			}
			trueBranch := tokens[i+1 : j]
			// falseBranch is remainder after j
			falseBranch := tokens[j+1:]

			// create Tern token value by serializing sub-tokens to string (we'll re-tokenize these when needed)
			tb := tokensToString(trueBranch)
			fb := tokensToString(falseBranch)
			out = append(out, Token{Tern, fmt.Sprintf("tern(%s,%s,%s)", cond.value, tb, fb)})
			// Done with whole expression; break the loop because falseBranch consumed rest
			break
		}

		out = append(out, tokens[i])
		i++
	}
	return out
}

func tokensToString(ts []Token) string {
	var b strings.Builder
	for i, t := range ts {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(t.value)
	}
	return b.String()
}

// toRPN (Shunting-yard). Treat Tern as an operand token (it is already a single token).
func toRPN(tokens []Token) []Token {
	var out []Token
	var stack []Token

	for _, tok := range tokens {
		switch tok.typ {
		case Var, Num, Tern:
			out = append(out, tok)
		case Op:
			// handle '(' / ')'
			if tok.value == "(" {
				stack = append(stack, tok)
				continue
			}
			if tok.value == ")" {
				for len(stack) > 0 && stack[len(stack)-1].value != "(" {
					out = append(out, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				if len(stack) > 0 && stack[len(stack)-1].value == "(" {
					stack = stack[:len(stack)-1]
				}
				continue
			}
			for len(stack) > 0 && stack[len(stack)-1].typ == Op {
				top := stack[len(stack)-1]
				// if top has higher or equal precedence, pop it
				if precedence[top.value] >= precedence[tok.value] {
					out = append(out, top)
					stack = stack[:len(stack)-1]
					continue
				}
				break
			}
			stack = append(stack, tok)
		case Paren:
			if tok.value == "(" {
				stack = append(stack, tok)
			} else {
				for len(stack) > 0 && stack[len(stack)-1].value != "(" {
					out = append(out, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				if len(stack) > 0 && stack[len(stack)-1].value == "(" {
					stack = stack[:len(stack)-1]
				}
			}
		}
	}

	for len(stack) > 0 {
		out = append(out, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	return out
}

// rpnToPipeline: safe, checks stack underflow, supports Tern token.
func rpnToPipeline(rpn []Token) string {
	// fallback: nếu có gì sai -> trả lại RPN gốc
	original := tokensToString(rpn)

	var stack []string

	for _, tok := range rpn {
		switch tok.typ {

		case Var, Num:
			stack = append(stack, tok.value)

		case Op:
			fn, ok := opMap[tok.value]
			if !ok {
				// unknown operator → fallback
				return original
			}

			// Unary operator
			if tok.value == "!" {
				if len(stack) < 1 {
					return original
				}
				a := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				stack = append(stack, fmt.Sprintf("%s | %s", a, fn))
				continue
			}

			// Binary operator
			if len(stack) < 2 {
				return original
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			stack = append(stack, fmt.Sprintf("%s | %s %s", a, fn, b))

		default:
			return original
		}
	}

	// Invalid RPN
	if len(stack) != 1 {
		return original
	}

	return stack[0]
}

// splitTopLevel splits s by sep but only at top level (not inside parentheses)
func splitTopLevel(s string, sep rune) []string {
	var parts []string
	level := 0
	cur := strings.Builder{}
	for _, r := range s {
		if r == '(' {
			level++
		} else if r == ')' {
			if level > 0 {
				level--
			}
		}
		if r == sep && level == 0 {
			parts = append(parts, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteRune(r)
	}
	parts = append(parts, cur.String())
	return parts
}

// replaceVars: replace $a.b with .var.a.b except when followed by := or = or : (assignment)
// use simple regex, but compute replacements by indices to avoid infinite loops
func replaceVars(expr string) string {
	re := regexp.MustCompile(`\$(\w+(?:\.\w+)*)`)
	matches := re.FindAllStringIndex(expr, -1)
	if len(matches) == 0 {
		return expr
	}
	var out strings.Builder
	last := 0
	for _, pos := range matches {
		start, end := pos[0], pos[1]
		out.WriteString(expr[last:start])
		ref := expr[start:end]
		key := strings.TrimPrefix(ref, "$")
		after := ""
		if end < len(expr) {
			after = expr[end:]
		}
		// skip assignment forms
		if strings.HasPrefix(after, ":=") || strings.HasPrefix(after, "=") || strings.HasPrefix(after, ":") {
			out.WriteString(ref)
		} else {
			// replace with dot-access .var.xxx
			out.WriteString("Vars." + key)
		}
		last = end
	}
	out.WriteString(expr[last:])
	return out.String()
}

// Preprocessor: replace $vars, then convert expressions to pipeline if they contain ops
func Preprocessor(tmpl string) string {
	start := 0
	var result strings.Builder

	for start < len(tmpl) {
		idx := strings.Index(tmpl[start:], "{{")
		if idx == -1 {
			result.WriteString(tmpl[start:])
			break
		}
		result.WriteString(tmpl[start : start+idx])

		endIdx := strings.Index(tmpl[start+idx:], "}}")
		if endIdx == -1 {
			result.WriteString(tmpl[start+idx:])
			break
		}

		expr := strings.TrimSpace(tmpl[start+idx+2 : start+idx+endIdx])

		// 1) replace $vars -> .var.xxx
		expr = replaceVars(expr)

		// 2) if expression contains operators, parse and convert to pipeline
		if strings.ContainsAny(expr, "+-*/><=!&|?:") {

			tokens := tokenize(expr)
			tokens = parseTernary(tokens)
			rpn := toRPN(tokens)
			newExpr := rpnToPipeline(rpn)
			newExpr = strings.ReplaceAll(newExpr, "Vars.", "$.Vars.")
			result.WriteString("{{ " + newExpr + " }}")
		} else {
			expr = strings.ReplaceAll(expr, "Vars.", "$.Vars.")
			result.WriteString("{{ " + expr + " }}")
		}

		start = start + idx + endIdx + 2
	}
	return result.String()
}
