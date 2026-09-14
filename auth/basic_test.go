// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/moogar0880/problems"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/log"
)

func TestBasicAuthorizer_GetGinHandler(t *testing.T) {
	testAuthorizer := &BasicAuthorizer{
		logger: log.Named("test"),
		users: map[string]*basicAuthUser{
			"foo": &basicAuthUser{
				// Pa55w0rd$
				Password: "$2b$12$yH/i2alYaIrVbKFkaYu5HOSf3JiZ0zPJlooufSRdO.6V3X/hXgTOq",
				Roles: []string{"test"},
			},
			"bar": &basicAuthUser{
				// l3Tm31n!!!
				Password: "$2b$12$AiqmFNuYYubFwyrL3be4Pe21yqwwIshQ0VQjYeKSshcp1r37zWovO",
				Roles: []string{"test2"},

			},
		},
	}

	authHandler := testAuthorizer.GetGinHandler("test")

	testCases := []struct{
		title string
		header string
		err string
	}{
		{
			title: "ok",
			header: buildAuthHeader("foo", "Pa55w0rd$"),
		},
		{
			title: "err no auth",
			err: "authorization failed",
		},
		{
			title: "err bad password",
			header: buildAuthHeader("foo", "password"),
			err: "authorization failed",
		},
		{
			title: "err unknown user",
			header: buildAuthHeader("baz", "Pa55w0rd$"),
			err: "authorization failed",
		},
		{
			title: "err no role",
			header: buildAuthHeader("bar", "l3Tm31n!!!"),
			err: "API unauthorized for user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(writer)

			ginContext.Request, _ = http.NewRequest(http.MethodPost, "/", http.NoBody)

			if tc.header != "" {
				ginContext.Request.Header.Add("Authorization", tc.header)
			}

			authHandler(ginContext)
			result := writer.Result()

			if tc.err == "" {
				assert.Equal(t, "200 OK", result.Status)
			} else {
				assert.Equal(t, "401 Unauthorized", result.Status)

				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)

				var prob problems.DefaultProblem
				err = json.Unmarshal(body, &prob)
				require.NoError(t, err)
				assert.Equal(t, tc.err, prob.Detail)
			}

		})
	}
}

func buildAuthHeader(username, password string) string {
	toEncode := fmt.Sprintf("%s:%s", username, password)
	encoded := base64.StdEncoding.EncodeToString([]byte(toEncode))
	return fmt.Sprintf("Basic %s", encoded)
}
