package models

import "fmt"

type DropsMetadata struct {
	Source     string
	Arena      string
	Challenger string
	Difficulty string
}

func (metadata *DropsMetadata) Copy() *DropsMetadata {
	copy := new(DropsMetadata)
	copy.Source = metadata.Source
	copy.Arena = metadata.Arena
	copy.Challenger = metadata.Challenger
	copy.Difficulty = metadata.Difficulty
	return copy
}

func (first *DropsMetadata) Combine(second *DropsMetadata) (*DropsMetadata, error) {
	if first.Arena != second.Arena {
		return nil, fmt.Errorf("tried to combine two metadata that did not have the same arena: %s and %s", first, second)
	}

	copy := first.Copy()
	copy.Source = "(multiple sources)"
	if first.Challenger != second.Challenger {
		copy.Challenger = "(multiple challengers)"
	}
	if first.Difficulty != second.Difficulty {
		copy.Difficulty = "(multiple difficulties)"
	}
	return copy, nil
}

func (metadata *DropsMetadata) String() string {
	return fmt.Sprintf("%s - %s - %s - %s", metadata.Source, metadata.Arena, metadata.Challenger, metadata.Difficulty)
}
