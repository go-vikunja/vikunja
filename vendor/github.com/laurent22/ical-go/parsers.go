package ical

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

func ParseCalendar(data string) (*Node, error) {
	r := regexp.MustCompile("([\r|\t| ]*\n[\r|\t| ]*)+")
	lines := r.Split(strings.TrimSpace(data), -1)
	node, _, err, _ := parseCalendarNode(lines, 0)

	return node, err
}

func parseCalendarNode(lines []string, lineIndex int) (*Node, bool, error, int) {
	line := strings.TrimSpace(lines[lineIndex])
	_ = log.Println
	colonIndex := strings.Index(line, ":")
	if colonIndex <= 0 {
		return nil, false, errors.New("Invalid value/pair: " + line), lineIndex + 1
	}
	name := line[0:colonIndex]
	splitted := strings.Split(name, ";")
	var parameters map[string]string
	if len(splitted) >= 2 {
		name = splitted[0]
		parameters = make(map[string]string)
		for i := 1; i < len(splitted); i++ {
			p := strings.Split(splitted[i], "=")
			if len(p) != 2 {
				panic("Invalid parameter format: " + name)
			}
			parameters[p[0]] = p[1]
		}
	}
	value := line[colonIndex+1 : len(line)]

	if name == "BEGIN" {
		node := new(Node)
		node.Name = value
		node.Type = 1
		lineIndex = lineIndex + 1
		for {
			child, finished, _, newLineIndex := parseCalendarNode(lines, lineIndex)
			if finished {
				return node, false, nil, newLineIndex
			} else {
				if child != nil {
					node.Children = append(node.Children, child)
				}
				lineIndex = newLineIndex
			}
		}
	} else if name == "END" {
		return nil, true, nil, lineIndex + 1
	} else {
		node := new(Node)
		node.Name = name
		if name == "DESCRIPTION" || name == "SUMMARY" {
			text, newLineIndex := parseTextType(lines, lineIndex)
			node.Value = text
			node.Parameters = parameters
			return node, false, nil, newLineIndex
		} else {
			node.Value = value
			node.Parameters = parameters
			return node, false, nil, lineIndex + 1
		}
	}

	panic("Unreachable")
	return nil, false, nil, lineIndex + 1
}

func parseTextType(lines []string, lineIndex int) (string, int) {
	line := lines[lineIndex]
	colonIndex := strings.Index(line, ":")
	output := strings.TrimSpace(line[colonIndex+1 : len(line)])
	lineIndex++
	for {
		line := lines[lineIndex]
		if line == "" || line[0] != ' ' {
			return unescapeTextType(output), lineIndex
		}
		output += line[1:len(line)]
		lineIndex++
	}
	return unescapeTextType(output), lineIndex
}

func escapeTextType(input string) string {
	output := strings.Replace(input, "\\", "\\\\", -1)
	output = strings.Replace(output, ";", "\\;", -1)
	output = strings.Replace(output, ",", "\\,", -1)
	output = strings.Replace(output, "\n", "\\n", -1)
	return output
}

func unescapeTextType(s string) string {
	s = strings.Replace(s, "\\;", ";", -1)
	s = strings.Replace(s, "\\,", ",", -1)
	s = strings.Replace(s, "\\n", "\n", -1)
	s = strings.Replace(s, "\\\\", "\\", -1)
	return s
}
