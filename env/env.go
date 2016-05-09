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

const PlexCredApiSeparator = ">"
const PlexUserPassSeparator = "^"

func GetCoordinates(cfEnvs string) ([]Coord, error) {
	var coords []Coord
	triples := GetTriples(cfEnvs)
	for _, triple := range triples {
		coord, err := GetCoordinate(triple)
		if err != nil {
			return nil, err
		}
		coords = append(coords, coord)
	}
	return coords, nil
}

func GetTriples(cfEnvs string) []string {
	return strings.Split(cfEnvs, ";")
}

func GetCoordinate(triple string) (coord Coord, err error) {
	if strings.Count(triple, PlexCredApiSeparator) != 1 || strings.Count(triple, PlexUserPassSeparator) != 1 {
		return coord, errors.New(triple + " is invalid")
	}
	credsAndApi := strings.Split(triple, PlexCredApiSeparator)
	creds := strings.Split(credsAndApi[0], PlexUserPassSeparator)
	username := creds[0]
	password := creds[1]
	api := credsAndApi[1]

	return Coord{Username: username, Password: password, Api: api}, err
}
