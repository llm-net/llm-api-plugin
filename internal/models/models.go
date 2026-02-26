package models

import "encoding/json"

// Param describes a single parameter for a model.
type Param struct {
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Options     []string `json:"options,omitempty"`
	Default     string   `json:"default,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// Model describes one model's capabilities and parameters.
type Model struct {
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Capabilities []string         `json:"capabilities"`
	Params       map[string]Param `json:"params,omitempty"`
}

// Registry holds a list of models for a CLI tool.
type Registry struct {
	Tool   string  `json:"tool"`
	Models []Model `json:"models"`
}

// JSON returns the registry as indented JSON bytes.
func (r *Registry) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// FindModel returns the model with the given name, or nil.
func (r *Registry) FindModel(name string) *Model {
	for i := range r.Models {
		if r.Models[i].Name == name {
			return &r.Models[i]
		}
	}
	return nil
}
