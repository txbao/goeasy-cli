package templates

import "embed"

// AllTemplates 内嵌全部脚手架模板（项目 + 模块 + 变体）。
//
//go:embed project monolith auth system payment module crud proto repository event aggregate
var AllTemplates embed.FS
