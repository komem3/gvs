package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type specifyVersion interface {
	version() string
}

type (
	same     string
	asterisk struct{}
)

func (i same) version() string     { return string(i) }
func (i asterisk) version() string { return "" }

type version struct {
	major specifyVersion
	minor specifyVersion
	patch specifyVersion
}

func compareVersion(str string, version specifyVersion) bool {
	switch v := version.(type) {
	case same:
		return version.version() == str
	case asterisk:
		return true
	default:
		panic(fmt.Sprintf("unknown type %T", v))
	}
}

func parseVersionString(str string) (*version, error) {
	str = strings.Trim(str, "gov/")
	splits := strings.Split(str, ".")
	if len(splits) == 0 {
		return nil, fmt.Errorf("%s is not support format", str)
	}
	v := &version{
		major: asterisk{},
		minor: asterisk{},
		patch: asterisk{},
	}
	for i, str := range splits {
		switch i {
		case 0:
			v.major = same(str)
		case 1:
			v.minor = same(str)
		case 2:
			v.patch = same(str)
		}
	}

	return v, nil
}

func compareVersionString(numberStrs []string, v *version) bool {
	for i, str := range numberStrs {
		switch i {
		case 0:
			if !compareVersion(str, v.major) {
				return false
			}
		case 1:
			if !compareVersion(str, v.minor) {
				return false
			}
		case 2:
			if !compareVersion(str, v.patch) {
				return false
			}
		}
	}
	return true
}

func mustParse(numberStr string) int {
	num, err := strconv.ParseInt(numberStr, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(num)
}

func calcPriority(splitVersion []string) int {
	var priority int
	for i, verStr := range splitVersion {
		priority += mustParse(verStr) * int(math.Pow(10, float64(3-i)))
	}
	return priority
}
