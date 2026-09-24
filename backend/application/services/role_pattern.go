package services

import "regexp"

var roleRe = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,31}$`)

func mustCompileRole() *regexp.Regexp { return roleRe }
