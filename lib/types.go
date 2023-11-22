package lib

// FileLocation specifies the line and column number location in a file.
type FileLocation struct {
	Line, Col int
}

// FileBlock specifies the start and end of a segment in a file.
type FileBlock struct {
	Start, End FileLocation
}

// Stats holds coverage related statistics.
type Stats struct {
	NumStatements int
	HitCount      int
	Coverage      float64
}

// Coverage holds global coverage statistics.
type Coverage struct {
	Stats
	Files map[string]map[FileBlock]*Stats
}

// NewCoverage returns a new empty Coverage instance.
func NewCoverage() *Coverage {
	return &Coverage{Files: make(map[string]map[FileBlock]*Stats)}
}
