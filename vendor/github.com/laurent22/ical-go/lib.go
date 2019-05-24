package ical

import (
	"bytes"
	"fmt"
)

type strBuffer struct {
	buffer bytes.Buffer
}

func (b *strBuffer) Write(format string, elem ...interface{}) {
	b.buffer.WriteString(fmt.Sprintf(format, elem...))
}

func (b *strBuffer) String() string {
	return b.buffer.String()
}
