package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsValidTagKey(t *testing.T) {
	res := IsValidTagKey("test-test")
	assert.False(t, res)
	res = IsValidTagKey("_4324hdj_fas213")
	assert.True(t, res)
	res = IsValidTagKey("4324hdj_fas213")
	assert.False(t, res)
	res = IsValidTagKey("是3213dsa")
	assert.False(t, res)
	res = IsValidTagKey("cls-xxxx-eks-eip-tag-pod-uid")
	assert.False(t, res)
}
