package lieu

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	_ "log"
	"strings"
)

func EnsureProperties(feature []byte) error {

	possible_names := []string{
		"properties.name",
		"properties.wof:name",
	}

	possible_streets := []string{
		"properties.addr:street",
	}

	possible_housenumbers := []string{
		"properties.addr:house_number",
	}

	if !hasProperty(feature, possible_names) {
		return errors.New("Missing name property")
	}

	if !hasProperty(feature, possible_streets) {
		return errors.New("Missing street property")
	}

	if !hasProperty(feature, possible_housenumbers) {
		return errors.New("Missing house_number property")
	}

	return nil
}

func hasProperty(feature []byte, possible []string) bool {

	has_property := false

	for _, path := range possible {

		v := gjson.GetBytes(feature, path)

		if v.Exists() {
			has_property = true
			break
		}
	}

	return has_property
}

func Prepare(in io.Reader, out io.Writer) error {

	reader := bufio.NewReader(in)

	sep := byte('\n') // note that double-quotes will _freak_ Go out...

	for {

		b, err := reader.ReadBytes(sep)

		if err != nil {

			if err == io.EOF {
				break
			}

			return err
		}

		b2, err := PrepareFeature(b)

		if err != nil {
			return err
		}

		out.Write(b2)
	}

	return nil
}

func PrepareFeature(feature []byte) ([]byte, error) {

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
