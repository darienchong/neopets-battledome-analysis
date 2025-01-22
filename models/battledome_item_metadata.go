package models

import "fmt"

type Arena string

type Challenger string

type Difficulty string

type BattledomeItemMetadata struct {
	Arena      Arena
	Challenger Challenger
	Difficulty Difficulty
}

func (first BattledomeItemMetadata) Combine(second BattledomeItemMetadata) (BattledomeItemMetadata, error) {
	if first.Arena != second.Arena {
		return BattledomeItemMetadata{}, fmt.Errorf("tried to combine two metadata that did not have the same arena: %s and %s", first, second)
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

func (first BattledomeItemMetadata) Copy() BattledomeItemMetadata {
	return BattledomeItemMetadata{
		Arena:      first.Arena,
		Challenger: first.Challenger,
		Difficulty: first.Difficulty,
	}
}

type DropsMetadataWithSource struct {
	Source string
	BattledomeItemMetadata
}

func (md *DropsMetadataWithSource) Copy() *DropsMetadataWithSource {
	copy := new(DropsMetadataWithSource)
	copy.Source = md.Source
	copy.BattledomeItemMetadata = md.BattledomeItemMetadata
	return copy
}

func (first *DropsMetadataWithSource) Combine(second *DropsMetadataWithSource) (*DropsMetadataWithSource, error) {
	if first.Arena != second.Arena {
		return nil, fmt.Errorf("tried to combine two metadata that did not have the same arena: %s and %s", first, second)
	}

	copy := first.Copy()
	copy.Source = "(multiple sources)"
	return copy, nil
}

func (md *DropsMetadataWithSource) String() string {
	return fmt.Sprintf("%s - %s - %s - %s", md.Source, md.Arena, md.Challenger, md.Difficulty)
}

func (md *BattledomeItemMetadata) String() string {
	return fmt.Sprintf("%s - %s - %s", md.Arena, md.Challenger, md.Difficulty)
}

func GeneratedMetadata(arena Arena) *DropsMetadataWithSource {
	return &DropsMetadataWithSource{
		Source: "(generated)",
		BattledomeItemMetadata: BattledomeItemMetadata{
			Arena:      arena,
			Challenger: "(generated)",
			Difficulty: "(generated)",
		},
	}
}
