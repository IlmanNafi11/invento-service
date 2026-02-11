package helper

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTusMetadata_EmptyHeader_ReturnsEmptyMap(t *testing.T) {
	parsed := ParseTusMetadata("")
	assert.Empty(t, parsed)
}

func TestParseTusMetadata_SinglePair_ParsesValue(t *testing.T) {
	value := base64.StdEncoding.EncodeToString([]byte("project.zip"))
	parsed := ParseTusMetadata("filename " + value)

	assert.Equal(t, "project.zip", parsed["filename"])
}

func TestParseTusMetadata_MultiplePairs_ParsesAll(t *testing.T) {
	name := base64.StdEncoding.EncodeToString([]byte("doc.zip"))
	typev := base64.StdEncoding.EncodeToString([]byte("application/zip"))
	user := base64.StdEncoding.EncodeToString([]byte("123"))

	parsed := ParseTusMetadata("filename " + name + ",content_type " + typev + ",user_id " + user)

	assert.Equal(t, "doc.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content_type"])
	assert.Equal(t, "123", parsed["user_id"])
}

func TestParseTusMetadata_KeyWithoutValue_SetsEmptyValue(t *testing.T) {
	parsed := ParseTusMetadata("filename")
	assert.Equal(t, "", parsed["filename"])
}

func TestParseTusMetadata_InvalidBase64_SkipsGracefullyWithEmptyValue(t *testing.T) {
	parsed := ParseTusMetadata("filename !!!")
	assert.Equal(t, "", parsed["filename"])
}

func TestParseTusMetadata_ExtraWhitespace_StillParses(t *testing.T) {
	value := base64.StdEncoding.EncodeToString([]byte(" spaced name.zip "))
	parsed := ParseTusMetadata("   filename   " + value + "   ")

	assert.Equal(t, " spaced name.zip ", parsed["filename"])
}

func TestParseTusMetadata_CommaSeparatedPairs_HandlesMixedEntries(t *testing.T) {
	a := base64.StdEncoding.EncodeToString([]byte("A"))
	b := base64.StdEncoding.EncodeToString([]byte("B"))

	parsed := ParseTusMetadata("k1 " + a + ", invalidpair, k2 " + b + ",novalue")

	assert.Equal(t, "A", parsed["k1"])
	assert.Equal(t, "B", parsed["k2"])
	assert.Equal(t, "", parsed["invalidpair"])
	assert.Equal(t, "", parsed["novalue"])
}
