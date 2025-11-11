package config

import (
	"encoding/json"
	"errors"
	"strings"

	schemaData "github.com/viktorprogger/universal-linux-installer/schema"
	"github.com/xeipuuv/gojsonschema"
)

var schemaJSON = schemaData.Bytes

func ValidateAgainstSchema(cfg Config) error {
	if len(schemaJSON) == 0 {
		return errors.New("schema not embedded")
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	docLoader := gojsonschema.NewBytesLoader(b)
	res, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return err
	}
	if res.Valid() {
		return nil
	}
	var msgs []string
	for _, e := range res.Errors() {
		msgs = append(msgs, e.String())
	}
	return errors.New("schema validation failed: " + strings.Join(msgs, "; "))
}
