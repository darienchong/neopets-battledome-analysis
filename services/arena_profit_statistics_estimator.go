package services

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/montanaflynn/stats"
)

type ArenaProfitStatisticsParser struct {
	GeneratedDropsService *GeneratedDropsService
}

func NewArenaProfitStatisticsParser() *ArenaProfitStatisticsParser {
	return &ArenaProfitStatisticsParser{
		GeneratedDropsService: NewGeneratedDropsService(),
	}
}

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

type ArenaProfitStatisticsEstimator struct {
	GeneratedDropsService *GeneratedDropsService
}

func NewArenaProfitStatisticsEstimator() *ArenaProfitStatisticsEstimator {
	return &ArenaProfitStatisticsEstimator{
		GeneratedDropsService: NewGeneratedDropsService(),
	}
}

func generateStatistics(arena string, items map[string]*models.BattledomeItem) (*models.ArenaProfitStatistics, error) {

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

	return &models.ArenaProfitStatistics{
		Arena:             arena,
		Mean:              mean,
		Median:            median,
		StandardDeviation: stdev,
	}, nil
}

func (estimator *ArenaProfitStatisticsEstimator) Estimate() (map[string]*models.ArenaProfitStatistics, error) {
	drops := map[string]*models.BattledomeDrops{}
	for _, arena := range constants.ARENAS {
		dropsByArena, err := estimator.GeneratedDropsService.GetDrops(arena)
		if err != nil {
			return nil, err
		}
		drops[arena] = dropsByArena
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
