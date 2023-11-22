package lib

type FileLocation struct {
	Line, Col int
}

type FileBlock struct {
	Start, End FileLocation
}

type Stats struct {
	NumStatements int
	HitCount      int
	Coverage      float64
}

type Coverage struct {
	Stats
	Files map[string]map[FileBlock]*Stats
}

func NewCoverage() *Coverage {
	return &Coverage{Files: make(map[string]map[FileBlock]*Stats)}
}
