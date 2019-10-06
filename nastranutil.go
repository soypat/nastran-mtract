package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

//var elementTypes = []string{"CHEXA","CTETRA","CBEAM"}
var constraintTypes = []string{"RBE2", "RBE3", "RBE1"}
var reElementStart = regexp.MustCompile(`[A-Z]{2,7}[\s\d]+[\+\n]`)
var reBeamStart = regexp.MustCompile(`[A-Z]{2,7}[\s\d]+[.]{1}`)
var reElementType = regexp.MustCompile(`^[A-Z]{2,7}[\d]{0,2}`)
var reElemEnd = regexp.MustCompile(`PROPERTY CARDS`)
var reEOLContinue = regexp.MustCompile(`[+]{1}$`)
var reSOLContinue = regexp.MustCompile(`^[+]{1}`)
var reInteger = regexp.MustCompile(`\s+\d+[^.A-Za-z+\$]`)
var reDecimal = regexp.MustCompile(`\d{1}\.\d{4}`)
var reNonNumerical = regexp.MustCompile(`[A-Za-z\*\+\-\s,]+`)
var reNASTRANcomment = regexp.MustCompile(`^[$]`)

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
	number      int
	Type        string
	nodeIndex   []int
	collector   int
	orientation [3]float64 // Almost exclusively for beams
}

func NewElement() element {
	return element{}
}

type collectors map[string]int

type constraint struct {
	number    int
	Type      string
	master    int
	slaves    []int
	dof       string
	collector int
}

func NewConstraint() constraint {
	return constraint{}
}

//const spacedInteger string = "%d\t"
const separator string = ","
const newline string = "\n"

func readMeshCollectors(nastrandir string) (collectors, error) {
	dataFile, err := os.Open(nastrandir)
	if err != nil {
		return collectors{}, err
	}
	defer dataFile.Close()
	scanner := bufio.NewScanner(dataFile)
	meshcol := collectors{}
	reElemStart := regexp.MustCompile(`ELEMENT CARDS`)
	var line = 0
	for scanner.Scan() { // Saltea las lineas hasta el comienzo de los elementos

		line++
		if reElemStart.MatchString(scanner.Text()) {
			break
		}

	}
	reMeshCol := regexp.MustCompile(`\$\*\s{2}Mesh Collector:\s{1}`)
	for scanner.Scan() {
		text := scanner.Text()
		text = text + ""
		line++
		if reElemEnd.MatchString(scanner.Text()) {
			return meshcol, nil
		} // Si encontramos con el final devolvemos lo que encontramos
		if reMeshCol.MatchString(scanner.Text()) { // I first find my mesh collector here
			collectorName := scanner.Text()[20:]
			// meshcol.tags =append(meshcol.tags, scanner.Text()[20:])
			currentConstraint, currentElement, linesRead, err := readNextElement(scanner)
			if err != nil {
				return meshcol, err
			}
			line += linesRead
			if currentConstraint.slaves != nil {
				meshcol[collectorName] = currentConstraint.collector
				continue
			}
			if currentElement.nodeIndex != nil && currentConstraint.slaves == nil {
				meshcol[collectorName] = currentElement.collector
			} else {
				break
			} //We done looking for collectors
		}
	}
	return meshcol, nil
}

func writeMeshCollector(nastrandir string, Acol string, numbering int) error {
	dataFile, err := os.Open(nastrandir)
	if err != nil {
		return err
	}
	//writers := make(map[string]*bufio.Writer)

	defer dataFile.Close()
	scanner := bufio.NewScanner(dataFile)
	reCollectorName := regexp.MustCompile(Acol)
	line := 0
	for scanner.Scan() {
		text := scanner.Text()
		text = text + ""
		line++
		if reCollectorName.MatchString(scanner.Text()) {
			for scanner.Scan() {
				line++
				if reNASTRANcomment.MatchString(scanner.Text()) {
					break
				}
				//TODO do the map[string]writer thingamagig

			}

		}

	}
	//nodeFile, err := os.Create(writedir)
	return nil
}

func readNextElement(scanner *bufio.Scanner) (*constraint, *element, int, error) {
	eleitem := NewElement()
	constraintitem := NewConstraint()
	linesScanned := 0
	var err error
	var eleType, eleLine string
	eleType = reElementType.FindString(scanner.Text()) // Tal vez estoy parado sobre comienzo de un elemento
	if reElemEnd.MatchString(scanner.Text()) {
		return &constraintitem, &eleitem, linesScanned, nil
	}
	if eleType == "" {
		for scanner.Scan() {
			//text := scanner.Text()
			//text = text+""
			linesScanned++
			eleType = reElementType.FindString(scanner.Text())
			if reElemEnd.MatchString(scanner.Text()) {
				return &constraintitem, &eleitem, linesScanned, nil
			} // Verfificacion EOF
			if eleType == "" {
				continue
			} else {
				break
			} // Si no encuentro un elemento, sigo!
		}
	}
	eleLine = scanner.Text()
	if reEOLContinue.MatchString(eleLine) {
		for scanner.Scan() {
			linesScanned++
			eleLine = reEOLContinue.ReplaceAllString(eleLine, "") + reSOLContinue.ReplaceAllString(scanner.Text(), "")
			if reEOLContinue.MatchString(eleLine) {
				continue
			} else {
				break
			} // Llegado a este punto arme eleLine!
		}
	}
	for _, v := range constraintTypes {
		if v == eleType { //if true we are dealing with a constraint
			integerStrings := reInteger.FindAllString(eleLine, -1)
			constraintitem.number, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[0], ""))
			if err != nil {
				return &constraintitem, &eleitem, linesScanned, err
			}
			constraintitem.Type = v
			constraintitem.master, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[1], ""))
			if err != nil {
				return &constraintitem, &eleitem, linesScanned, err
			}
			constraintitem.dof = integerStrings[2]
			constraintitem.slaves, err = Aslicetoi(integerStrings[3:])
			return &constraintitem, &eleitem, linesScanned, err
		}
	}
	if !reElementStart.MatchString(eleLine) && reBeamStart.MatchString(eleLine) { // We are dealing with a beam
		integerStrings := reInteger.FindAllString(eleLine, -1)
		decimalStrings := reDecimal.FindAllString(eleLine, -1)
		eleitem.number, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[0], ""))
		if err != nil {
			return &constraintitem, &eleitem, linesScanned, err
		}
		if len(decimalStrings) != 3 {
			return &constraintitem, &eleitem, linesScanned, fmt.Errorf("Error leyendo orientación del elemento %d. Se esperaban 3 numeros", eleitem.number)
		}
		orient1, err := strconv.ParseFloat(decimalStrings[0], 64)
		if err != nil {
			return &constraintitem, &eleitem, linesScanned, fmt.Errorf("Error leyendo orientación del elemento %d. Orientacion 1 mal parseada.", eleitem.number)
		}
		orient2, err := strconv.ParseFloat(decimalStrings[1], 64)
		if err != nil {
			return &constraintitem, &eleitem, linesScanned, fmt.Errorf("Error leyendo orientación del elemento %d. Orientacion 2 mal parseada.", eleitem.number)
		}
		orient3, err := strconv.ParseFloat(decimalStrings[2], 64)
		if err != nil {
			return &constraintitem, &eleitem, linesScanned, fmt.Errorf("Error leyendo orientación del elemento %d. Orientacion 3 mal parseada.", eleitem.number)
		}
		eleitem.orientation = [3]float64{orient1, orient2, orient3}
		eleitem.collector, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[1], ""))
		if err != nil {
			return &constraintitem, &eleitem, linesScanned, err
		}
		eleitem.Type = "CBEAM"
		eleitem.nodeIndex, err = Aslicetoi(integerStrings[2:4])

		return &constraintitem, &eleitem, linesScanned, err

	}
	integerStrings := reInteger.FindAllString(eleLine, -1)
	integerSlice, err := Aslicetoi(integerStrings)
	if err != nil {
		return &constraintitem, &eleitem, linesScanned, err
	}
	eleitem.number = integerSlice[0]
	eleitem.Type = eleType
	eleitem.collector = integerSlice[1]
	eleitem.nodeIndex = integerSlice[2:]
	return &constraintitem, &eleitem, linesScanned, err
}

func Aslicetoi(stringSlice []string) ([]int, error) {
	var intSlice []int
	for _, v := range stringSlice {
		item, err := strconv.Atoi(reNonNumerical.ReplaceAllString(v, ""))
		if err != nil {
			return nil, err
		}
		intSlice = append(intSlice, item)
	}
	return intSlice, nil
}

func (meshcol collectors) KeySlice() []string {
	keys := make([]string, 0, len(meshcol))
	for k := range meshcol {
		keys = append(keys, k)
	}
	return keys
}

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
		err2 := nodeWriter.Flush()
		if err != nil || err2 != nil {
			return fmt.Errorf("Error escribiendo nodos. Verificar nodos.csv")
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
