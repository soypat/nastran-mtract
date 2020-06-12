# NASTRAN Mesh Extractor
### Mesh wrangler for NASTRAN finite element files.

Purpose is to extract node position/numbering to nodos.csv file and generate element index files for easy Matlab processing. 
~Only manages to extract nodes for now.~ Fully working executable in [whittileaks.com](https://sites.google.com/site/whittileak/home) under _Elementos Finitos_.

## Instructions
Run executable in directory above that of the NASTRAN file, which should be in .dat format.

Follow on screen instructions (spanish). Use arrow-keys to navigate and press [ENTER]/[INTRO] 
to select option.

Boundary conditions (SPC) will be printed with a zero column which does not correspond to any value.

## Outputs

### nodos.csv

nodeNumber,  x,  y,  z,  csys/t

### ElementType(NumberofNodes)-(CollectorNumber).csv

elementNumber, Node1, Node2, Node3, ... ,  NodeN

### ConstraintType-(CollectorNumber).csv

affectedDOF , ElementNumber,  NodeMaster, NodeSlave1, NodeSlave2, ... ,  NodeSlaveN

### BoundaryCondition-(CollectorNumber).csv

AffectedDOF , 0,  NodeNumber

### Force-(CollectorNumber).csv

csys , AffectedNode,  x-Force, y-Force, z-Force



Written in Go.
