package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type einfo struct {
	ordering    []int
	itemNameLen int
	oriented    bool
}

var elemInfo = map[string]einfo{
	ctriaItem: {
		ordering:    []int{0, 1, 2},
		itemNameLen: strings.Index(ctriaItem, " "),
	},
	ctria6Item: {
		ordering:    []int{0, 1, 2, 3, 4, 5},
		itemNameLen: strings.Index(ctria6Item, " "),
	},
	chexa20Item: {
		ordering:    []int{5, 1, 2, 6, 4, 0, 3, 7, 13, 9, 14, 17, 12, 11, 15, 19, 16, 8, 10, 18},
		itemNameLen: strings.Index(chexa20Item, " "),
	},
	chexaItem: {
		ordering:    []int{5, 1, 2, 6, 4, 0, 5, 7},
		itemNameLen: strings.Index(chexaItem, " "),
	},
	ctetraItem: {
		ordering:    []int{0, 1, 2, 3, 4, 6, 7, 5, 9, 8},
		itemNameLen: strings.Index(ctetraItem, " "),
	},
	cbarItem: {
		oriented:    true,
		ordering:    []int{0, 1},
		itemNameLen: strings.Index(cbarItem, " "),
	},
	cbeamItem: {
		oriented:    true,
		ordering:    []int{0, 1},
		itemNameLen: strings.Index(cbeamItem, " "),
	},
}

const (
	ctriaItem   = "CTRIA   "
	ctria6Item  = "CTRIA6  "
	gridItem    = "GRID*   "
	ctetraItem  = "CTETRA  " // 10 node tetrahedron
	cbarItem    = "CBAR    "
	chexaItem   = "CHEXA   " // 8 node  hexahedron
	chexa20Item = "CHEXA20 " // 20 node hexahedron
	cbeamItem   = "CBEAM   "
	rbe2Item    = "RBE2    "
	rbe3Item    = "RBE3    "
	forceItem   = "FORCE   "
	spcItem     = "SPC     "
)

type nastran struct {
	nodes          []node
	elements       []element
	elemCollector  numerator
	connectivities []connectivity
	connCollector  numerator
	forces         []force
	forceCollector numerator
	spcs           []spc
	spcsCollector  numerator
}

func (n *nastran) parse(r io.Reader) error {
	n.reset()
	scan := bufio.NewScanner(r)
	var entity bytes.Buffer
	lineNo := 0
	entityStart := -1
	entityReset := func() {
		entity.Reset()
		entityStart = -1
	}
	currentCollector := ""
	for scan.Scan() {
		lineNo++
		line := scan.Bytes()
		if len(line) < 16 || bytes.HasPrefix(line, []byte{'$'}) {
			if entity.Len() > 0 {
				// bad formatting. Invalid entity.
				entityReset()
			}
			if cname, found := parseCollector(line); found {
				currentCollector = cname
			}
			continue
		}
		elen := entity.Len()
		midEntity := bytes.HasSuffix(line, []byte{'+'})
		if elen == 0 {
			entityStart = lineNo
		}
		if midEntity && elen == 0 {
			// Case first line of entity, still needs more data
			entity.Write(line[:len(line)-1])
			continue
		} else if midEntity {
			// On line 2+ of entity, still need more data before parsing
			entity.Write(line[8 : len(line)-1])
			continue
		} else if elen == 0 {
			// case first line of entity, no more data needed
			entity.Write(line[:])
		} else {
			// case last line of entity, no more data needed
			entity.Write(line[8:])
		}

		itemStr := entity.String()
		entity := itemStr[:8]
		switch entity {
		case gridItem:
			node, err := parseGrid(itemStr)
			if err != nil {
				return fmt.Errorf("parsing Grid at line(s) %d..%d: %s", entityStart, lineNo, err)
			}
			n.nodes = append(n.nodes, node)

		case ctetraItem, cbarItem, cbeamItem:
			einfo := elemInfo[entity]
			el, err := parseElement(itemStr, len(einfo.ordering), einfo.oriented)
			if err != nil {
				name := entity[:einfo.itemNameLen]
				return fmt.Errorf("parsing %q element at line(s) %d..%d: %s", name, entityStart, lineNo, err)
			}
			el.Type = entity
			n.addElement(el, currentCollector)

		case rbe2Item, rbe3Item:
			conn, err := parseConnectivity(itemStr)
			if err != nil {
				return fmt.Errorf("parsing RBE2 connectivity at line(s) %d..%d: %s", entityStart, lineNo, err)
			}
			conn.Type = entity
			n.addConnectivity(conn, currentCollector)

		case forceItem:
			F, err := parseForce(itemStr)
			if err != nil {
				return fmt.Errorf("parsing FORCE entity at line(s) %d..%d: %s", entityStart, lineNo, err)
			}
			n.addForce(F, currentCollector)

		case spcItem:
			s, err := parseSPC(itemStr)
			if err != nil {
				return fmt.Errorf("parsing SPC entity at line(s) %d..%d: %s", entityStart, lineNo, err)
			}
			n.addSPC(s, currentCollector)
		}
		entityReset()
	}
	return nil
}

func (n *nastran) addElement(e element, collectorIfAny string) {
	if collectorIfAny == "" {
		e.collector = -1
	} else if n.elemCollector.number(collectorIfAny) < 0 {
		n.elemCollector.addDirect(e.collector, collectorIfAny)
	}
	n.elements = append(n.elements, e)
}

func (n *nastran) addConnectivity(conn connectivity, collectorIfAny string) {
	if collectorIfAny != "" {
		num := n.connCollector.number(collectorIfAny)
		if num < 0 {
			// does not exist so we add it.
			num = n.connCollector.addIndirect(collectorIfAny)
		}
		conn.collector = num
	} else {
		conn.collector = -1
	}
	n.connectivities = append(n.connectivities, conn)
}

func (n *nastran) reset() {
	n.elemCollector = newNumerator()
	n.connCollector = newNumerator()
	n.forceCollector = newNumerator()
	n.spcsCollector = newNumerator()
	n.elements = nil
	n.connectivities = nil
	n.forces = nil
	n.nodes = nil
	n.spcs = nil
}

func (n *nastran) addForce(f force, collectorIfAny string) {
	if collectorIfAny != "" && n.forceCollector.number(collectorIfAny) < 0 {
		n.forceCollector.addDirect(f.collector, collectorIfAny)
	} else {
		f.collector = -1
	}
	n.forces = append(n.forces, f)
}

func (n *nastran) addSPC(s spc, collectorIfAny string) {
	if collectorIfAny != "" && n.spcsCollector.number(collectorIfAny) < 0 {
		n.spcsCollector.addDirect(s.collector, collectorIfAny)
	} else {
		s.collector = -1
	}
	n.spcs = append(n.spcs, s)
}
