package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractThoughtSignature(t *testing.T) {
	t.Run("extracts snake_case from part", func(t *testing.T) {
		part := map[string]interface{}{
			"thought_signature": "sig_part_snake",
		}
		sig := extractThoughtSignature(nil, part, nil)
		assert.Equal(t, "sig_part_snake", sig)
	})

	t.Run("extracts camelCase from part", func(t *testing.T) {
		part := map[string]interface{}{
			"thoughtSignature": "sig_part_camel",
		}
		sig := extractThoughtSignature(nil, part, nil)
		assert.Equal(t, "sig_part_camel", sig)
	})

	t.Run("extracts from functionCall map", func(t *testing.T) {
		fc := map[string]interface{}{
			"name":              "test_tool",
			"thought_signature": "sig_fc_snake",
		}
		sig := extractThoughtSignature(nil, nil, fc)
		assert.Equal(t, "sig_fc_snake", sig)
	})

	t.Run("extracts from candidate map", func(t *testing.T) {
		cand := map[string]interface{}{
			"thought_signature": "sig_cand",
		}
		sig := extractThoughtSignature(cand, nil, nil)
		assert.Equal(t, "sig_cand", sig)
	})

	t.Run("extracts from candidate content map", func(t *testing.T) {
		cand := map[string]interface{}{
			"content": map[string]interface{}{
				"thoughtSignature": "sig_cand_content",
			},
		}
		sig := extractThoughtSignature(cand, nil, nil)
		assert.Equal(t, "sig_cand_content", sig)
	})

	t.Run("returns empty string when no signature exists", func(t *testing.T) {
		sig := extractThoughtSignature(nil, map[string]interface{}{"text": "hello"}, nil)
		assert.Equal(t, "", sig)
	})
}
