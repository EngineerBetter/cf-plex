package env

import (
	"strings"
)

type Coord struct {
	Username string
	Password string
	Api      string
}

func GetCoordinates(cfEnvs string) []Coord {
	var coords []Coord
	triples := GetTriples(cfEnvs)
	for _, triple := range triples {
		coords = append(coords, GetCoordinate(triple))
	}
	return coords
}

func GetTriples(cfEnvs string) []string {
	return strings.Split(cfEnvs, ";")
}

func GetCoordinate(triple string) Coord {
	credsAndApi := strings.Split(triple, "@")
	creds := strings.Split(credsAndApi[0], ":")
	username := creds[0]
	password := creds[1]
	api := credsAndApi[1]

	return Coord{Username: username, Password: password, Api: api}
}
