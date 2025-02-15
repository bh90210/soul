package excludedsearchphrases

import (
	"io"

	"github.com/bh90210/soul"
)

// Code ExcludedSearchPhrases.
const Code soul.UInt = 160

type Response struct {
	Phrases []string
}

func Deserialize(reader io.Reader) *Response {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 160

	numberOfPhrases := soul.ReadUInt(reader)

	phrases := make([]string, 0)
	for i := 0; i < int(numberOfPhrases); i++ {
		phrase := soul.ReadString(reader)

		phrases = append(phrases, phrase)
	}

	return &Response{phrases}

}
