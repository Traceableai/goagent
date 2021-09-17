package main

func TestCheckMinVersion(t *testing.T) {
	assert.True(t, checkMinVersion("20", "20.08"))
	assert.True(t, checkMinVersion("10", "10"))
	assert.False(t, checkMinVersion("3.9", "3.8.6"))
	assert.True(t, checkMinVersion("3.9", "3.12"))
}
