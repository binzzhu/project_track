//go:build tools
// +build tools

package main

// 这个文件用于强制 go mod vendor 包含间接依赖
// 这些包被 modernc.org/libc 和 github.com/leodido/go-urn 间接使用
import (
	_ "golang.org/x/exp/constraints"
	_ "github.com/leodido/go-urn/scim/schema"
)
