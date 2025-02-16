package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

// Code ExcludedSearchPhrases.
const ExcludedSearchPhrasesCode Code = 160

type ExcludedSearchPhrases struct {
	Phrases []string
}

func (e *ExcludedSearchPhrases) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 160
	if err != nil {
		return err
	}

	if code != uint32(ExcludedSearchPhrasesCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ExcludedSearchPhrasesCode, code))
	}

	numberOfPhrases, err := soul.ReadUint32(reader)
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
