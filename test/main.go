package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected a filename argument")
	}
	fp, _ := os.Open(os.Args[1])
	var nast nastran
	err := nast.parse(fp)
	if err != nil {
		log.Fatal(err)
	}
	fpc, _ := os.Create("RBE2.csv")
	for _, conn := range nast.connectivities {
		fmt.Fprintf(fpc, "%d,%d", conn.collector, conn.main)
		for _, d := range conn.affected {
			fmt.Fprintf(fpc, ",%d", d)
		}
		fmt.Fprintln(fpc)
	}
	fpc.Close()
	fpc, _ = os.Create("forces.csv")
	for _, force := range nast.forces {
		fmt.Fprintf(fpc, "%d,%d,%g,%g,%g,%g\n", force.collector, force.affected, force.magnitude,
			force.orientation[0], force.orientation[1], force.orientation[2])
	}
	fpc.Close()
	recollector := newNumerator()
	var fullElemName = func(el element) string {
		col := nast.elemCollector.name(el.collector)
		if col == "" {
			return ""
		}
		return fmt.Sprintf("%s%d-%d", strings.TrimSpace(el.Type), len(el.nodeIndex), el.collector)
	}
	for _, el := range nast.elements {
		colName := nast.elemCollector.name(el.collector)
		if colName != "" && recollector.name(el.collector) == "" {
			recollector.addDirect(el.collector, fullElemName(el))
		}
	}

	for num, name := range recollector.direct {
		fpc, _ = os.Create(name)
		for _, el := range nast.elements {
			fullname := fullElemName(el)
			if num != el.collector || name != fullname {
				continue
			}
			info := elemInfo[el.Type]
			fmt.Fprintf(fpc, "%d", num)
			for i := range el.nodeIndex {
				fmt.Fprintf(fpc, ",%d", el.nodeIndex[info.ordering[i]])
			}
			if el.orientation != [3]float64{} {
				for _, d := range el.orientation {
					fmt.Fprintf(fpc, ",%g", d)
				}
			}
			fmt.Fprintln(fpc)
		}
	}
	fpc.Close()
	fpc, _ = os.Create("spcs.csv")
	for _, spc := range nast.spcs {
		fmt.Fprintf(fpc, "%d,%d,%d,%g\n", spc.collector, spc.affected, spc.dofs, spc.last)
	}
	fpc.Close()
	fpc, _ = os.Create("nodos.csv")
	for _, nod := range nast.nodes {
		fmt.Fprintf(fpc, "%d,%g,%g,%g,%g,%d\n", nod.number, nod.x, nod.y, nod.z, nod.t, nod.csys)
	}
	fpc.Close()

}

func parseGrid(item string) (node, error) {
	if len(item) < 104 {
		return node{}, fmt.Errorf("grid item string too short (%d)", len(item))
	}
	nn, err := strconv.Atoi(trimFortran(item[16:24]))
	if err != nil {
		return node{}, fmt.Errorf("node number: %s", err)
	}
	const (
		dimOffset = 40
		f64Len    = 16
	)
	x, err := strconv.ParseFloat(trimFortran(item[dimOffset:dimOffset+f64Len]), 64)
	if err != nil {
		return node{}, fmt.Errorf("x dim: %s", err)
	}
	y, err := strconv.ParseFloat(trimFortran(item[dimOffset+f64Len:dimOffset+2*f64Len]), 64)
	if err != nil {
		return node{}, fmt.Errorf("y dim: %s", err)
	}
	z, err := strconv.ParseFloat(trimFortran(item[dimOffset+2*f64Len:dimOffset+3*f64Len]), 64)
	if err != nil {
		return node{}, fmt.Errorf("z dim: %s", err)
	}
	t, err := strconv.ParseFloat(trimFortran(item[dimOffset+3*f64Len:dimOffset+4*f64Len]), 64)
	if err != nil {
		return node{}, fmt.Errorf("t dim: %s", err)
	}
	return node{
		number: nn,
		dims:   dims{x: x, y: y, z: z, t: t},
	}, nil
}

func parseElement(item string, oriented bool) (element, error) {
	if len(item) < 24 {
		return element{}, fmt.Errorf("element item too short (%d)", len(item))
	}
	nodStrLen := len(item) - 24
	if nodStrLen%8 != 0 || nodStrLen == 0 {
		return element{}, fmt.Errorf("length of node items must be multiple of 8. got length %d", nodStrLen)
	}
	en, err := strconv.Atoi(trimFortran(item[8:16]))
	if err != nil {
		return element{}, fmt.Errorf("element number: %s", err)
	}
	collector, err := strconv.Atoi(trimFortran(item[16:24]))
	if err != nil {
		return element{}, fmt.Errorf("collector number: %s", err)
	}
	var nodes int
	if oriented {
		nodes = (nodStrLen - 24) / 8
	} else {
		nodes = nodStrLen / 8
	}
	var enodes []int
	for iele := 0; iele < nodes; iele++ {
		offset := 24 + iele*8
		elNod, err := strconv.Atoi(trimFortran(item[offset : offset+8]))
		if err != nil {
			return element{}, fmt.Errorf("parsing %dth node of element: %s", iele, err)
		}
		enodes = append(enodes, elNod)
	}
	el := element{
		number:    en,
		nodeIndex: enodes,
		collector: collector,
	}
	if oriented {
		offset := 24 + nodes*8
		el.orientation, err = parseSmallOrientation(item[offset : offset+24])
		if err != nil {
			return element{}, err
		}
	}
	return el, nil
}

func init() {
	items := []string{gridItem}
	for i := range items {
		if len(items[i]) != 8 {
			panic("bad item length. must be 8")
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
	nodeIndex   []int
	collector   int
	orientation [3]float64 // Almost exclusively for beams and shells
	Type        string
}

type connectivity struct {
	main      int
	affected  []int
	collector int
	Type      string
}

type force struct {
	affected    int
	collector   int
	csys        int
	magnitude   float64
	orientation [3]float64
}

type spc struct {
	collector int
	affected  int
	dofs      int
	last      float64
}

func parseSPC(item string) (spc, error) {
	if len(item) < 40 {
		return spc{}, fmt.Errorf("expected item length of at least 40, got %d", len(item))
	}
	collector, err := strconv.Atoi(trimFortran(item[8:16]))
	if err != nil {
		return spc{}, fmt.Errorf("parsing collector number: %s", err)
	}
	affected, err := strconv.Atoi(trimFortran(item[16:24]))
	if err != nil {
		return spc{}, fmt.Errorf("parsing node number: %s", err)
	}
	dofs, err := strconv.Atoi(trimFortran(item[24:32]))
	if err != nil {
		return spc{}, fmt.Errorf("parsing dof number: %s", err)
	}
	smol, err := parseSmallFortranFloat(trimFortran(item[32:40]))
	if err != nil {
		return spc{}, fmt.Errorf("parsing last number: %s", err)
	}
	return spc{
		collector: collector,
		affected:  affected,
		dofs:      dofs,
		last:      smol,
	}, nil
}

func trimFortran(item string) string {
	return strings.TrimLeft(item, " ")
}

func parseSmallOrientation(items string) (a [3]float64, err error) {
	if len(items) != 24 {
		return a, fmt.Errorf("orientation parsing expected length %d, got %d", 24, len(items))
	}
	for i := 0; i < 3; i++ {
		offset := i * 8
		a[i], err = parseSmallFortranFloat(trimFortran(items[offset : offset+8]))
		if err != nil {
			return a, fmt.Errorf("parsing orientation %dth value: %s", i, err)
		}
	}
	return a, nil
}

func parseSmallFortranFloat(item string) (f float64, err error) {
	isNegative := item[0] == '-'
	if isNegative {
		item = item[1:]
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("parsing small fortran float %q: %s", item, err)
		}
		if isNegative {
			f = -f
		}
	}()
	if !strings.Contains(item, "E") && (strings.Index(item, "-") > 0 || strings.Index(item, "+") > 0) {
		num, exp, found := strings.Cut(item, "-")
		if found {
			f, err = strconv.ParseFloat(strings.Join([]string{num, exp}, "E-"), 32)
			return f, err
		}
		num, exp, found = strings.Cut(item, "+")
		if found {
			f, err = strconv.ParseFloat(strings.Join([]string{num, exp}, "E"), 32)
			return f, err
		}
	}
	return strconv.ParseFloat(item, 32)
}

func parseForce(item string) (force, error) {
	if len(item) < 64 {
		return force{}, fmt.Errorf("expected item length of at least 64, got %d", len(item))
	}
	collector, err := strconv.Atoi(trimFortran(item[8:16]))
	if err != nil {
		return force{}, fmt.Errorf("parsing collector number: %s", err)
	}
	node, err := strconv.Atoi(trimFortran(item[16:24]))
	if err != nil {
		return force{}, fmt.Errorf("parsing node number: %s", err)
	}
	csys, err := strconv.Atoi(trimFortran(item[24:32]))
	if err != nil {
		return force{}, fmt.Errorf("parsing csys number: %s", err)
	}
	magnitude, err := parseSmallFortranFloat(trimFortran(item[32:40]))
	if err != nil {
		return force{}, fmt.Errorf("parsing magnitude: %s", err)
	}
	orientation, err := parseSmallOrientation(item[40 : 40+24])
	if err != nil {
		return force{}, err
	}
	return force{
		collector:   collector,
		affected:    node,
		magnitude:   magnitude,
		csys:        csys,
		orientation: orientation,
	}, nil
}

func parseConnectivity(item string) (connectivity, error) {
	if len(item) < 24 {
		return connectivity{}, fmt.Errorf("expected item length of at least 24, got %d", len(item))
	}
	m, err := strconv.Atoi(trimFortran(item[8:16]))
	if err != nil {
		return connectivity{}, fmt.Errorf("parsing main node: %s", err)
	}
	affectedNum := (len(item) - 24) / 8
	var affected []int
	for i := 0; i < affectedNum; i++ {
		offset := 16 + i*8
		s, err := strconv.Atoi(trimFortran(item[offset : offset+8]))
		if err != nil {
			return connectivity{}, fmt.Errorf("parsing affected node %d: %s", i, err)
		}
		affected = append(affected, s)
	}
	return connectivity{main: m, affected: affected}, nil
}

type numerator struct {
	direct map[int]string
	ind    map[string]int
	max    int
}

func newNumerator() numerator {
	return numerator{
		direct: make(map[int]string),
		ind:    make(map[string]int),
	}
}

func (c *numerator) addDirect(num int, name string) {
	if name == "" {
		panic("attempted to add empty collector name")
	}
	if c.number(name) >= 0 {
		panic("attempted to add existing collector")
	}
	if num > c.max {
		c.max = num
	}
	c.direct[num] = name
	c.ind[name] = num
}

func (c *numerator) addIndirect(name string) int {
	num := c.max + 1
	c.addDirect(num, name)
	return num
}

// number returns existing collector number or -1.
func (c *numerator) number(name string) int {
	if name == "" {
		return -1
	}
	num, exist := c.ind[name]
	if !exist {
		return -1
	}
	return num
}

// name returns existing collector name or "".
func (c *numerator) name(num int) string {
	if num < 0 {
		return ""
	}
	return c.direct[num]
}

var (
	collectorLabels = [][]byte{
		[]byte("$*  Mesh Collector: "),
		[]byte("$*  Constraint: "),
		[]byte("$*  Load: "),
	}
)

func parseCollector(line []byte) (string, bool) {
	if len(line) < 5 || line[0] != '$' || line[1] != '*' {
		return "", false
	}
	for _, label := range collectorLabels {
		if bytes.HasPrefix(line, label) {
			return string(line[len(label):]), true
		}
	}
	return "", false
}
