// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package dotenv

import "testing"

const TEST_EXPORTED = `export OPTION_A=2
export OPTION_B='\n' # foo
#export OPTION_C=3
export OPTION_D=
export OPTION_E="foo"
`

func TestDotEnvExported(t *testing.T) {
	env := MustParseEnv(TEST_EXPORTED)

	if env["OPTION_A"] != "2" {
		t.Error("OPTION_A")
	}
	if env["OPTION_B"] != "\\n" {
		t.Error("OPTION_B")
	}
	if env["OPTION_C"] != "" {
		t.Error("OPTION_C", env["OPTION_C"])
	}
	if v, ok := env["OPTION_D"]; !(v == "" && ok) {
		t.Error("OPTION_D")
	}
	if env["OPTION_E"] != "foo" {
		t.Error("OPTION_E")
	}
}

const TEST_PLAIN = `OPTION_A=1
OPTION_B=2
OPTION_C= 3
OPTION_D =4
OPTION_E = 5
OPTION_F=
OPTION_G =
`

func TestDotEnvPlain(t *testing.T) {
	env := MustParseEnv(TEST_PLAIN)

	if env["OPTION_A"] != "1" {
		t.Error("OPTION_A")
	}
	if env["OPTION_B"] != "2" {
		t.Error("OPTION_B")
	}
	if env["OPTION_C"] != "3" {
		t.Error("OPTION_C")
	}
	if env["OPTION_D"] != "4" {
		t.Error("OPTION_D")
	}
	if env["OPTION_E"] != "5" {
		t.Error("OPTION_E")
	}
	if v, ok := env["OPTION_F"]; !(v == "" && ok) {
		t.Error("OPTION_F")
	}
	if v, ok := env["OPTION_G"]; !(v == "" && ok) {
		t.Error("OPTION_G")
	}
}

const TEST_QUOTED = `OPTION_A='1'
OPTION_B='2'
OPTION_C=''
OPTION_D='\n'
OPTION_E="1"
OPTION_F="2"
OPTION_G=""
OPTION_H="\n"
#OPTION_I="3"
`

func TestDotEnvQuoted(t *testing.T) {
	env := MustParseEnv(TEST_QUOTED)

	if env["OPTION_A"] != "1" {
		t.Error("OPTION_A")
	}
	if env["OPTION_B"] != "2" {
		t.Error("OPTION_B")
	}
	if env["OPTION_C"] != "" {
		t.Error("OPTION_C")
	}
	if env["OPTION_D"] != "\\n" {
		t.Error("OPTION_D")
	}
	if env["OPTION_E"] != "1" {
		t.Error("OPTION_E")
	}
	if env["OPTION_F"] != "2" {
		t.Error("OPTION_F")
	}
	if env["OPTION_G"] != "" {
		t.Error("OPTION_G")
	}
	if env["OPTION_H"] != "\n" {
		t.Error("OPTION_H")
	}
	if env["OPTION_I"] != "" {
		t.Error("OPTION_I")
	}
}

const TEST_YAML = `OPTION_A: 1
OPTION_B: '2'
OPTION_C: ''
OPTION_D: '\n'
#OPTION_E: '333'
OPTION_F: 
`

func TestDotEnvYAML(t *testing.T) {
	env := MustParseEnv(TEST_YAML)

	if env["OPTION_A"] != "1" {
		t.Error("OPTION_A")
	}
	if env["OPTION_B"] != "2" {
		t.Error("OPTION_B")
	}
	if env["OPTION_C"] != "" {
		t.Error("OPTION_C")
	}
	if env["OPTION_D"] != "\\n" {
		t.Error("OPTION_D")
	}
	if env["OPTION_E"] != "" {
		t.Error("OPTION_E")
	}
	if v, ok := env["OPTION_F"]; !(v == "" && ok) {
		t.Error("OPTION_F")
	}
}
