// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package dotenv

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// LINE is the regexp matching a single line
const LINE = `
\A
\s*
(?:|#.*|          # comment line
(?:export\s+)?    # optional export
([\w\.]+)         # key
(?:\s*=\s*|:\s+?) # separator
(                 # optional value begin
  '(?:\'|[^'])*'  #   single quoted value
  |               #   or
  "(?:\"|[^"])*"  #   double quoted value
  |               #   or
  [^#\n]+         #   unquoted value
)?                # value end
\s*
(?:\#.*)?         # optional comment
)
\z
`

var (
	escapeRegex = regexp.MustCompile("\\\\([^$])")
	linesRegex  = regexp.MustCompile("[\\r\\n]+")
	lineRegex   = regexp.MustCompile(
		regexp.MustCompile("\\s+").ReplaceAllLiteralString(
			regexp.MustCompile("\\s+# .*").ReplaceAllLiteralString(LINE, ""), ""))
)

// ParseEnv reads a string in the .env format and returns a map of the extracted key=values.
// see https://github.com/bkeepers/dotenv
func ParseEnv(data string) (map[string]string, error) {
	var dotenv = make(map[string]string)
	for _, line := range linesRegex.Split(data, -1) {
		if !lineRegex.MatchString(line) {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		match := lineRegex.FindStringSubmatch(line)
		// commented or empty line
		if len(match) == 0 {
			continue
		}
		if len(match[1]) == 0 {
			continue
		}
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		err := parseEnvValue(key, value, dotenv)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s, %s: %s", key, value, err)
		}
	}
	return dotenv, nil
}

// MustParseEnv works the same as Parse but panics on error
func MustParseEnv(data string) map[string]string {
	env, err := ParseEnv(data)
	if err != nil {
		panic(err)
	}
	return env
}

func parseEnvValue(key string, value string, dotenv map[string]string) error {
	if len(value) <= 1 {
		dotenv[key] = value
		return nil
	}
	var singleQuoted bool
	if value[0:1] == "'" && value[len(value)-1:] == "'" {
		// single-quoted string, do not expand
		singleQuoted = true
		value = value[1 : len(value)-1]
	} else if value[0:1] == `"` && value[len(value)-1:] == `"` {
		value = value[1 : len(value)-1]
		value = expandNewLines(value)
		value = unescapeCharacters(value)
	}
	if !singleQuoted {
		value = expandEnv(value, dotenv)
	}
	dotenv[key] = value
	return nil
}

func unescapeCharacters(value string) string {
	return escapeRegex.ReplaceAllString(value, "$1")
}

func expandEnv(value string, dotenv map[string]string) string {
	return os.Expand(value, func(value string) string {
		expanded, found := dotenv[value]
		if found {
			return expanded
		} else {
			return os.Getenv(value)
		}
	})
}

func expandNewLines(value string) string {
	value = strings.Replace(value, "\\n", "\n", -1)
	value = strings.Replace(value, "\\r", "\r", -1)
	return value
}
