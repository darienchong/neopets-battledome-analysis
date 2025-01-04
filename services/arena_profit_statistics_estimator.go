package services

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/montanaflynn/stats"
)

type ArenaProfitStatisticsParser struct{}

func (parser *ArenaProfitStatisticsParser) Save(arenaStats []*models.ArenaProfitStatistics, expiry time.Time, filePath string) error {
	slog.Info(fmt.Sprintf("Saving arena statistics to %s", filePath))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(expiry.Format(constants.DATA_EXPIRY_TIME_LAYOUT))
	if err != nil {
		return err
	}

	for _, arenaStat := range arenaStats {
		_, err := file.WriteString(arenaStat.String())
		if err != nil {
			return err
		}
	}
	return nil
}

func (parser *ArenaProfitStatisticsParser) Parse(filePath string) ([]*models.ArenaProfitStatistics, time.Time, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, time.Now(), fmt.Errorf("arena profit statistics file does not exist! The file path provided was \"%s\"", filePath)
	}

	stats := []*models.ArenaProfitStatistics{}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, time.Now(), err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	expiry := time.Now().AddDate(0, 0, 7)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.Contains(line, "|") {
			parsedExpiry, err := time.Parse(constants.DATA_EXPIRY_TIME_LAYOUT, line)
			if err != nil {
				return nil, time.Now(), err
			}
			expiry = parsedExpiry
			continue
		}
		tokens := strings.Split(line, "|")
		mean, err := strconv.ParseFloat(tokens[1], 64)
		if err != nil {
			return nil, time.Now(), err
		}
		median, err := strconv.ParseFloat(tokens[2], 64)
		if err != nil {
			return nil, time.Now(), err
		}
		stdev, err := strconv.ParseFloat(tokens[3], 64)
		if err != nil {
			return nil, time.Now(), err
		}
		stats = append(stats, &models.ArenaProfitStatistics{
			Arena:             strings.TrimSpace(tokens[0]),
			Mean:              mean,
			Median:            median,
			StandardDeviation: stdev,
		})
	}
	return stats, expiry, nil
}

type ArenaProfitStatisticsEstimator struct{}

func (statsEstimator *ArenaProfitStatisticsEstimator) generateDrops(arenaToGenerate string) (map[string]*models.BattledomeDrops, error) {
	estimator := new(DropRateEstimator)
	cache := caches.GetItemPriceCacheInstance()
	defer cache.Close()

	itemWeights, err := new(ItemWeightParser).Parse(constants.GetItemWeightsFilePath())
	if err != nil {
		return nil, err
	}

	drops := map[string]*models.BattledomeDrops{}
	for _, arena := range constants.ARENAS {
		if arenaToGenerate != "" && arenaToGenerate != arena {
			continue
		}

		relevantItemWeights := helpers.Filter(itemWeights, func(itemWeight models.ItemWeight) bool {
			return itemWeight.Arena == arena
		})

		items := map[string]*models.BattledomeItem{}
		slog.Info(fmt.Sprintf("Generating arena statistics for %s", arena))
		itemNames := estimator.GenerateItems(relevantItemWeights, constants.NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_PROFIT_STATISTICS)
		for _, generatedItem := range itemNames {
			item, isInItems := items[generatedItem]
			if !isInItems {
				items[generatedItem] = &models.BattledomeItem{
					Name:            generatedItem,
					Quantity:        1,
					IndividualPrice: cache.GetPrice(generatedItem),
				}
			} else {
				item.Quantity++
			}
		}

		drops[arena] = models.NewBattledomeDrops()
		drops[arena].Metadata = models.DropsMetadata{
			Source:     "(generated)",
			Arena:      arena,
			Challenger: "(generated)",
			Difficulty: "(generated)",
		}
		drops[arena].Items = items
	}

	return drops, nil
}

func generateStatistics(arena string, items map[string]*models.BattledomeItem) (*models.ArenaProfitStatistics, error) {
	arenaStats := &models.ArenaProfitStatistics{}
	profitData := []float64{}
	for _, item := range items {
		for j := 0; j < int(item.Quantity); j++ {
			profitData = append(profitData, item.IndividualPrice)
		}
	}

	mean, err := stats.Mean(profitData)
	if err != nil {
		return nil, err
	}
	median, err := stats.Median(profitData)
	if err != nil {
		return nil, err
	}
	stdev, err := stats.StandardDeviationSample(profitData)
	if err != nil {
		return nil, err
	}

	arenaStats.Arena = arena
	arenaStats.Mean = mean
	arenaStats.Median = median
	arenaStats.StandardDeviation = stdev

	return arenaStats, nil
}

func (estimator *ArenaProfitStatisticsEstimator) Estimate() (map[string]*models.ArenaProfitStatistics, error) {
	drops := map[string]*models.BattledomeDrops{}
	for _, arena := range constants.ARENAS {
		if helpers.IsFileExists(constants.GetGeneratedDropsFilePath(arena)) {
			parsedDrops, err := new(GeneratedDropsParser).Parse(constants.GetGeneratedDropsFilePath(arena))
			if err != nil {
				return nil, err
			}

			if len(parsedDrops) > 1 {
				panic(fmt.Errorf("encountered mixed arena data in generated drops; there should only be a single arena's data per file"))
			}

			for parsedArena, parsedArenaDrops := range parsedDrops {
				drops[parsedArena] = parsedArenaDrops
			}
		} else {
			generatedDrops, err := estimator.generateDrops(arena)
			if err != nil {
				return nil, err
			}

			err = new(GeneratedDropsParser).Save(generatedDrops, constants.GetGeneratedDropsFilePath(arena))
			if err != nil {
				return nil, err
			}

			drops = generatedDrops
		}
	}

	stats := map[string]*models.ArenaProfitStatistics{}
	for arena, dropsByArena := range drops {
		arenaStats, err := generateStatistics(arena, dropsByArena.Items)
		if err != nil {
			return nil, err
		}
		stats[arena] = arenaStats
	}

	return stats, nil
}
