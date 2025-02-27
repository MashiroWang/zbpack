package php

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

// DefaultPHPVersion represents the default PHP version.
const DefaultPHPVersion = "8.1"

// GetPHPVersion gets the php version of the project.
func GetPHPVersion(source afero.Fs) string {
	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return DefaultPHPVersion
	}

	versionRange, ok := composerJSON.GetRequire("php")
	if !ok || versionRange == "" {
		return DefaultPHPVersion
	}

	isVersion, _ := regexp.MatchString(`^\d+(\.\d+){0,2}$`, versionRange)
	if isVersion {
		return versionRange
	}
	ranges := strings.Split(versionRange, " ")
	for _, r := range ranges {
		if strings.HasPrefix(r, ">=") {
			minVerString := strings.TrimPrefix(r, ">=")
			return minVerString
		} else if strings.HasPrefix(r, ">") {
			minVerString := strings.TrimPrefix(r, ">")
			value, err := strconv.ParseFloat(minVerString, 64)
			if err != nil {
				log.Println("parse php version error", err)
				continue
			}
			value += 0.1
			minVerString = fmt.Sprintf("%f", value)
			return minVerString
		} else if strings.HasPrefix(r, "<=") {
			maxVerString := strings.TrimPrefix(r, "<=")
			return maxVerString

		} else if strings.HasPrefix(r, "<") {
			maxVerString := strings.TrimPrefix(r, "<=")
			value, err := strconv.ParseFloat(maxVerString, 64)
			if err != nil {
				log.Println("parse php version error", err)
				continue
			}
			value -= 0.1

			maxVerString = fmt.Sprintf("%f", value)
			return maxVerString
		}
	}

	return DefaultPHPVersion
}

// DetermineProjectFramework determines the framework of the project.
func DetermineProjectFramework(source afero.Fs) types.PHPFramework {
	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return types.PHPFrameworkNone
	}

	if _, isLaravel := composerJSON.GetRequire("laravel/framework"); isLaravel {
		return types.PHPFrameworkLaravel
	}

	if _, isThinkPHP := composerJSON.GetRequire("topthink/framework"); isThinkPHP {
		return types.PHPFrameworkThinkphp
	}

	if _, isCodeIgniter := composerJSON.GetRequire("codeigniter4/framework"); isCodeIgniter {
		return types.PHPFrameworkCodeigniter
	}

	return types.PHPFrameworkNone
}

var depMap = map[string][]string{
	"ext-openssl": {"libssl-dev"},
	"ext-zip":     {"libzip-dev"},
	"ext-curl":    {"libcurl4-openssl-dev", "libssl-dev"},
	"ext-gd":      {"libpng-dev"},
	"ext-gmp":     {"libgmp-dev"},
}

var baseDep = []string{"libicu-dev", "jq", "pkg-config", "unzip", "git"}

// DetermineAptDependencies determines the required apt dependencies of the project.
//
// We install Nginx server unless server is "swoole".
func DetermineAptDependencies(source afero.Fs, server string) []string {
	// deep copy the base dependencies
	dependencies := append([]string{}, baseDep...)

	// If Octane Server is not "swoole", we should install Nginx.
	//
	// TODO: support RoadRunner
	if server != "swoole" {
		dependencies = append(dependencies, "nginx")
	}

	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return dependencies
	}

	if composerJSON.Require == nil {
		return dependencies
	}

	// loop through the composer.json dependencies and
	// check if any dependency need some additional apt dependencies
	for dep := range *composerJSON.Require {
		if val, ok := depMap[dep]; ok {
			dependencies = append(dependencies, val...)
		}
	}

	return dependencies
}

// DetermineApplication determines what application the project is using.
// Therefore, we can apply some custom fixes such as the nginx configuration.
func DetermineApplication(source afero.Fs) (types.PHPApplication, types.PHPProperty) {
	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return types.PHPApplicationDefault, types.PHPPropertyNone
	}

	if composerJSON.Name == "lizhipay/acg-faka" {
		return types.PHPApplicationAcgFaka, types.PHPPropertyComposer
	}

	return types.PHPApplicationDefault, types.PHPPropertyComposer
}
