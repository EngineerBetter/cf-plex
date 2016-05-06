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
	if strings.Count(triple, "@") != 1 || strings.Count(triple, ":") != 1 {
		return coord, errors.New(triple + " is invalid")
	}
	credsAndApi := strings.Split(triple, "@")
	creds := strings.Split(credsAndApi[0], ":")
	username := creds[0]
	password := creds[1]
	api := credsAndApi[1]

	return Coord{Username: username, Password: password, Api: api}, err
}
