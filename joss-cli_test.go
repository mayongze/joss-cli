package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	hasRunArg, hasRunExample := false, false

	//检测运行的结果
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.run=") {
			exp := strings.Split(arg, "=")[1]
			match, err := regexp.MatchString(exp, "Example")
			hasRunExample = (err == nil && match)
			hasRunArg = true
			break
		}
	}

	if !hasRunArg {
		// 在没有-test.run的情况下强制运行Test*的测试,以免发生错误
		os.Args = append(os.Args, "-test.run=Test")
	}

	if !hasRunExample {

	}

	v := m.Run()
	os.Exit(v)
}
