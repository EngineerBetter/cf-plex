package env

import (
	"errors"
	"strings"
)

type Coord struct {
	Username string
	Password string
	Api      string
}

const PlexTripleSeparator = ";"
const PlexCredApiSeparator = ">"
const PlexUserPassSeparator = "^"

func GetCoordinates(cfEnvs, tripleSeparator, credApiSeparator, userPassSeparator string) ([]Coord, error) {
	var coords []Coord
	triples := GetTriples(cfEnvs, tripleSeparator)
	for _, triple := range triples {
		coord, err := GetCoordinate(triple, credApiSeparator, userPassSeparator)
		if err != nil {
			return nil, err
		}
		coords = append(coords, coord)
	}
	return coords, nil
}

func GetTriples(cfEnvs, tripleSeparator string) []string {
	return strings.Split(cfEnvs, tripleSeparator)
}

func GetCoordinate(triple, credApiSeparator, userPassSeparator string) (coord Coord, err error) {
	if strings.Count(triple, credApiSeparator) != 1 ||
		strings.Count(triple, userPassSeparator) != 1 {
		return coord, errors.New(triple + " is invalid")
	}
	credsAndApi := strings.Split(triple, credApiSeparator)
	creds := strings.Split(credsAndApi[0], userPassSeparator)
	username := creds[0]
	password := creds[1]
	api := credsAndApi[1]

	return Coord{Username: username, Password: password, Api: api}, err
}
