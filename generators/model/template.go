package model

const filters = `Filter{ {{.Column}}, {{if not .NoPointer}}*{{end}}{{.ShortVarName}}s.{{.Value}}, {{.SearchType}}, {{.Exclude}}}.Apply(query)`

const modelTemplate = `//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package {{.Package}}{{if .HasImports}}

import ({{range .Imports}}
    "{{.}}"{{end}}
){{end}}

var Columns = struct { {{range .Entities}}
	{{.Name}} struct{ 
		{{range $i, $e := .Columns}}{{if $i}}, {{end}}{{.Name}}{{end}} string{{if .HasRelations}}

		{{range $i, $e := .Relations}}{{if $i}}, {{end}}{{.Name}}{{end}} string{{end}}
	}{{end}}
}{ {{range .Entities}}
	{{.Name}}: struct { 
		{{range $i, $e := .Columns}}{{if $i}}, {{end}}{{.Name}}{{end}} string{{if .HasRelations}}

		{{range $i, $e := .Relations}}{{if $i}}, {{end}}{{.Name}}{{end}} string{{end}}
	}{ {{range .Columns}}
		{{.Name}}: "{{.DBName}}",{{end}}{{if .HasRelations}}
		{{range .Relations}}
		{{.Name}}: "{{.Name}}",{{end}}{{end}}
	},{{end}}
}

var Tables = struct { {{range .Entities}}
	{{.Name}} struct {
		Name, Alias string
	}{{end}}
}{ {{range .Entities}}
	{{.Name}}: struct {
		Name, Alias string
	}{ 
		Name: "{{.Table}}",
		Alias: "{{.Alias}}",
	},{{end}}
}
{{range $model := .Entities}}
type {{.Name}} struct {
	tableName struct{} {{.Tag}}
	{{range .Columns}}
	{{.Name}} {{.GoType}} {{.Tag}} {{.Comment}}{{end}}{{if .HasRelations}}
	{{range .Relations}}
	{{.Name}} *{{.Type}} {{.Tag}} {{.Comment}}{{end}}{{end}}
}
{{end}}
`

const searchTemplate = `//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package {{.Package}}

import ({{if .HasImports}}{{range .Imports}}
	"{{.}}"{{end}}
	{{end}}
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

const condition =  "?.? = ?"

// base filters
type applier func(query *orm.Query) (*orm.Query, error)

type search struct {
	appliers[] applier
}

func (s *search) apply(query *orm.Query) {
	for _, applier := range s.appliers {
		query.Apply(applier)
	}
}

func (s *search) where(query *orm.Query, table, field string, value interface{}) {
	query.Where(condition, pg.F(table), pg.F(field), value)
}

func (s *search) WithApply(a applier) {
	if s.appliers == nil {
		s.appliers = []applier{}
	}
	s.appliers = append(s.appliers, a)
}

func (s *search) With(condition string, params ...interface{}) {
	s.WithApply(func(query *orm.Query) (*orm.Query, error) {
		return query.Where(condition, params...), nil
	})
}

// Searcher is interface for every generated filter
type Searcher interface {
	Apply(query *orm.Query) *orm.Query
	Q() applier

	With(condition string, params ...interface{})
	WithApply(a applier)
}

{{range $model := .Entities}}
type {{.Name}}Search struct {
	search 

	{{range .Columns}}
	{{.Name}} {{.GoType}}{{end}}
}

func ({{$model.ShortVarName}}s *{{.Name}}Search) Apply(query *orm.Query) *orm.Query { {{range .Columns}}
	{{if .IsArray}} if len({{$model.ShortVarName}}s.{{.Name}}) > 0 { {{else}}if {{$model.ShortVarName}}s.{{.Name}} != nil { {{end}}{{if .UseCustomRender}}
		{{.CustomRender}}{{else}} 
		{{$model.ShortVarName}}s.where(query, Tables.{{$model.Name}}.Alias, Columns.{{$model.Name}}.{{.Name}}, {{$model.ShortVarName}}s.{{.Name}}){{end}}
	}{{end}}

	{{$model.ShortVarName}}s.apply(query)
	
	return query
}

func ({{$model.ShortVarName}}s *{{.Name}}Search) Q() applier {
	return func(query *orm.Query) (*orm.Query, error) {
		return {{$model.ShortVarName}}s.Apply(query), nil
	}
}
{{end}}
`

const validateTemplate = `//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package {{.Package}}{{if .HasImports}}

import ({{range .Imports}}
    "{{.}}"{{end}}
){{end}}

const (
	ErrEmptyValue = "empty"
	ErrMaxLength  = "len"
	ErrWrongValue = "value"
)

{{range $model := .Entities}}
func ({{$model.ShortVarName}} {{.Name}}) Validate() (errors map[string]string, valid bool) {
	errors = map[string]string{}

	{{range .Columns}}
	{{if eq .Check "nil" }}
	if {{$model.ShortVarName}}.{{.Name}} == nil {
		errors[Columns.{{$model.Name}}.{{.Name}}] = ErrEmptyValue
	}	
	{{else if eq .Check "zero"}}
	if {{$model.ShortVarName}}.{{.Name}} == 0 {
		errors[Columns.{{$model.Name}}.{{.Name}}] = ErrEmptyValue
	}
	{{else if eq .Check "pzero"}}
	if {{$model.ShortVarName}}.{{.Name}} != nil && *{{$model.ShortVarName}}.{{.Name}} == 0 {
		errors[Columns.{{$model.Name}}.{{.Name}}] = ErrEmptyValue
	}
	{{else if eq .Check "len"}}
	if utf8.RuneCountInString({{$model.ShortVarName}}.{{.Name}}) > {{.Max}} {
		errors[Columns.{{$model.Name}}.{{.Name}}] = ErrMaxLength
	}
	{{else if eq .Check "plen"}}
	if {{$model.ShortVarName}}.{{.Name}} != nil && utf8.RuneCountInString(*{{$model.ShortVarName}}.{{.Name}}) > {{.Max}} {
		errors[Columns.{{$model.Name}}.{{.Name}}] = ErrMaxLength
	}
	{{end}}
	{{end}}

	return errors, len(errors) == 0
}
{{end}}
`
