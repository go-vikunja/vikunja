package ical

// import (
// 	"time"
// 	"strconv"
// 	"strings"
// )
//
// func TodoFromNode(node *Node) Todo {
// 	if node.Name != "VTODO" { panic("Node is not a VTODO") }
//
// 	var todo Todo
// 	todo.SetId(node.PropString("UID", ""))
// 	todo.SetSummary(node.PropString("SUMMARY", ""))
// 	todo.SetDescription(node.PropString("DESCRIPTION", ""))
// 	todo.SetDueDate(node.PropDate("DUE", time.Time{}))
// 	//todo.SetAlarmDate(this.TimestampBytesToTime(reminderDate))
// 	todo.SetCreatedDate(node.PropDate("CREATED", time.Time{}))
// 	todo.SetModifiedDate(node.PropDate("DTSTAMP", time.Time{}))
// 	todo.SetPriority(node.PropInt("PRIORITY", 0))
// 	todo.SetPercentComplete(node.PropInt("PERCENT-COMPLETE", 0))
// 	return todo
// }
//
// type Todo struct {
// 	CalendarItem
// 	dueDate time.Time
// }
//
// func (this *Todo) SetDueDate(v time.Time) { this.dueDate = v }
// func (this *Todo) DueDate() time.Time { return this.dueDate }
//
// func (this *Todo) ICalString(target string) string {
// 	s := "BEGIN:VTODO\n"
//
// 	if target == "macTodo" {
// 		status := "NEEDS-ACTION"
// 		if this.PercentComplete() == 100 {
// 			status = "COMPLETED"
// 		}
// 		s += "STATUS:" + status + "\n"
// 	}
//
// 	s += encodeDateProperty("CREATED", this.CreatedDate()) + "\n"
// 	s += "UID:" + this.Id() + "\n"
// 	s += "SUMMARY:" + escapeTextType(this.Summary()) + "\n"
// 	if this.PercentComplete() == 100 && !this.CompletedDate().IsZero() {
// 		s += encodeDateProperty("COMPLETED", this.CompletedDate()) + "\n"
// 	}
// 	s += encodeDateProperty("DTSTAMP", this.ModifiedDate()) + "\n"
// 	if this.Priority() != 0 {
// 		s += "PRIORITY:" + strconv.Itoa(this.Priority()) + "\n"
// 	}
// 	if this.PercentComplete() != 0 {
// 		s += "PERCENT-COMPLETE:" + strconv.Itoa(this.PercentComplete()) + "\n"
// 	}
// 	if target == "macTodo" {
// 		s += "SEQUENCE:" + strconv.Itoa(this.Sequence()) + "\n"
// 	}
// 	if this.Description() != "" {
// 		s += "DESCRIPTION:" + encodeTextType(this.Description()) + "\n"
// 	}
//
// 	s += "END:VTODO\n"
//
// 	return s
// }
//
// func encodeDateProperty(name string, t time.Time) string {
// 	var output string
// 	zone, _ := t.Zone()
// 	if zone != "UTC" && zone != "" {
// 		output = ";TZID=" + zone + ":" + t.Format("20060102T150405")
// 	} else {
// 		output = ":" + t.Format("20060102T150405") + "Z"
// 	}
// 	return name + output
// }
//
//
// func encodeTextType(s string) string {
// 	output := ""
// 	s = escapeTextType(s)
// 	lineLength := 0
// 	for _, c := range s {
// 		if lineLength + len(string(c)) > 75 {
// 			output += "\n "
// 			lineLength = 1
// 		}
// 		output += string(c)
// 		lineLength += len(string(c))
// 	}
// 	return output
// }
