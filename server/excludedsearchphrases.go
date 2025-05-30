package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// Code ExcludedSearchPhrases.
const CodeExcludedSearchPhrases Code = 160

type ExcludedSearchPhrases struct {
	Phrases []string
}

func (e *ExcludedSearchPhrases) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 160
	if err != nil {
		return err
	}

	if code != uint32(CodeExcludedSearchPhrases) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeExcludedSearchPhrases, code))
	}

	numberOfPhrases, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for range int(numberOfPhrases) {
		phrase, err := internal.ReadString(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		e.Phrases = append(e.Phrases, phrase)
	}

	return err
}
