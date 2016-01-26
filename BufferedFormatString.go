package presilo

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

/*
  Represents a buffered string which
*/
type BufferedFormatString struct {
	buffer          bytes.Buffer
	indentSeparator string
	indentation     int
}

/*
  Returns a new buffered format string.
  The given [separator] is used to properly indent lines as content is written.
  Best practice is to use the tab character ("\t"), but some languages
  may have legacy standards encouraging spaces or other whitespace characters,
  which should be used if the language expects them.
*/
func NewBufferedFormatString(separator string) *BufferedFormatString {

	var ret *BufferedFormatString

	ret = new(BufferedFormatString)
	ret.indentSeparator = separator
	return ret
}

/*
  Adds the given [count] of indentation points to the current running total of indentation.
  Note that [count] can be negative, in order to decrement indentation.

  Indentation will never go lower than zero.
*/
func (this *BufferedFormatString) AddIndentation(count int) {

	// I generally try to use "max" when possible, since standard library implementations in many, many places
	// have specific hardware / assembly paths for it, instead of using a jump.
	// Since Go doesn't have an integer max, i'm honestly not sure if casting two floats, doing max,
	// then casting back is cheaper. I know jumps are expensive, but not sure if they're worth three stack allocations.
	this.indentation = int(math.Max(float64(this.indentation)+float64(count), 0))
}

/*
  Appends the given [source] into this buffer.
*/
func (this *BufferedFormatString) Print(source string) {

	source = this.matchIndentation(source, this.indentation)
	this.buffer.WriteString(source)
}

/*
  Appends the given [source] into this buffer, interpolating the string in the same fashion as "fmt.Printf".
*/
func (this *BufferedFormatString) Printf(source string, parameters ...interface{}) {

	var toWrite string

	source = this.matchIndentation(source, this.indentation)
	toWrite = fmt.Sprintf(source, parameters...)

	this.buffer.WriteString(toWrite)
}

/*
  Same as "Printf", except appends a newline to the end of the given [source].
*/
func (this *BufferedFormatString) Printfln(source string, parameters ...interface{}) {

	this.Printf(source+"\n", parameters...)
}

/*
  Returns the current string representation of this buffer.
*/
func (this *BufferedFormatString) String() string {
	return this.buffer.String()
}

/*
	Indents all lines in the given [source] string with the given amount of [tabs].
*/
func (this *BufferedFormatString) matchIndentation(source string, tabs int) string {

	var replacement bytes.Buffer

	replacement.WriteRune('\n')
	for i := 0; i < tabs; i++ {
		replacement.WriteString(this.indentSeparator)
	}

	return strings.Replace(source, "\n", replacement.String(), -1)
}
