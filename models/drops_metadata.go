package models

import "fmt"

type DropsMetadata struct {
	Arena      string
	Challenger string
	Difficulty string
}

func (first DropsMetadata) Combine(second *DropsMetadataWithSource) (DropsMetadata, error) {
	if first.Arena != second.Arena {
		return DropsMetadata{}, fmt.Errorf("tried to combine two metadata that did not have the same arena: %s and %s", first, second)
	}

	copy := first.Copy()
	if first.Challenger != second.Challenger {
		copy.Challenger = "(multiple challengers)"
	}
	if first.Difficulty != second.Difficulty {
		copy.Difficulty = "(multiple difficulties)"
	}
	return copy, nil
}

func (first DropsMetadata) Copy() DropsMetadata {
	return DropsMetadata{
		Arena:      first.Arena,
		Challenger: first.Challenger,
		Difficulty: first.Difficulty,
	}
}

type DropsMetadataWithSource struct {
	Source string
	DropsMetadata
}

func (metadata *DropsMetadataWithSource) Copy() *DropsMetadataWithSource {
	copy := new(DropsMetadataWithSource)
	copy.Source = metadata.Source
	copy.Arena = metadata.Arena
	copy.Challenger = metadata.Challenger
	copy.Difficulty = metadata.Difficulty
	return copy
}

func (first *DropsMetadataWithSource) Combine(second *DropsMetadataWithSource) (*DropsMetadataWithSource, error) {
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

func (metadata *DropsMetadataWithSource) String() string {
	return fmt.Sprintf("%s - %s - %s - %s", metadata.Source, metadata.Arena, metadata.Challenger, metadata.Difficulty)
}

func (metadata *DropsMetadata) String() string {
	return fmt.Sprintf("%s - %s - %s", metadata.Arena, metadata.Challenger, metadata.Difficulty)
}

func GeneratedMetadata(arena string) *DropsMetadataWithSource {
	return &DropsMetadataWithSource{
		Source: "(generated)",
		DropsMetadata: DropsMetadata{
			Arena:      arena,
			Challenger: "(generated)",
			Difficulty: "(generated)",
		},
	}
}
