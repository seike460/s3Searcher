package s3ry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewS3ry(t *testing.T) {
	s := NewS3ry()
	operations := s.ListOperation()
	buckets := s.ListBuckets()
	uploadItem := s.ListUpload("seike460")
	s.SaveObjectList("seike460")
	items := s.ListObjectsPages("seike460")
	selectObject := s.ListObjects("seike460")
	assert.NotNil(t, operations)
	assert.NotNil(t, buckets)
	assert.NotNil(t, uploadItem)
	assert.NotNil(t, items)
	assert.NotNil(t, selectObject)
	CheckLocalExists("main.go")
	//s.ListObjectsPages("seike460")
}
