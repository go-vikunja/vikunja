# ical-go

iCal package for Go (Golang)

## Installation

    go get github.com/laurent22/ical-go

## Status

Currently, the package doesn't support the full iCal specification. It's still a work in progress towards that goal.

The most useful function in the package is:

```go
func ParseCalendar(data string) (*Node, error)
```
Parses a VCALENDAR string, unwrap and unfold lines, etc. and put all this into a usable structure (a collection of `Node`s with name, value, type, etc.).

With the `Node` in hand, you can use several of its functions to, e.g., find specific parameters, children, etc.


## License

MIT
