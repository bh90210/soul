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
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 160
	if err != nil {
		return err
	}

	if code != ExcludedSearchPhrasesCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ExcludedSearchPhrasesCode, code))
	}

	numberOfPhrases, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(numberOfPhrases); i++ {
		phrase, err := soul.ReadString(reader)
		if err != nil {
			return err
		}

		e.Phrases = append(e.Phrases, phrase)
	}

	return nil
}
