package lexer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Lexer struct {
	chunk         string
	chunkName     string
	line          int
	nextTokenLine int
	nextTokenKind int
	nextToken     string
}

var (
	reNewLine            = regexp.MustCompile("\r\n|\n\r|\n|\r")
	reIdentifier         = regexp.MustCompile(`^[_\d\w]+`)
	reNumber             = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
	reShortStr           = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
	reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)
	reDecEscapeSeq       = regexp.MustCompile(`^\\[0-9]{1,3}`)
	reHexEscapeSeq       = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
	reUnicodeEscapeSeq   = regexp.MustCompile(`^\\u{[0-9a-fA-F]+}`)
)

func NewLexer(chunk, chunkName string) *Lexer {
	return &Lexer{
		chunk:     chunk,
		chunkName: chunkName,
		line:      1,
	}
}

func (self *Lexer) skip(n int) {
	self.chunk = self.chunk[n:]
}

func (self *Lexer) test(s string) bool {
	return strings.HasPrefix(self.chunk, s)
}

func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

func (self *Lexer) skipWhiteSpaces() {
	for len(self.chunk) > 0 {
		if self.test("--") {
			self.skipComment()
		} else if self.test("\r\n") || self.test("\n\r") {
			self.skip(2)
			self.line++
		} else if isNewLine(self.chunk[0]) {
			self.skip(1)
			self.line++
		} else if isWhiteSpace(self.chunk[0]) {
			self.skip(1)
		} else {
			break
		}
	}
}

func (self *Lexer) skipComment() {
	self.skip(2)
	if self.test("[") {
		if reOpeningLongBracket.FindString(self.chunk) != "" {
			self.scanLongString()
			return
		}
	}
	for len(self.chunk) > 0 && !isNewLine(self.chunk[0]) {
		self.skip(1)
	}
}

func (self *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf("%s:%d: %s", self.chunkName, self.line, fmt.Sprintf(f, a...))
	panic(err)
}

func (self *Lexer) scanLongString() string {
	openingLongBracket := reOpeningLongBracket.FindString(self.chunk)
	if openingLongBracket == "" {
		self.error("invalid long string near '%s'", self.chunk[0:2])
	}
	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIndex := strings.Index(self.chunk, closingLongBracket)
	if closingLongBracketIndex < 0 {
		self.error("unfinished long string or comment")
	}
	str := self.chunk[len(openingLongBracket):closingLongBracketIndex]
	self.skip(closingLongBracketIndex + len(closingLongBracket))
	reNewLine.ReplaceAllString(str, "\n")
	self.line += strings.Count(str, "\n")
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}
	return str
}

func (self *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, _kind, token := self.NextToken()
	if kind != _kind {
		self.error("syntax error near '%s'", token)
	}
	return
}

func (self *Lexer) NextIdentifier() (line int, token string) {
	return self.NextTokenOfKind(TOKEN_IDENTIFIER)
}

func (self *Lexer) Line() int {
	return self.line
}

func (self *Lexer) NextToken() (line, kind int, token string) {
	if self.nextTokenLine > 0 {
		line = self.nextTokenLine
		kind = self.nextTokenKind
		token = self.nextToken
		self.line = self.nextTokenLine
		self.nextTokenLine = 0
		return
	}

	self.skipWhiteSpaces()
	if len(self.chunk) == 0 {
		return self.line, TOKEN_EOF, "EOF"
	}

	switch self.chunk[0] {
	case ';':
		self.skip(1)
		return self.line, TOKEN_SEP_SEMI, ";"
	case ',':
		self.skip(1)
		return self.line, TOKEN_SEP_COMMA, ","
	case '(':
		self.skip(1)
		return self.line, TOKEN_SEP_LPAREN, "("
	case ')':
		self.skip(1)
		return self.line, TOKEN_SEP_RPAREN, ")"
	case ']':
		self.skip(1)
		return self.line, TOKEN_SEP_RBRACK, "]"
	case '{':
		self.skip(1)
		return self.line, TOKEN_SEP_LCURLY, "{"
	case '}':
		self.skip(1)
		return self.line, TOKEN_SEP_RCURLY, "}"
	case '+':
		self.skip(1)
		return self.line, TOKEN_OP_ADD, "+"
	case '-':
		self.skip(1)
		return self.line, TOKEN_OP_MINUS, "-"
	case '*':
		self.skip(1)
		return self.line, TOKEN_OP_MUL, "*"
	case '^':
		self.skip(1)
		return self.line, TOKEN_OP_POW, "^"
	case '%':
		self.skip(1)
		return self.line, TOKEN_OP_MOD, "%"
	case '&':
		self.skip(1)
		return self.line, TOKEN_OP_BAND, "&"
	case '|':
		self.skip(1)
		return self.line, TOKEN_OP_BOR, "|"
	case '#':
		self.skip(1)
		return self.line, TOKEN_OP_LEN, "#"
	case ':':
		if self.test("::") {
			self.skip(2)
			return self.line, TOKEN_SEP_LABEL, "::"
		} else {
			self.skip(1)
			return self.line, TOKEN_SEP_COLON, ":"
		}
	case '/':
		if self.test("//") {
			self.skip(2)
			return self.line, TOKEN_OP_IDIV, "//"
		} else {
			self.skip(1)
			return self.line, TOKEN_OP_DIV, "/"
		}
	case '~':
		if self.test("~=") {
			self.skip(2)
			return self.line, TOKEN_OP_NE, "~="
		} else {
			self.skip(1)
			return self.line, TOKEN_OP_WAVE, "~"
		}
	case '=':
		if self.test("==") {
			self.skip(2)
			return self.line, TOKEN_OP_EQ, "=="
		} else {
			self.skip(1)
			return self.line, TOKEN_OP_ASSIGN, "="
		}
	case '<':
		if self.test("<<") {
			self.skip(2)
			return self.line, TOKEN_OP_SHL, "<<"
		} else if self.test("<=") {
			self.skip(2)
			return self.line, TOKEN_OP_LE, "<="
		} else {
			self.skip(1)
			return self.line, TOKEN_OP_LT, "<"
		}
	case '>':
		if self.test(">>") {
			self.skip(2)
			return self.line, TOKEN_OP_SHR, ">>"
		} else if self.test(">=") {
			self.skip(2)
			return self.line, TOKEN_OP_GE, ">="
		} else {
			self.skip(1)
			return self.line, TOKEN_OP_GT, ">"
		}
	case '.':
		if self.test("...") {
			self.skip(3)
			return self.line, TOKEN_VARARG, "..."
		} else if self.test("..") {
			self.skip(2)
			return self.line, TOKEN_OP_CONCAT, ".."
		} else if len(self.chunk) == 1 || !isDigit(self.chunk[1]) {
			self.skip(1)
			return self.line, TOKEN_SEP_DOT, "."
		}
	case '[':
		if self.test("[[") || self.test("[=") {
			return self.line, TOKEN_STRING, self.scanLongString()
		} else {
			self.skip(1)
			return self.line, TOKEN_SEP_LBRACK, "["
		}
	case '\'', '"':
		return self.line, TOKEN_STRING, self.scanShortString()
	}

	c := self.chunk[0]
	if c == '.' || isDigit(c) {
		token := self.scan(reNumber)
		return self.line, TOKEN_NUMBER, token
	}
	if c == '_' || isLetter(c) {
		token := self.scanIdentifier()
		if kind, found := keywords[token]; found {
			return self.line, kind, token // keyword
		} else {
			return self.line, TOKEN_IDENTIFIER, token
		}
	}
	self.error("unexpected symbol near %q", c)
	return
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (self *Lexer) scanShortString() string {
	if str := reShortStr.FindString(self.chunk); str != "" {
		self.skip(len(str))
		str = str[1 : len(str)-1]
		if strings.Index(str, `\`) >= 0 {
			self.line += len(reNewLine.FindAllString(str, -1))
			str = self.escape(str)
		}
		return str
	}
	self.error("unfinished short string")
	return ""
}

func (self *Lexer) scanIdentifier() string {
	return self.scan(reIdentifier)
}

func (self *Lexer) escape(str string) string {
	var buf bytes.Buffer
	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}
		if len(str) == 1 {
			self.error("unfinished string")
		}
		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				self.error("decimal escape too large near '%s'", found)
			}
		case 'x':
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u':
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				self.error("too large UTF-8 value near '%s'", found)
			}
		case 'z':
			str = str[2:]
			if len(str) > 0 && isWhiteSpace(str[0]) {
				str = str[1:]
			}
			continue
		}
		self.error("invalid escape sequence near '\\%c", str[1])
	}
	return buf.String()
}

func (self *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(self.chunk); token != "" {
		self.skip(len(token))
		return token
	}
	panic("unreachable")
}

func (self *Lexer) LookAhead() int {
	if self.nextTokenLine > 0 {
		return self.nextTokenKind
	}
	currentline := self.line
	line, kind, token := self.NextToken()
	self.line = currentline
	self.nextTokenLine = line
	self.nextTokenKind = kind
	self.nextToken = token
	return kind
}
