package types

import "fmt"

// FileLocation specifies the line and column number location in a file.
type FileLocation struct {
	Line, Col int
}

// FileBlock specifies the start and end of a segment in a file.
type FileBlock struct {
	Start, End FileLocation
}

// UnmarshalText unmarshals a string in the form of
// "<startLine>.<startCol>,<endLine>.<endCol>".
func (fb *FileBlock) UnmarshalText(text []byte) error {
	var startLine, startCol, endLine, endCol int

	n, err := fmt.Sscanf(string(text), "%d.%d,%d.%d",
		&startLine, &startCol, &endLine, &endCol)
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("wrong file block format: %s", text)
	}

	fb.Start = FileLocation{Line: startLine, Col: startCol}
	fb.End = FileLocation{Line: endLine, Col: endCol}

	return nil
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
