package util

import (
	"testing"
	"url-shortener/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestGenerateShortCode(t *testing.T) {
	t.Run("Generates Non-Empty Code", func(t *testing.T) {
		code := GenerateShortCode()
		assert.NotEmpty(t, code)
	})

	t.Run("Generates 8 Character Code", func(t *testing.T) {
		code := GenerateShortCode()
		assert.Equal(t, 8, len(code))
	})

	t.Run("Generates Unique Codes", func(t *testing.T) {
		codes := make(map[string]bool)

		// Generate 100 codes and check for uniqueness
		for i := 0; i < 100; i++ {
			code := GenerateShortCode()
			assert.False(t, codes[code], "Generated duplicate code: "+code)
			codes[code] = true
		}
	})

	t.Run("Code Contains URL-Safe Characters", func(t *testing.T) {
		code := GenerateShortCode()

		// URL-safe base64 characters: A-Z, a-z, 0-9, -, _
		for _, char := range code {
			isValid := (char >= 'A' && char <= 'Z') ||
				(char >= 'a' && char <= 'z') ||
				(char >= '0' && char <= '9') ||
				char == '-' || char == '_'
			assert.True(t, isValid, "Invalid character in code: %c", char)
		}
	})
}

func TestParseInt(t *testing.T) {
	t.Run("Parse Valid Integer", func(t *testing.T) {
		result := ParseInt("123")
		assert.Equal(t, 123, result)
	})

	t.Run("Parse Zero", func(t *testing.T) {
		result := ParseInt("0")
		assert.Equal(t, 0, result)
	})

	t.Run("Parse Negative Integer", func(t *testing.T) {
		result := ParseInt("-456")
		assert.Equal(t, -456, result)
	})

	t.Run("Invalid String Returns Zero", func(t *testing.T) {
		result := ParseInt("invalid")
		assert.Equal(t, 0, result)
	})

	t.Run("Empty String Returns Zero", func(t *testing.T) {
		result := ParseInt("")
		assert.Equal(t, 0, result)
	})

	t.Run("Float String Returns Truncated Integer", func(t *testing.T) {
		result := ParseInt("123.456")
		assert.Equal(t, 0, result) // strconv.Atoi returns error for floats
	})
}

func TestResponseSuccess(t *testing.T) {
	t.Run("Creates Success Response With String Data", func(t *testing.T) {
		response := ResponseSuccess("test message")

		assert.True(t, response.Status)
		assert.Equal(t, "test message", response.Data)
		assert.Empty(t, response.Message)
	})

	t.Run("Creates Success Response With Map Data", func(t *testing.T) {
		data := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		response := ResponseSuccess(data)

		assert.True(t, response.Status)
		assert.Equal(t, data, response.Data)
	})

	t.Run("Creates Success Response With Nil Data", func(t *testing.T) {
		response := ResponseSuccess(nil)

		assert.True(t, response.Status)
		assert.Nil(t, response.Data)
	})

	t.Run("Response Has Correct Type", func(t *testing.T) {
		response := ResponseSuccess("test")

		assert.IsType(t, models.APIResponse{}, response)
	})
}

func TestResponseError(t *testing.T) {
	t.Run("Creates Error Response With Message", func(t *testing.T) {
		response := ResponseError("error occurred")

		assert.False(t, response.Status)
		assert.Equal(t, "error occurred", response.Message)
		assert.Nil(t, response.Data)
	})

	t.Run("Creates Error Response With Empty Message", func(t *testing.T) {
		response := ResponseError("")

		assert.False(t, response.Status)
		assert.Equal(t, "", response.Message)
	})

	t.Run("Response Has Correct Type", func(t *testing.T) {
		response := ResponseError("test error")

		assert.IsType(t, models.APIResponse{}, response)
	})
}

func TestResponseStructure(t *testing.T) {
	t.Run("Success And Error Responses Have Different Status", func(t *testing.T) {
		successResp := ResponseSuccess("success")
		errorResp := ResponseError("error")

		assert.NotEqual(t, successResp.Status, errorResp.Status)
		assert.True(t, successResp.Status)
		assert.False(t, errorResp.Status)
	})

	t.Run("Success Response Has Data, Error Has Message", func(t *testing.T) {
		successResp := ResponseSuccess("test data")
		errorResp := ResponseError("test error")

		assert.NotNil(t, successResp.Data)
		assert.Empty(t, successResp.Message)

		assert.Nil(t, errorResp.Data)
		assert.NotEmpty(t, errorResp.Message)
	})
}
