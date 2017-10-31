package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	localTZ *time.Location
	aliases map[string]string
}

func mustLoadAliases(in io.Reader) map[string]string {
	bb, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatalf("couldn't load config: %s", err)
	}
	var user = make(map[string]string)
	err = yaml.Unmarshal(bb, &user)
	if err != nil {
		log.Fatalf("couldn't unmarshal config: %s", err)
	}
	for k, v := range defaultAlias {
		_, exists := user[k]
		if !exists {
			user[k] = v
		}
	}

	return user
}

// https://stackoverflow.com/a/41786440
func userHomeDir() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	return os.Getenv(env)
}

func tzFromShell() (*time.Location, error) {
	bb, err := exec.Command("date", "+%z").Output()
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(string(bb))

	dateRX := regexp.MustCompile(`^([+\-])(\d\d)(\d\d)\s?$`)
	match := dateRX.FindSubmatch(bb)
	if len(match) == 0 {
		return nil, errors.New("date +%z output didn't match expected regex")
	}

	signString := string(match[1][0])
	sign := 1
	if signString == "-" {
		sign = -1
	}

	if len(match[2]) != 2 {
		return nil, errors.New(`date +%z output did not have two hour bytes`)
	}
	if len(match[3]) != 2 {
		return nil, errors.New(`date +%z output did not have two minute bytes`)
	}

	hh := string(match[2])
	mm := string(match[3])

	h, err := strconv.Atoi(hh)
	if err != nil {
		return nil, errors.New("second and third bytes of date +%z were not integers")
	}
	m, err := strconv.Atoi(mm)
	if err != nil {
		return nil, errors.New("fourth and fifth bytes of date +%z were not integers")
	}

	s := (60*m + 3600*h) * sign
	return time.FixedZone(name, s), nil
}
