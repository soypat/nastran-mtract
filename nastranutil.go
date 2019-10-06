package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type dims struct {
	x    float64
	y    float64
	z    float64
	t    float64
	csys int // dont think ill use it (coordinate system)
}

type node struct {
	number int
	dims
}

type element struct {
	number    int
	Type      string
	nodeIndex []int
	collector int
}
type constraint struct {
	number    int
	Type      string
	master    int
	slaves    []int
	collector int
}

//const spacedInteger string = "%d\t"
const separator string = ","
const newline string = "\n"

func writeNodosCSV(writedir string, nastrandir string) error {
	dataFile, err := os.Open(nastrandir)
	if err != nil {
		return err
	}
	defer dataFile.Close()
	nodeFile, err := os.Create(writedir)
	if err != nil {
		return err
	}
	line := 1
	nodeWriter := bufio.NewWriter(nodeFile)
	defer nodeFile.Sync()
	defer nodeFile.Close()
	scanner := bufio.NewScanner(dataFile)
	reGridStart := regexp.MustCompile(`GRID\*`)
	reNodeNumber := regexp.MustCompile(`(?:GRID\*\s+)([\d]+)`)
	reNonNumerical := regexp.MustCompile(`[A-Za-z\*\+\-\s,]+`)
	reLineContinueFlag := regexp.MustCompile(`\+{1}\n*$`)
	reFloat := regexp.MustCompile(`\d{1}\.\d{4,16}E{1}[\+\-]{1}\d{2}`)
	var nodeNumberString, currentText string
	var floatStrings []string //integerStrings
	var nodoLineText string

	for scanner.Scan() {
		var currentNode node
		line++
		if reGridStart.MatchString(scanner.Text()) {
			currentText = reLineContinueFlag.ReplaceAllString(scanner.Text(), "")
			nodeNumberString = reNodeNumber.FindString(currentText)
			currentNode.number, err = strconv.Atoi(reNonNumerical.ReplaceAllString(nodeNumberString, ""))
			scanner.Scan()
			line++
			currentText = currentText + scanner.Text()
			floatStrings = reFloat.FindAllString(currentText, 4)
			assignDims(&currentNode.dims, floatStrings)
		} else { // No es un nodo, sigo!
			continue
		}
		nodoLineText = fmt.Sprintf("%d%s %e%s %e%s %e%s %e %s", currentNode.number, separator, currentNode.x, separator, currentNode.y, separator, currentNode.z, separator, currentNode.t, newline)
		_, err = nodeWriter.WriteString(nodoLineText)
		nodeWriter.Flush()
		if err != nil {
			return fmt.Errorf("Error escribiendo nodos.")
		}
	}
	return nil
}

func assignDims(dimensions *dims, floatString []string) {
	for i := range floatString {
		switch i {
		case 0:
			dimensions.x, _ = strconv.ParseFloat(floatString[i], 64)
		case 1:
			dimensions.y, _ = strconv.ParseFloat(floatString[i], 64)
		case 2:
			dimensions.z, _ = strconv.ParseFloat(floatString[i], 64)
		case 3:
			dimensions.t, _ = strconv.ParseFloat(floatString[i], 64)
		default:
			panic("Unreachable")
		}
	}
}
