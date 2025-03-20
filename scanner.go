package mtparser

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"text/scanner"
)

func New(r *bufio.Reader) (Parser, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Parser{}, err
	}
	// We remove carriage returns from the data before parsing
	// because Windows uses \r\n as line endings and we only want \n.
	clean := strings.ReplaceAll(string(data), "\r", "")
	newData := bufio.NewReader(strings.NewReader(clean))

	var s Parser
	s.Init(newData)
	s.Mode = scanner.ScanIdents
	s.Whitespace = 1<<'\t' | 1<<'\r'
	s.IsIdentRune = func(ch rune, i int) bool {
		switch ch {
		case '{', '}', ':', '\n', scanner.EOF:
			return false
		}
		return true
	}
	s.ErrPrefix = "We could not parse the payment message provided."
	s.Map = map[string]map[string]Node{}
	return s, nil
}

func (s *Parser) ErrMessage(c rune, x bool) string {
	ln := strconv.Itoa(s.Pos().Line)
	cl := strconv.Itoa(s.Pos().Column)
	xp := "Expected"
	if !x {
		xp = "Unexpected"
	}
	return s.ErrPrefix + " " + xp + " '" + string(c) + "' at line " + ln + " column " + cl
}

func (s *Parser) Parse() error {
	var err error

	for s.Peek() != scanner.EOF {
		if s.Scan() != '{' {
			return errors.New(s.ErrMessage('{', true))
		}

		s.Scan()
		s.blk.Key = s.TokenText()

		if s.Scan() != ':' {
			return errors.New(s.ErrMessage(':', true))
		}

		switch s.Peek() {
		case '\n':
			if err = s.scanBody(); err != nil {
				return err
			}
			break
		case '{':
			if err = s.scanBlocks(); err != nil {
				return err
			}
			break
		default:
			if err = s.scanHeader(); err != nil {
				return err
			}
			break
		}

		if s.Scan() != '}' {
			return errors.New(s.ErrMessage('}', true))
		}

		s.Blocks = append(s.Blocks, *&s.blk)
	}

	return nil
}
