package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

// Code ExcludedSearchPhrases.
const ExcludedSearchPhrasesCode soul.UInt = 160

type ExcludedSearchPhrases struct {
	Phrases []string
}

func (e *ExcludedSearchPhrases) Deserialize(reader io.Reader) error {
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 160
	if code != ExcludedSearchPhrasesCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ExcludedSearchPhrasesCode, code))

	}

	numberOfPhrases := soul.ReadUInt(reader)
	for i := 0; i < int(numberOfPhrases); i++ {
		phrase := soul.ReadString(reader)

		e.Phrases = append(e.Phrases, phrase)
	}

	return nil
}
