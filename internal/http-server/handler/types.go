package handler

import "github.com/Onnywrite/tinkoff-prod/pkg/ero"

// It exists because I'm lazy to wrap wrap wrap and wrap ero.Error again.
// Of course it's better to not have 2 same structs expecially
// when ero.Error with the same purpose exists
type errorMessage string

func (e errorMessage) Blob() []byte {
	return []byte(`{"Service":"` + ero.CurrentService + `","ErrorMessage":"` + string(e) + `"}`)
}
