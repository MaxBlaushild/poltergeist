// Package stlbbox computes an axis-aligned bounding box directly from a
// binary STL mesh. R-5.2's rule 1 ("bounding box exceeds 210mm on any axis")
// is pure geometry — it doesn't need the slicer, and computing it directly
// from the mesh we just generated is exact, not an estimate.
package stlbbox

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

type Box struct {
	MinX, MinY, MinZ float64
	MaxX, MaxY, MaxZ float64
}

func (b Box) XMm() float64 { return b.MaxX - b.MinX }
func (b Box) YMm() float64 { return b.MaxY - b.MinY }
func (b Box) ZMm() float64 { return b.MaxZ - b.MinZ }

// MaxDimensionMm is the largest of the three axis extents — what R-5.2's
// "exceeds 210mm on any axis" actually checks against.
func (b Box) MaxDimensionMm() float64 {
	return math.Max(b.XMm(), math.Max(b.YMm(), b.ZMm()))
}

const (
	headerSize  = 80
	triCountLen = 4
	triRecord   = 50 // 12 bytes normal + 3*12 bytes vertices + 2 bytes attribute
)

// FromFile parses a binary STL (the format R-2.6/R-2.7's pipeline produces —
// see generate.Render) and returns its bounding box.
func FromFile(path string) (Box, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Box{}, fmt.Errorf("stlbbox: %w", err)
	}
	return FromBytes(data)
}

func FromBytes(data []byte) (Box, error) {
	if len(data) < headerSize+triCountLen {
		return Box{}, fmt.Errorf("stlbbox: file too small to be a binary STL (%d bytes)", len(data))
	}
	count := binary.LittleEndian.Uint32(data[headerSize : headerSize+triCountLen])
	wantSize := headerSize + triCountLen + int(count)*triRecord
	if len(data) != wantSize {
		return Box{}, fmt.Errorf("stlbbox: size %d doesn't match binary STL layout for %d triangles (want %d) — is this an ASCII STL?", len(data), count, wantSize)
	}
	if count == 0 {
		return Box{}, fmt.Errorf("stlbbox: STL has zero triangles")
	}

	box := Box{
		MinX: math.Inf(1), MinY: math.Inf(1), MinZ: math.Inf(1),
		MaxX: math.Inf(-1), MaxY: math.Inf(-1), MaxZ: math.Inf(-1),
	}

	offset := headerSize + triCountLen
	for i := uint32(0); i < count; i++ {
		// Skip the 12-byte normal vector; read the 3 vertices (12 bytes each).
		base := offset + int(i)*triRecord + 12
		for v := 0; v < 3; v++ {
			vOff := base + v*12
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[vOff : vOff+4])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[vOff+4 : vOff+8])))
			z := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[vOff+8 : vOff+12])))
			box.MinX, box.MaxX = math.Min(box.MinX, x), math.Max(box.MaxX, x)
			box.MinY, box.MaxY = math.Min(box.MinY, y), math.Max(box.MaxY, y)
			box.MinZ, box.MaxZ = math.Min(box.MinZ, z), math.Max(box.MaxZ, z)
		}
	}

	return box, nil
}
