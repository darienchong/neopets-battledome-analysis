@echo off
go test ./... -coverpkg ./... -coverprofile C:/Users/froze/OneDrive/Desktop/neopets_battledome_analysis_go/coverage.out
go tool cover -html coverage.out