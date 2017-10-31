package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	zoneRX              = regexp.MustCompile(`^[a-zA-Z+\-]`)
	fragRX              = regexp.MustCompile("^[0-9]")
	errParsingTimestamp = errors.New("found numbers that did not match a known timestamp format")
	errNoLocalTimezone  = errors.New("parse called without configuring a local timezone")
)

type outputZone struct {
	alias string
	loc   *time.Location
}

func parse(cfg config, args []string) (outputZones []outputZone, t time.Time, matchedTimestampFormat string, err error) {
	var (
		inputZone *time.Location
		tmpZone   *time.Location
	)
	t = nullTime
	inputZone = cfg.localTZ
	if inputZone == nil {
		return outputZones, t, "", errNoLocalTimezone
	}
	if len(args) == 0 {
		return outputZones, t, "", errNoArgs
	}

	outputZoneStrings, timeFrags, inputZoneString := parseToStrings(args)

	if inputZoneString != "" {
		inputZone, err = parseZone(cfg, inputZoneString)
		if err != nil {
			return outputZones, t, "", err
		}
	}

	for _, tzString := range outputZoneStrings {
		tmpZone, err = parseZone(cfg, tzString)
		if err != nil {
			return outputZones, t, "", err
		}
		outputZones = append(outputZones, outputZone{
			alias: tzString,
			loc:   tmpZone,
		})
	}

	if len(timeFrags) > 0 {
		timeString := strings.Join(timeFrags, " ")
		for _, matchedTimestampFormat = range tsFormats {
			t, err = time.ParseInLocation(matchedTimestampFormat, timeString, inputZone)
			if err == nil {
				break
			}
		}
	}

	return outputZones, t, matchedTimestampFormat, err
}

func parseToStrings(args []string) (outputZoneStrings []string, timeFrags []string, inputZoneString string) {
	for _, arg := range args {
		if arg == "at" || arg == "in" || arg == "from" || arg == "to" {
			continue
		}
		isFrag := fragRX.MatchString(arg)
		if isFrag {
			timeFrags = append(timeFrags, arg)
			continue
		}
		isZone := zoneRX.MatchString(arg)
		if isZone {
			if len(timeFrags) == 0 {
				outputZoneStrings = append(outputZoneStrings, arg)
				continue
			}
			inputZoneString = arg
			alias, ok := defaultAlias[inputZoneString]
			if ok {
				inputZoneString = alias
			}
		}
	}
	return outputZoneStrings, timeFrags, inputZoneString
}

func parseZone(cfg config, s string) (*time.Location, error) {
	tmp, err := time.LoadLocation(s)
	if err != nil {
		// retry with alias:
		alias, ok := cfg.aliases[s]
		if !ok {
			return nil, fmt.Errorf("could not parse %s as timezone", s)
		}
		tmp, err = time.LoadLocation(alias)
	}
	return tmp, err
}
