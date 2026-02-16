package http_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	httpcontroller "invento-service/internal/controller/http"
	"invento-service/internal/dto"
	app_testing "invento-service/internal/testing"
)

func TestProjectController_GetList_PaginationBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		limit         int
		expectError   bool
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "Very large page",
			page:          999999,
			limit:         10,
			expectError:   false,
			expectedPage:  999999,
			expectedLimit: 10,
		},
		{
			name:          "Very large limit",
			page:          1,
			limit:         1000,
			expectError:   false,
			expectedPage:  1,
			expectedLimit: 1000,
		},
		{
			name:          "Negative page",
			page:          -1,
			limit:         10,
			expectError:   false,
			expectedPage:  1, // Default
			expectedLimit: 10,
		},
		{
			name:          "Negative limit",
			page:          1,
			limit:         -1,
			expectError:   false,
			expectedPage:  1,
			expectedLimit: 10, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(MockProjectUsecase)
			controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

			expectedData := &dto.ProjectListData{
				Items: []dto.ProjectListItem{},
				Pagination: dto.PaginationData{
					Page:       tt.expectedPage,
					Limit:      tt.expectedLimit,
					TotalItems: 0,
					TotalPages: 0,
				},
			}

			mockUC.On("GetList", "user-1", "", 0, "", tt.expectedPage, tt.expectedLimit).Return(expectedData, nil)

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_id", "user-1")
				c.Locals("user_email", "test@example.com")
				c.Locals("user_role", "user")
				return c.Next()
			})
			app.Get("/api/v1/project", controller.GetList)

			token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
			url := "/api/v1/project"
			if tt.page > 0 {
				url += fmt.Sprintf("?page=%d", tt.page)
			}
			if tt.limit > 0 {
				if strings.Contains(url, "?") {
					url += fmt.Sprintf("&limit=%d", tt.limit)
				} else {
					url += fmt.Sprintf("?limit=%d", tt.limit)
				}
			}
			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			resp, err := app.Test(req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
			} else {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			}

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)

			mockUC.AssertExpectations(t)
		})
	}
}
