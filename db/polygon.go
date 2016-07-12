// Very heavily inspired by https://github.com/nferruzzi/gormGIS
// Therefore credit is due to nferruzzi. Thank you!
package db

import (
	"bytes"
	"fmt"
	"strings"

	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"github.com/nferruzzi/gormGIS"
)

type GeoPolygon struct {
	Points []gormGIS.GeoPoint
}

func (p *GeoPolygon) Scan(val interface{}) error {

	b, err := hex.DecodeString(string(val.([]uint8)))
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)
	var wkbByteOrder uint8
	if err := binary.Read(r, binary.LittleEndian, &wkbByteOrder); err != nil {
		return err
	}

	var byteOrder binary.ByteOrder
	switch wkbByteOrder {
	case 0:
		byteOrder = binary.BigEndian
	case 1:
		byteOrder = binary.LittleEndian
	default:
		return fmt.Errorf("Invalid byte order %d", wkbByteOrder)
	}

	var wkbGeometryType uint64
	if err := binary.Read(r, byteOrder, &wkbGeometryType); err != nil {
		return err
	}

	// TODO: Evaluate this. Might be two fields.
	var wkbTmp uint64
	if err := binary.Read(r, byteOrder, &wkbTmp); err != nil {
		return err
	}

	// Now extract the real points from well-known binary.
	var point gormGIS.GeoPoint
	polygon := make([]gormGIS.GeoPoint, 0)

	pErr := binary.Read(r, byteOrder, &point)
	for pErr == nil {
		polygon = append(polygon, point)
		pErr = binary.Read(r, byteOrder, &point)
	}

	// Save collected points in GeoPolygon structure.
	p.Points = polygon

	return nil
}

func (p GeoPolygon) Value() (driver.Value, error) {

	var polygon string

	if len(p.Points) > 0 {

		// Concatenate all points in polygon.
		for _, point := range p.Points {
			polygon = fmt.Sprintf("%s%v %v,", polygon, point.Lng, point.Lat)
		}

		// Delete last comma added above.
		polygon = strings.TrimRight(polygon, ",")

		// Format into final representation.
		polygon = fmt.Sprintf("SRID=4326;POLYGON((%v))", polygon)

		return polygon, nil
	} else {
		return nil, nil
	}
}
