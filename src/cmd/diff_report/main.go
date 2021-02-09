package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
)

type linterOutput struct {
	Reports []*linter.Report
	Errors  []string
}

func loadReportsFile(filename string) *linterOutput {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("read reports file: %v", err)
	}
	var output linterOutput
	if err := json.Unmarshal(data, &output); err != nil {
		log.Fatalf("unmarshal reports file: %v", err)
	}
	return &output
}

func formatReportLines(reports []*linter.Report) []string {
	sort.SliceStable(reports, func(i, j int) bool {
		return reports[i].Filename < reports[j].Filename
	})
	var parts []string
	for _, r := range reports {
		part := strings.ReplaceAll(cmd.FormatReport(r), "\r", "")
		parts = append(parts, strings.Split(part, "\n")...)
	}
	parts = append(parts, "") // Trailing EOL
	return parts
}

func main() {
	reports1 := loadReportsFile("vk.json")
	reports2 := loadReportsFile("vk1.json")

	var diff []*linter.Report

	for _, report2 := range reports1.Reports {
		var contains bool
		for _, report1 := range reports2.Reports {
			if report1.Context == report2.Context && report1.Filename == report2.Filename {
				contains = true
				break
			}
		}

		if !contains {
			diff = append(diff, report2)
		}
	}

	for _, report := range formatReportLines(diff) {
		log.Println(report)
	}
}
