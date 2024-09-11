// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/veraison/services/capability"
	"github.com/veraison/services/provisioning/provisioner"
	"github.com/veraison/services/verification/verifier"
	"go.uber.org/zap"
)

var (
	tenantID = "0"
)

type IHandler interface {
	Submit(c *gin.Context)
	GetWellKnownProvisioningInfo(c *gin.Context)
}

type Handler struct {
	Provisioner provisioner.IProvisioner

	logger *zap.SugaredLogger
}

func NewHandler(
	p provisioner.IProvisioner,
	logger *zap.SugaredLogger,
) IHandler {
	return &Handler{
		Provisioner: p,
		logger:      logger,
	}
}

type ProvisioningSession struct {
	Status        string  `json:"status"`
	Expiry        string  `json:"expiry"`
	FailureReason *string `json:"failure-reason,omitempty"`
}

const (
	ProvisioningSessionMediaType = "application/vnd.veraison.provisioning-session+json"
)

func (o *Handler) Submit(c *gin.Context) {
	// read the accept header and make sure that it's compatible with what we
	// support
	offered := c.NegotiateFormat(ProvisioningSessionMediaType)
	if offered != ProvisioningSessionMediaType {
		ReportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ProvisioningSessionMediaType),
		)
		return
	}

	// read media type
	mediaType := c.Request.Header.Get("Content-Type")

	isSupported, err := o.Provisioner.IsSupportedMediaType(mediaType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Unwrap(err) == verifier.ErrInputParam {
			status = http.StatusBadRequest
		}

		ReportProblem(c, status, fmt.Sprintf("could not check media type with verifier: %v", err))
		return
	}

	if !isSupported {
		supportedMediaTypes, err := o.Provisioner.SupportedMediaTypes()
		if err != nil {
			ReportProblem(c,
				http.StatusInternalServerError,
				fmt.Sprintf("could not get supported media types from provisioner: %v",
					err),
			)
			return
		}

		c.Header("Accept", strings.Join(supportedMediaTypes, ", "))
		ReportProblem(c,
			http.StatusUnsupportedMediaType,
			fmt.Sprintf("no active plugin found for %s", mediaType),
		)
		return
	}

	// read body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		ReportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("error reading body: %s", err),
		)
		return
	}

	if len(payload) == 0 {
		ReportProblem(c,
			http.StatusBadRequest,
			"empty body",
		)
		return
	}

	err = o.Provisioner.SubmitEndorsements(tenantID, payload, mediaType)
	if err != nil {
		o.logger.Errorw("submit endorsement failed", "error", err)

		if errors.Is(err, errors.New("no connection")) {
			ReportProblem(c,
				http.StatusInternalServerError,
				err.Error(),
			)
			return
		}

		sendFailedProvisioningSession(
			c,
			fmt.Sprintf("submit endorsement returned error: %s", err),
		)
		return
	}

	sendSuccessfulProvisioningSession(c)
}

func sendFailedProvisioningSession(c *gin.Context, failureReason string) {
	c.Header("Content-Type", ProvisioningSessionMediaType)
	c.JSON(
		http.StatusOK,
		&ProvisioningSession{
			Status:        "failed",
			Expiry:        time.Now().Format(time.RFC3339),
			FailureReason: &failureReason,
		},
	)
}

func sendSuccessfulProvisioningSession(c *gin.Context) {
	c.Header("Content-Type", ProvisioningSessionMediaType)
	c.JSON(
		http.StatusOK,
		&ProvisioningSession{
			Status: "success",
			Expiry: time.Now().Format(time.RFC3339),
		},
	)
}

func (o *Handler) getProvisioningServerVersionAndState() (string, string, error) {
	vtsState, err := o.Provisioner.GetVTSState()
	if err != nil {
		return "", "", err
	}
	version := vtsState.ServerVersion
	state := vtsState.Status.String()
	return version, state, nil
}

func getProvisioningEndpoints() map[string]string {
	return publicApiMap
}

func (o *Handler) GetWellKnownProvisioningInfo(c *gin.Context) {
	offered := c.NegotiateFormat(capability.WellKnownMediaType)
	if offered != capability.WellKnownMediaType && offered != gin.MIMEJSON {
		ReportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", capability.WellKnownMediaType),
		)
		return
	}

	// Get provisioning media types
	mediaTypes, err := o.Provisioner.SupportedMediaTypes()
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	// Get provisioning server version and state
	version, state, err := o.getProvisioningServerVersionAndState()
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	// Get provisioning server API endpoints
	endpoints := getProvisioningEndpoints()

	// Get final object with well known information
	obj, err := capability.NewWellKnownInfoObj(nil, mediaTypes, nil, version, state, endpoints)
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	c.Header("Content-Type", capability.WellKnownMediaType)
	c.JSON(http.StatusOK, obj)
}
