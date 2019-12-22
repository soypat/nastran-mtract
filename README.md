# NASTRAN Mesh Extractor
### Mesh wrangler for NASTRAN finite element files.

Purpose is to extract node position/numbering to nodos.csv file and generate element index files for easy Matlab processing. 
~Only manages to extract nodes for now.~ Fully working executable in [whittileaks.com](https://sites.google.com/site/whittileak/home) under _Elementos Finitos_.

## Instructions
Run executable in directory above that of the NASTRAN file, which should be in .dat format.

Follow on screen instructions (spanish). Use arrow-keys to navigate and press [ENTER]/[INTRO] 
to select option.

ADINA element node indexing only works for beams, constraints and H8, H20, T4, T10 (3-D) elements.


OUTPUT  NODES: nodos.csv

NodeNumber,  x,  y,  z,  csys/t

OUTPUT ELEMENT:  ElementType(NumberofNodes)-(CollectorNumber).csv

ElementNumber, Node1, Node2, Node3, ... ,  NodeN

OUTPUT CONSTRAINT:  ConstraintType-(CollectorName).csv

ElementNumber,  NodeMaster, NodeSlave1, NodeSlave2, ... ,  NodeSlaveN


Written in Go.
