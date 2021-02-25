// Copyright © 2020-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package dotenv

import (
	"io/ioutil"
	"os"
	"strconv"
)

// 环境变量
type Env map[string]string

var _env = make(Env)

func (e Env) Get(key string) string {
	val, found := e[key]
	if !found {
		val = os.Getenv(key)
	}
	return val
}

func (e Env) GetBool(key string) bool {
	v := e.Get(key)
	b, _ := strconv.ParseBool(v)
	return b
}

func (e Env) GetInt(key string) int {
	v := e.Get(key)
	n, _ := strconv.Atoi(v)
	return n
}

func (e Env) GetFloat(key string) float64 {
	v := e.Get(key)
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

// 加载.env变量配置
func Load(filename string, overload bool) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	envMap, err := ParseEnv(string(data))
	if err != nil {
		return err
	}
	Add(envMap, overload)
	return nil
}

func Add(envMap map[string]string, overload bool) {
	if overload {
		for k, v := range envMap {
			_env[k] = v
		}
	} else {
		for k, v := range envMap {
			if _, found := _env[k]; !found {
				_env[k] = v
			}
		}
	}
}

func Get(key string) string {
	return _env.Get(key)
}

func GetBool(key string) bool {
	return _env.GetBool(key)
}

func GetInt(key string) int {
	return _env.GetInt(key)
}

func GetFloat(key string) float64 {
	return _env.GetFloat(key)
}
