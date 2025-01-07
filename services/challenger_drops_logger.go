package services

type ChallengerDropsLogger struct{}

func NewChallengerDropsLogger() *ChallengerDropsLogger {
	return &ChallengerDropsLogger{}
}

// func (logger *ChallengerDropsLogger) LogChallengerDrops() error {
// 	slog.Info("Logging challenger drops")
// 	drops, err := NewDropDataService().GetAllDrops(constants.BATTLEDOME_DROPS_FOLDER)
// 	if err != nil {
// 		return fmt.Errorf("failed to log challenger drops: %w", err)
// 	}

// 	groupedDrops := helpers.GroupBy(drops, func(drop *models.BattledomeDrops) models.DropsMetadata {
// 		return drop.Metadata
// 	})

// 	orderedGroupedDrops := helpers.OrderByDescending(helpers.ToSlice(groupedDrops), func(tuple helpers.Tuple) float64 {
// 		challengerDrops := tuple.Elements[1].([]*models.BattledomeDrops)
// 		challengerItems := helpers.Reduce(challengerDrops, func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
// 			combined, _ := first.Union(second)
// 			return combined
// 		})

// 	})
// }
