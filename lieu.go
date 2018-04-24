package lieu

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"log"
	"strings"
)

func HasRequiredProperties(feature []byte) (bool, error) {

	if !HasName(feature) {
		return false, errors.New("Missing or invalid name property")
	}

	if !HasStreet(feature) {
		return false, errors.New("Missing or invalid street property")
	}

	if !HasHouseNumber(feature) {
		return false, errors.New("Missing or invalid house_number property")
	}

	if !HasCoordinates(feature) {
		return false, errors.New("Missing or invalid coordinates property")
	}

	return true, nil
}

func HasName(feature []byte) bool {

	possible_names := []string{
		"properties.name",
		"properties.wof:name",
	}

	return HasPropertyNotEmpty(feature, possible_names)
}

func HasStreet(feature []byte) bool {

	possible_streets := []string{
		"properties.addr:street",
	}

	return HasPropertyNotEmpty(feature, possible_streets)
}

func HasHouseNumber(feature []byte) bool {

	possible_housenumbers := []string{
		"properties.addr:house_number",
	}

	return HasPropertyNotEmpty(feature, possible_housenumbers)
}

func HasCoordinates(feature []byte) bool {

	geom_type := gjson.GetBytes(feature, "geometry.type")

	if !geom_type.Exists() {
		return false
	}

	if geom_type.String() != "Point" {
		return false
	}

	coords := gjson.GetBytes(feature, "geometry.coordinates")

	if !coords.Exists() {
		return false
	}

	for i, c := range coords.Array() {

		v := c.Float()

		if v == 0.0 {
			return false
		}

		if i == 0 {
			if v > 180.0 || v < -180.0 {
				return false
			}
		} else if i == 1 {
			if v > 90.0 || v < -90.0 {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func HasPropertyNotEmpty(feature []byte, possible []string) bool {

	prop, has_prop := HasProperty(feature, possible)

	if !has_prop {
		return false
	}

	if strings.Trim(prop, "") == "" {
		return false
	}

	return true
}

func HasProperty(feature []byte, possible []string) (string, bool) {

	has_property := false

	property := ""
	for _, path := range possible {

		v := gjson.GetBytes(feature, path)

		if v.Exists() {
			property = v.String()
			has_property = true
			break
		}
	}

	return property, has_property
}

// maybe probably add an Options thingy to be strict or liberal
// about whether or not to skip records that fail the required
// properties test... but not today (20180424/thisisaaronland)

func Prepare(in io.Reader, out io.Writer) error {

	reader := bufio.NewReader(in)

	sep := byte('\n') // note that double-quotes will _freak_ Go out...

	line_number := 0

	for {

		b, err := reader.ReadBytes(sep)

		if err != nil {

			if err == io.EOF {
				break
			}

			return err
		}

		line_number += 1

		ok, err := HasRequiredProperties(b)

		if !ok {
			log.Printf("%s at line number %d\n", err, line_number)
			continue
		}

		b2, err := EnstringifyProperties(b)

		if err != nil {
			return err
		}

		out.Write(b2)
	}

	return nil
}

func EnstringifyProperties(feature []byte) ([]byte, error) {

	props := gjson.GetBytes(feature, "properties")

	if !props.Exists() {
		return nil, errors.New("Feature is missing properties")
	}

	var err error

	for k, v := range props.Map() {

		if strings.HasPrefix(k, "addr:") {

			str_v := v.String()

			if k == "addr:house_number" {
				str_v = strings.Replace(str_v, " ", "", -1)
				str_v = strings.Replace(str_v, "-", "", -1)
			}

			if k == "addr:postcode" && isISO(feature, "US") && len(str_v) > 5 {
				str_v = str_v[0:5]
			}

			path := fmt.Sprintf("properties.%s", k)

			feature, err = sjson.SetBytes(feature, path, str_v)

			if err != nil {
				return nil, err
			}
		}
	}

	return feature, nil
}

func isISO(feature []byte, code string) bool {

	match := false

	paths := []string{
		"properties.addr:country",
		"properties.wof:country",
		"properties.iso:country",
	}

	for _, p := range paths {

		v := gjson.GetBytes(feature, p)

		if !v.Exists() {
			continue
		}

		if strings.ToUpper(v.String()) == code {
			match = true
			break
		}
	}

	return match
}
