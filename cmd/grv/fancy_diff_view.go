package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	fileAdded    = "added"
	fileModified = "modified"
	fileDeleted  = "deleted"
	fileRenamed  = "renamed"
)

var diffHeaderRegex = regexp.MustCompile(`^diff --git \w\/(.+?)(\s|\x00|$)`)
var hunkStartRegex = regexp.MustCompile(`@@ (-\d+(,\d+)?)? \+(\d+)(,\d+)? @@`)
var oldNewFileRegex = regexp.MustCompile(`^(\+\+\+|---)\s(\w\/)?`)
var oldNewModeRegex = regexp.MustCompile(`^(old|new)\smode\s`)
var renameFromToRegex = regexp.MustCompile(`^rename\s(from|to)\s`)
var trailingSpaceRegex = regexp.MustCompile(`\s+$`)

var separatorDiffLine *diffLineData

var emptyLineAdded = newDiffLineData(" ", dltLineAdded, CmpDiffviewFancyDifflineEmptyLineAdded)
var emptyLineRemoved = newDiffLineData(" ", dltLineRemoved, CmpDiffviewFancyDifflineEmptyLineRemoved)

var fileStatusStyling = map[string]ThemeComponentID{
	fileAdded:    CmpDiffviewFancyDifflineLineAddedChange,
	fileModified: CmpDiffviewFancyDiffLineFile,
	fileDeleted:  CmpDiffviewFancyDifflineLineRemovedChange,
}

func init() {
	var sections []*diffLineSection
	for i := 0; i < 1000; i++ {
		sections = append(sections, &diffLineSection{
			char:             AcsHline,
			themeComponentID: CmpDiffviewFancyDiffLineSeparator,
		})
	}

	separatorDiffLine = newSectionedDiffLineData(sections, dltNormal)
}

type fancyDiffProcessor struct{}

func (fancyDiffProcessor *fancyDiffProcessor) processDiff(lines []*diffLineData) (processedLines []*diffLineData, err error) {
	var generatedLines []*diffLineData
	var currentFile string

	for lineIndex, line := range lines {
		switch line.lineType {
		case dltGitDiffHeaderDiff:
			if generatedLines, currentFile, err = fancyDiffProcessor.processDiffHeader(lines, lineIndex); err != nil {
				return
			}

			processedLines = append(processedLines, generatedLines...)
		case dltGitDiffHeaderIndex,
			dltGitDiffHeaderNewFile,
			dltGitDiffHeaderOldFile,
			dltGitDiffHeaderNewMode,
			dltGitDiffHeaderOldMode,
			dltGitDiffHeaderNewFileMode,
			dltGitDiffHeaderDeletedFileMode,
			dltGitDiffHeaderSimilarityIndex,
			dltGitDiffHeaderRenameFrom,
			dltGitDiffHeaderRenameTo,
			dltGitDiffHeaderBinaryFile:
		case dltHunkStart:
			if generatedLines, err = fancyDiffProcessor.processHunkStart(lines, lineIndex, currentFile); err != nil {
				return
			}

			processedLines = append(processedLines, generatedLines...)
		case dltLineAdded:
			processedLines = append(processedLines, newDiffLineData(trimFirstCharacter(line.line), line.lineType, CmpDiffviewFancyDifflineLineAdded))
		case dltLineRemoved:
			processedLines = append(processedLines, newDiffLineData(trimFirstCharacter(line.line), line.lineType, CmpDiffviewFancyDifflineLineRemoved))
		case dltNormal:
			processedLines = append(processedLines, newDiffLineData(trimFirstCharacter(line.line), line.lineType, line.sections[0].themeComponentID))
		default:
			processedLines = append(processedLines, line)
		}
	}

	fancyDiffProcessor.highlightChanges(processedLines)

	return
}

func (fancyDiffProcessor *fancyDiffProcessor) processDiffHeader(lines []*diffLineData, diffHeaderIndex int) (generatedLines []*diffLineData, currentFile string, err error) {
	var oldFile, newFile, oldFileMode, newFileMode string
	var isBinary bool
	var headerSeen bool
	status := fileModified

OuterLoop:
	for lineIndex := diffHeaderIndex; lineIndex < len(lines); lineIndex++ {
		line := lines[lineIndex]

		switch line.lineType {
		case dltGitDiffHeaderDiff:
			if headerSeen {
				break OuterLoop
			}

			matches := diffHeaderRegex.FindStringSubmatch(line.line)
			if len(matches) != 3 {
				err = fmt.Errorf("line: \"%v\" doesn't have expected diff header format: %v", line.line, matches)
				return
			}

			newFile = matches[1]
			headerSeen = true
		case dltGitDiffHeaderNewFile:
			newFile = oldNewFileRegex.ReplaceAllString(line.line, "")
		case dltGitDiffHeaderOldFile:
			oldFile = oldNewFileRegex.ReplaceAllString(line.line, "")
		case dltGitDiffHeaderNewMode:
			newFileMode = oldNewModeRegex.ReplaceAllString(line.line, "")
		case dltGitDiffHeaderOldMode:
			oldFileMode = oldNewModeRegex.ReplaceAllString(line.line, "")
		case dltGitDiffHeaderNewFileMode:
			status = fileAdded
		case dltGitDiffHeaderDeletedFileMode:
			status = fileDeleted
		case dltGitDiffHeaderSimilarityIndex:
			status = fileRenamed
		case dltGitDiffHeaderRenameFrom:
			oldFile = renameFromToRegex.ReplaceAllString(line.line, "")
		case dltGitDiffHeaderRenameTo:
			newFile = renameFromToRegex.ReplaceAllString(line.line, "")
		case dltGitDiffHeaderBinaryFile:
			isBinary = true
		case dltGitDiffHeaderIndex:
		default:
			break OuterLoop
		}
	}

	if newFile == "" {
		err = fmt.Errorf("Unable to determine new file from diff headers")
		return
	}

	currentFile = newFile
	if status == fileDeleted && oldFile != "" {
		currentFile = oldFile
	}

	sections := []*diffLineSection{}

	if status == fileRenamed {
		commonPrefix, commonSuffix := determineCommonFixes(oldFile, newFile)
		oldFileLine := newDiffLineData(oldFile, dltNormal, CmpDiffviewFancyDiffLineFile)
		newFileLine := newDiffLineData(newFile, dltNormal, CmpDiffviewFancyDiffLineFile)
		highlightLine(oldFileLine, commonPrefix, commonSuffix, CmpDiffviewFancyDifflineLineRemovedChange)
		highlightLine(newFileLine, commonPrefix, commonSuffix, CmpDiffviewFancyDifflineLineAddedChange)

		sections = append(sections, &diffLineSection{
			text:             fmt.Sprintf("%v: ", status),
			themeComponentID: CmpDiffviewFancyDiffLineFile,
		})
		sections = append(sections, oldFileLine.sections...)
		sections = append(sections, &diffLineSection{
			text:             " to ",
			themeComponentID: CmpDiffviewFancyDiffLineFile,
		})
		sections = append(sections, newFileLine.sections...)
	} else {
		sections = append(sections,
			&diffLineSection{
				text:             fmt.Sprintf("%v: ", status),
				themeComponentID: CmpDiffviewFancyDiffLineFile,
			},
			&diffLineSection{
				text:             currentFile,
				themeComponentID: fileStatusStyling[status],
			})
	}

	if isBinary {
		sections = append(sections, &diffLineSection{
			text:             " (binary)",
			themeComponentID: CmpDiffviewFancyDiffLineFile,
		})
	}

	if oldFileMode != "" && newFileMode != "" {
		modeChange := fmt.Sprintf("%v changed file mode from %v to %v", currentFile, oldFileMode, newFileMode)
		generatedLines = append(generatedLines, newDiffLineData(modeChange, dltNormal, CmpDiffviewDifflineNormal))
	}

	generatedLines = append(generatedLines,
		separatorDiffLine,
		newSectionedDiffLineData(sections, dltGitDiffHeaderDiff),
		separatorDiffLine,
	)

	return
}

func (fancyDiffProcessor *fancyDiffProcessor) processHunkStart(lines []*diffLineData, hunkStartIndex int, currentFile string) (generatedLines []*diffLineData, err error) {
	hunkLine := lines[hunkStartIndex]
	matches := hunkStartRegex.FindStringSubmatch(hunkLine.line)
	if len(matches) != 5 {
		err = fmt.Errorf("Hunk start line didn't match expected format, matches: %v", matches)
		return
	}

	hunkStartLineNumber, err := strconv.Atoi(matches[3])
	if err != nil {
		err = fmt.Errorf("Failed to parse hunk start line number %v: %v", matches[3], err)
		return
	}

	var index int
	for index = hunkStartIndex + 1; index < len(lines); index++ {
		if lines[index].lineType == dltLineAdded || lines[index].lineType == dltLineRemoved {
			break
		}

		hunkStartLineNumber++
	}

	if index >= len(lines) {
		err = fmt.Errorf("Failed to find changes in hunk")
		return
	}

	hunkStartLineNumber = MaxInt(hunkStartLineNumber, 1)

	hunkParts := strings.Split(hunkLine.line, " @@")
	if len(hunkParts) != 2 {
		err = fmt.Errorf("Expected 2 hunk parts but got: %v", hunkParts)
		return
	}

	sections := []*diffLineSection{
		&diffLineSection{
			text:             fmt.Sprintf("@ %v:%v @", currentFile, hunkStartLineNumber),
			themeComponentID: CmpDiffviewDifflineHunkStart,
		},
		&diffLineSection{
			text:             hunkParts[1],
			themeComponentID: CmpDiffviewDifflineHunkHeader,
		},
	}

	generatedLines = append(generatedLines, newSectionedDiffLineData(sections, dltHunkStart))

	return
}

func (fancyDiffProcessor *fancyDiffProcessor) highlightChanges(lines []*diffLineData) {
	for lineIndex := 0; lineIndex < len(lines); lineIndex++ {
		if lines[lineIndex].lineType == dltLineRemoved {
			linesRemoved := []*diffLineData{lines[lineIndex]}
			for lineIndex++; lineIndex < len(lines) && lines[lineIndex].lineType == dltLineRemoved; lineIndex++ {
				linesRemoved = append(linesRemoved, lines[lineIndex])
			}

			var linesAdded []*diffLineData
			for ; lineIndex < len(lines) && lines[lineIndex].lineType == dltLineAdded; lineIndex++ {
				linesAdded = append(linesAdded, lines[lineIndex])
			}
			lineIndex--

			if len(linesRemoved) == len(linesAdded) {
				for i := 0; i < len(linesRemoved); i++ {
					commonPrefix, commonSuffix := determineCommonFixes(linesRemoved[i].line, linesAdded[i].line)
					if commonPrefix > 0 || commonSuffix > 0 {
						highlightLine(linesRemoved[i], commonPrefix, commonSuffix, CmpDiffviewFancyDifflineLineRemovedChange)
						highlightLine(linesAdded[i], commonPrefix, commonSuffix, CmpDiffviewFancyDifflineLineAddedChange)
					}
				}
			}
		}
	}

	for lineIndex := 0; lineIndex < len(lines); lineIndex++ {
		line := lines[lineIndex]
		if line.lineType == dltLineAdded {
			lines[lineIndex] = processWhitespaceLine(line, emptyLineAdded)
		} else if line.lineType == dltLineRemoved {
			lines[lineIndex] = processWhitespaceLine(line, emptyLineRemoved)
		}
	}
}

func processWhitespaceLine(line *diffLineData, emptyLine *diffLineData) *diffLineData {
	if line.line == "" {
		return emptyLine
	} else if matchIndexes := trailingSpaceRegex.FindStringIndex(line.line); len(matchIndexes) == 2 {
		section := line.sections[len(line.sections)-1]
		matchLength := MinInt(len(section.text), matchIndexes[1]-matchIndexes[0])

		if matchLength >= len(section.text) {
			section.themeComponentID = CmpDiffviewFancyDifflineTrailingWhitespace
		} else {
			section.text = section.text[:len(section.text)-matchLength]
			line.sections = append(line.sections, &diffLineSection{
				text:             line.line[len(line.line)-matchLength:],
				themeComponentID: CmpDiffviewFancyDifflineTrailingWhitespace,
			})
		}
	}

	return line
}

func determineCommonFixes(lineRemoved, lineAdded string) (commonPrefix, commonSuffix int) {
	removedChars := []rune(lineRemoved)
	addedChars := []rune(lineAdded)
	charLength := MinInt(len(removedChars), len(addedChars))

	for ; commonPrefix < charLength && removedChars[commonPrefix] == addedChars[commonPrefix]; commonPrefix++ {
	}

	for i, j := len(addedChars)-1, len(removedChars)-1; i > -1 && j > -1 && addedChars[i] == removedChars[j]; i, j = i-1, j-1 {
		commonSuffix++
	}

	return
}

func highlightLine(line *diffLineData, commonPrefix, commonSuffix int, highlightThemeComponentID ThemeComponentID) {
	lineString := []rune(line.line)
	lineLength := len(lineString)
	if commonPrefix+commonSuffix > lineLength {
		return
	}

	var commonPrefixString, commonSuffixString string
	if commonPrefix > 0 {
		commonPrefixString = string(lineString[:commonPrefix])
	}
	if commonSuffix > 0 {
		commonSuffixString = string(lineString[lineLength-commonSuffix:])
	}

	if !(commonPrefix == 0 || strings.TrimLeftFunc(commonPrefixString, unicode.IsSpace) != "" &&
		commonSuffix == 0 || strings.TrimRightFunc(commonSuffixString, unicode.IsSpace) != "") {
		return
	}

	sections := []*diffLineSection{}
	themeComponentID := line.sections[0].themeComponentID

	if commonPrefixString != "" {
		sections = append(sections, &diffLineSection{
			text:             commonPrefixString,
			themeComponentID: themeComponentID,
		})
	}

	sections = append(sections, &diffLineSection{
		text:             string(lineString[commonPrefix : lineLength-commonSuffix]),
		themeComponentID: highlightThemeComponentID,
	})

	if commonSuffixString != "" {
		sections = append(sections, &diffLineSection{
			text:             commonSuffixString,
			themeComponentID: themeComponentID,
		})
	}

	line.sections = sections
}

func trimFirstCharacter(line string) string {
	if line != "" {
		return line[1:]
	}

	return line
}
