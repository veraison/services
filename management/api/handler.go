// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moogar0880/problems"
	"github.com/veraison/services/capability"
	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/management"
	"github.com/veraison/services/policy"
	"go.uber.org/zap"
)

const (
	RulesMediaType    = "application/vnd.veraison.policy.opa"
	PolicyMediaType   = "application/vnd.veraison.policy+json"
	PoliciesMediaType = "application/vnd.veraison.policies+json"
)

var (
	tenantID = "0"
)

type Handler struct {
	Manager *management.PolicyManager
	Logger  *zap.SugaredLogger
}

func NewHandler(manager *management.PolicyManager, logger *zap.SugaredLogger) Handler {
	return Handler{
		Manager: manager,
		Logger:  logger,
	}
}

func (o Handler) CreatePolicy(c *gin.Context) {
	offered := c.NegotiateFormat(PolicyMediaType)
	if offered != PolicyMediaType {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s",
				PolicyMediaType),
		)
		return
	}

	mediaType := c.Request.Header.Get("Content-Type")
	if mediaType != RulesMediaType {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("the only supported rules format is %s",
				RulesMediaType),
		)
		return
	}

	scheme := c.Param("scheme")
	if !o.Manager.IsSchemeSupported(scheme) {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("unrecognised scheme %q", scheme),
		)
		return
	}

	name := c.Query("name")
	if name == "" {
		name = "default"
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		reportProblem(c, http.StatusBadRequest, fmt.Sprintf("error reading body: %s", err))
		return
	}

	if len(payload) == 0 {
		reportProblem(c, http.StatusBadRequest, "empty body")
		return
	}

	policyRules := string(payload)

	if err = o.Manager.Validate(c, policyRules); err != nil {
		reportProblem(c, http.StatusBadRequest, fmt.Sprintf("invalid policy: %s", err))
	}

	policy, err := o.Manager.Update(c, tenantID, scheme, name, policyRules)
	if err != nil {
		reportProblem(c,
			http.StatusInternalServerError,
			fmt.Sprintf("could not update policy: %s", err),
		)
	}

	respBytes, err := json.Marshal(&policy)
	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
	}

	c.Data(http.StatusCreated, PolicyMediaType, respBytes)
}

func (o Handler) GetActivePolicy(c *gin.Context) {
	offered := c.NegotiateFormat(PolicyMediaType)
	if offered != PolicyMediaType {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s",
				PolicyMediaType),
		)
		return
	}

	scheme := c.Param("scheme")
	if !o.Manager.IsSchemeSupported(scheme) {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("unrecognised scheme %q", scheme),
		)
		return
	}

	pol, err := o.Manager.GetActive(c, tenantID, scheme)
	o.respondToGet(c, PolicyMediaType, pol, err)
}

func (o Handler) GetPolicy(c *gin.Context) {
	offered := c.NegotiateFormat(PolicyMediaType)
	if offered != PolicyMediaType {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s",
				PolicyMediaType),
		)
		return
	}

	scheme := c.Param("scheme")
	if !o.Manager.IsSchemeSupported(scheme) {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("unrecognised scheme %q", scheme),
		)
		return
	}

	uuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("bad UUID %q", c.Param("uuid")),
		)
		return
	}

	pol, err := o.Manager.GetPolicy(c, tenantID, scheme, uuid)
	o.respondToGet(c, PolicyMediaType, pol, err)
}

func (o Handler) GetPolicies(c *gin.Context) {
	offered := c.NegotiateFormat(PoliciesMediaType)
	if offered != PoliciesMediaType {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s",
				PoliciesMediaType),
		)
		return
	}

	scheme := c.Param("scheme")
	if !o.Manager.IsSchemeSupported(scheme) {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("unrecognised scheme %q", scheme),
		)
		return
	}

	policies, err := o.Manager.GetPolicies(c, tenantID, scheme, c.Query("name"))
	o.respondToGet(c, PoliciesMediaType, policies, err)
}

func (o Handler) Activate(c *gin.Context) {
	scheme := c.Param("scheme")
	if !o.Manager.IsSchemeSupported(scheme) {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("unrecognised scheme %q", scheme),
		)
		return
	}

	uuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("bad UUID %q", c.Param("uuid")),
		)
		return
	}

	err = o.Manager.Activate(c, tenantID, scheme, uuid)
	o.respondSimple(c, err)
}

func (o Handler) DeactivateAll(c *gin.Context) {
	scheme := c.Param("scheme")
	if !o.Manager.IsSchemeSupported(scheme) {
		reportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("unrecognised scheme %q", scheme),
		)
		return
	}

	err := o.Manager.DeactivateAll(c, tenantID, scheme)
	o.respondSimple(c, err)
}

func (o Handler) respondSimple(c *gin.Context, err error) {
	if err == nil {
		c.Status(http.StatusOK)
	} else {
		if errors.Is(err, policy.ErrNoPolicy) {
			reportProblem(c, http.StatusNotFound, err.Error())
		} else {
			reportProblem(c, http.StatusInternalServerError, err.Error())
		}
	}
}

func (o Handler) GetManagementWellKnownInfo(c *gin.Context) {
	offered := c.NegotiateFormat(capability.WellKnownMediaType)
	if offered != capability.WellKnownMediaType && offered != gin.MIMEJSON {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s",
				capability.WellKnownMediaType),
		)
		return
	}

	obj, err := capability.NewWellKnownInfoObj(
		nil, // key
		nil, // media types
		o.Manager.SupportedSchemes,
		config.Version,
		"SERVICE_STATUS_READY",
		publicApiMap,
	)

	if err != nil {
		reportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	c.Header("Content-Type", capability.WellKnownMediaType)
	c.JSON(http.StatusOK, obj)

}

func (o Handler) respondToGet(c *gin.Context, mt string, ret interface{}, err error) {
	if err != nil {
		if errors.Is(err, policy.ErrNoPolicy) || errors.Is(err, policy.ErrNoActivePolicy) {
			reportProblem(c, http.StatusNotFound, err.Error())
		} else {
			reportProblem(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respBytes, err := json.Marshal(ret)
	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
	}

	c.Data(http.StatusOK, mt, respBytes)
}

func reportProblem(c *gin.Context, status int, details ...string) {
	prob := problems.NewStatusProblem(status)

	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	log.LogProblem(log.Named("api"), prob)

	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, prob)
}
