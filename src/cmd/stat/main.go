package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"github.com/gosuri/uitable"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
)

type linterOutput struct {
	Reports []*linter.Report
	Errors  []string
}

type ReportCount struct {
	CheckName string
	Count     int
	Added     int
	Deleted   int
}

type ReportDiff struct {
	Report  *linter.Report
	New     bool
	Deleted bool
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

func main() {
	var markdown bool
	var newReportsPath string
	var oldReportsPath string
	flag.BoolVar(&markdown, "m", false, "print with markdown table")
	flag.StringVar(&newReportsPath, "new", "reports.json", "reports from current branch")
	flag.StringVar(&oldReportsPath, "old", "reports-master.json", "reports from master branch")
	flag.Parse()

	reports := loadReportsFile(newReportsPath).Reports
	reportsMaster := loadReportsFile(oldReportsPath).Reports

	diff := reportsDiff(reports, reportsMaster)

	diffByType := getReportsDiffByType(diff)
	masterCountByType := getReportsByType(reportsMaster)
	diffSorted := getSortedDiffSlice(diffByType)

	if markdown {
		markdownDiff := getMarkdownReportDiff(diff)

		fmt.Println("## Changes in reports with the current master.")

		if len(diffSorted) == 0 {
			fmt.Println("No changes.")
		} else {
			fmt.Println("Name | Count | New | Deleted")
			fmt.Println("---- | :---: | :-: | :-----:")

			for _, info := range diffSorted {
				fmt.Printf("%s | %d | %d | %d\n", info.CheckName, masterCountByType[info.CheckName], info.Added, info.Deleted)
			}

			fmt.Println(markdownDiff)
		}
	} else {
		strDiff := getReportDiff(diff)

		table := uitable.New()

		table.AddRow("Name", "Count", "Added", "Deleted")

		for _, info := range diffSorted {
			table.AddRow(info.CheckName, masterCountByType[info.CheckName], info.Added, info.Deleted)
		}

		fmt.Println("Changes in reports with the current master.")
		fmt.Println()
		fmt.Println(table)
		fmt.Println()
		fmt.Println(strDiff)
	}
}

func getSortedDiffSlice(diffByType map[string]*ReportCount) []*ReportCount {
	var diffSorted []*ReportCount
	for _, reportCount := range diffByType {
		diffSorted = append(diffSorted, reportCount)
	}
	sort.Slice(diffSorted, func(i, j int) bool {
		if diffSorted[i].Added == diffSorted[j].Added {
			return diffSorted[i].Deleted > diffSorted[j].Deleted
		}

		return diffSorted[i].Deleted > diffSorted[j].Deleted
	})
	return diffSorted
}

func getReportDiff(diff []ReportDiff) string {
	var diffReportsString string

	for _, reportDiff := range diff {
		addOrDeleteSymbol := "-"
		if reportDiff.New {
			addOrDeleteSymbol = "+"
		}

		formattedReport := cmd.FormatReport(reportDiff.Report)
		formattedReportParts := strings.Split(formattedReport, "\n")
		formattedReport = strings.Join(formattedReportParts, "\n"+addOrDeleteSymbol)
		formattedReport = addOrDeleteSymbol + formattedReport + "\n"

		diffReportsString += formattedReport
	}

	return diffReportsString
}

func getMarkdownReportDiff(diff []ReportDiff) string {
	return fmt.Sprintf("\n<details>\n"+
		"<summary>%s</summary>\n\n"+
		"```diff\n"+
		"%s"+
		"```\n\n"+
		"</details>", "diff", getReportDiff(diff))
}

func reportsDiff(now, master []*linter.Report) (diff []ReportDiff) {
	for _, reportFromNow := range now {
		var found bool
		for _, reportFromMaster := range master {
			if reportFromMaster.Context == reportFromNow.Context && reportFromMaster.Message == reportFromNow.Message {
				found = true
			}
		}

		if !found {
			diff = append(diff, ReportDiff{
				Report:  reportFromNow,
				New:     true,
				Deleted: false,
			})
		}
	}
	for _, reportFromMaster := range master {
		var found bool
		for _, reportFromNow := range now {
			if reportFromMaster.Context == reportFromNow.Context && reportFromMaster.Message == reportFromNow.Message {
				found = true
			}
		}

		if !found {
			diff = append(diff, ReportDiff{
				Report:  reportFromMaster,
				New:     false,
				Deleted: true,
			})
		}
	}

	return diff
}

func getReportsByType(reports []*linter.Report) map[string]int {
	reportsByType := map[string]int{}
	for _, report := range reports {
		reportsByType[report.CheckName]++
	}
	return reportsByType
}

func getReportsDiffByType(reportDiffs []ReportDiff) map[string]*ReportCount {
	reportsByType := map[string]*ReportCount{}
	for _, diff := range reportDiffs {
		rep, ok := reportsByType[diff.Report.CheckName]
		if !ok {
			info := &ReportCount{
				CheckName: diff.Report.CheckName,
			}
			if diff.New {
				info.Added++
			} else if diff.Deleted {
				info.Deleted++
			}

			reportsByType[diff.Report.CheckName] = info
		} else {
			if diff.New {
				rep.Added++
			} else if diff.Deleted {
				rep.Deleted++
			}
		}
	}
	return reportsByType
}
