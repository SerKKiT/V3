package middleware

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// Validator содержит правила валидации
type Validator struct {
	usernameRegex *regexp.Regexp
	emailRegex    *regexp.Regexp
}

// NewValidator создаёт новый validator
func NewValidator() *Validator {
	return &Validator{
		usernameRegex: regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`),
		emailRegex:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	}
}

// ValidateAuthInput валидирует входные данные для register/login
func (v *Validator) ValidateAuthInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Читаем body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		// Восстанавливаем body для последующего чтения
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Парсим JSON
		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			c.Abort()
			return
		}

		// Валидация username (если присутствует)
		if username, ok := input["username"].(string); ok {
			if err := v.validateUsername(username); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}

		// Валидация email
		if email, ok := input["email"].(string); ok {
			if err := v.validateEmail(email); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}

		// Валидация password
		if password, ok := input["password"].(string); ok {
			if err := v.validatePassword(password); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}

		// Восстанавливаем оригинальный body для proxy
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		c.Next()
	}
}

// ValidateStreamInput валидирует входные данные для создания/обновления контента
func (v *Validator) ValidateStreamInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Читаем body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		// Восстанавливаем body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Парсим JSON
		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			c.Abort()
			return
		}

		// Валидация title
		if title, ok := input["title"].(string); ok {
			if err := v.validateTitle(title); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
			// ✅ АКТИВНАЯ САНАЦИЯ: Экранируем HTML теги
			input["title"] = v.sanitizeString(title)
		}

		// Валидация description
		if description, ok := input["description"].(string); ok {
			if err := v.validateDescription(description); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
			// ✅ АКТИВНАЯ САНАЦИЯ: Экранируем HTML теги
			input["description"] = v.sanitizeString(description)
		}

		// ✅ Конвертируем sanitized data обратно в JSON
		sanitizedBody, err := json.Marshal(input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process input"})
			c.Abort()
			return
		}

		// Восстанавливаем модифицированный body для proxy
		c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitizedBody))

		c.Next()
	}
}

// Валидационные функции

func (v *Validator) validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if len(username) < 3 {
		return &ValidationError{"Username must be at least 3 characters"}
	}
	if len(username) > 30 {
		return &ValidationError{"Username must be at most 30 characters"}
	}
	if !v.usernameRegex.MatchString(username) {
		return &ValidationError{"Username can only contain letters, numbers, underscores and hyphens"}
	}
	return nil
}

func (v *Validator) validateEmail(email string) error {
	email = strings.TrimSpace(email)

	if len(email) == 0 {
		return &ValidationError{"Email is required"}
	}
	if len(email) > 254 {
		return &ValidationError{"Email is too long"}
	}
	if !v.emailRegex.MatchString(email) {
		return &ValidationError{"Invalid email format"}
	}
	return nil
}

func (v *Validator) validatePassword(password string) error {
	if len(password) < 8 {
		return &ValidationError{"Password must be at least 8 characters"}
	}
	if len(password) > 72 {
		return &ValidationError{"Password is too long (max 72 characters)"}
	}

	hasLetter := false
	hasDigit := false

	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return &ValidationError{"Password must contain at least one letter and one digit"}
	}

	return nil
}

func (v *Validator) validateTitle(title string) error {
	title = strings.TrimSpace(title)

	if len(title) < 3 {
		return &ValidationError{"Title must be at least 3 characters"}
	}
	if len(title) > 100 {
		return &ValidationError{"Title must be at most 100 characters"}
	}
	if utf8.RuneCountInString(title) < 3 {
		return &ValidationError{"Title is too short"}
	}
	return nil
}

func (v *Validator) validateDescription(description string) error {
	description = strings.TrimSpace(description)

	if len(description) > 5000 {
		return &ValidationError{"Description is too long (max 5000 characters)"}
	}
	return nil
}

// sanitizeString защищает от XSS атак
func (v *Validator) sanitizeString(input string) string {
	input = strings.TrimSpace(input)
	input = html.EscapeString(input)

	// Удаление control characters
	input = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)

	return input
}

// ValidationError кастомная ошибка валидации
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
