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

func (parser *ArenaProfitStatisticsParser) Save(arenaStats []*models.DropsStatistics, expiry time.Time, filePath string) error {
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

func (parser *ArenaProfitStatisticsParser) Parse(filePath string) ([]*models.DropsStatistics, time.Time, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, time.Now(), fmt.Errorf("arena profit statistics file does not exist! The file path provided was \"%s\"", filePath)
	}

	stats := []*models.DropsStatistics{}
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
		stats = append(stats, &models.DropsStatistics{
			Arena:                       strings.TrimSpace(tokens[0]),
			MeanItemProfit:              mean,
			MedianItemProfit:            median,
			ItemProfitStandardDeviation: stdev,
		})
	}
	return stats, expiry, nil
}

type DropStatisticsService struct{}

func NewDropStatisticsService() *DropStatisticsService {
	return &DropStatisticsService{}
}

func (estimator *DropStatisticsService) Estimate(drop *models.BattledomeDrops) (*models.DropsStatistics, error) {
	var arena string = drop.Metadata.Arena
	profitData := []float64{}
	for _, item := range drop.Items {
		for j := 0; j < int(item.Quantity); j++ {
			profitData = append(profitData, item.IndividualPrice)
		}
	}

	if len(profitData) == 0 {
		return &models.DropsStatistics{
			Arena:                       arena,
			MeanItemProfit:              0,
			MedianItemProfit:            0,
			ItemProfitStandardDeviation: 0,
		}, nil
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
	return &models.DropsStatistics{Arena: arena, MeanItemProfit: mean, MedianItemProfit: median, ItemProfitStandardDeviation: stdev}, nil
}
