package presilo

import (
	"bytes"
	"unicode"
)

/*
  Converts the given target string to CamelCase; e.g.
  "something" becomes "Something"
*/
func ToCamelCase(target string) string {

	return iterateRunes(target, unicode.ToUpper)
}

/*
  Converts the given target string to javaCase, e.g.
  "SomethingElse" becomes "somethingElse"
*/
func ToJavaCase(target string) string {

	return iterateRunes(target, unicode.ToLower)
}

func iterateRunes(target string, transformer func(rune) rune) string {

	var ret bytes.Buffer
	var channel chan rune
	var character rune

	channel = make(chan rune)
	go getCharacterChannel(target, channel)

	character = <-channel
	character = transformer(character)
	ret.WriteString(string(character))

	for character = range channel {
		ret.WriteString(string(character))
	}

	return ret.String()
}

func getCharacterChannel(source string, channel chan rune) {

	defer close(channel)

	for _, rune := range source {
		channel <- rune
	}
}
