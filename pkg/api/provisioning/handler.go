// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package provisioning

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/services/pkg/api/common"
	"github.com/veraison/services/pkg/fault"
	"go.uber.org/zap"
)

var tenantID = "0"

type IHandler interface {
	Submit(c *gin.Context)
	SetActive(setActive bool) gin.HandlerFunc
	GetWellKnownProvisioningInfo(c *gin.Context)
}

type Handler struct {
	Provisioner IProvisioner

	WkCacheMaxAge  time.Duration
	MaxPayloadSize int64
	logger         *zap.SugaredLogger
}

func NewHandler(
	provisioner IProvisioner,
	logger *zap.SugaredLogger,
	wkCacheMaxAge time.Duration,
	maxPayloadSize int64,
) IHandler {
	return &Handler{
		Provisioner:    provisioner,
		logger:         logger,
		WkCacheMaxAge:  wkCacheMaxAge,
		MaxPayloadSize: maxPayloadSize,
	}
}

type ProvisioningSession struct {
	Status        string  `json:"status"`
	Expiry        string  `json:"expiry"`
	FailureReason *string `json:"failure-reason,omitempty"`
}

const (
	ProvisioningSessionMediaType = "application/vnd.veraison.provisioning-session+json"
	ELMQueryMediaType            = "application/vnd.veraison.elm-v1+cbor"
)

func (o *Handler) Submit(c *gin.Context) {
	offered := c.NegotiateFormat(ProvisioningSessionMediaType)
	if offered != ProvisioningSessionMediaType {
		common.ReportProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ProvisioningSessionMediaType),
		)
		return
	}

	mediaType := c.Request.Header.Get("Content-Type")

	isSupported, err := o.Provisioner.IsSupportedMediaType(mediaType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, fault.ErrBadInput) {
			status = http.StatusBadRequest
		}

		common.ReportProblem(o.logger, c, status, fmt.Sprintf(
			"could not check media type with provisioner: %v", err))
		return
	}

	if !isSupported {
		supportedMediaTypes, err := o.Provisioner.SupportedMediaTypes()
		if err != nil {
			common.ReportProblem(o.logger, c,
				http.StatusInternalServerError,
				fmt.Sprintf("could not get supported media types from provisioner: %v",
					err),
			)
			return
		}

		c.Header("Accept", strings.Join(supportedMediaTypes, ", "))
		common.ReportProblem(o.logger, c,
			http.StatusUnsupportedMediaType,
			fmt.Sprintf("no active plugin found for %s", mediaType),
		)
		return
	}

	// read body
	payload, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, o.MaxPayloadSize))
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusBadRequest,
			fmt.Sprintf("error reading body: %s", err),
		)
		return
	}

	if len(payload) == 0 {
		common.ReportProblem(o.logger, c,
			http.StatusBadRequest,
			"empty body",
		)
		return
	}

	err = o.Provisioner.SubmitEndorsements(tenantID, payload, mediaType)
	if err != nil {
		o.logger.Errorw("submit endorsement failed", "error", err)

		if errors.Is(err, fault.ErrNoConnection) {
			common.ReportProblem(o.logger, c,
				http.StatusInternalServerError,
				"something went wrong on our side, please try again",
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

func (o *Handler) SetActive(setActive bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// read media type
		mediaType := c.Request.Header.Get("Content-Type")
		if mediaType != ELMQueryMediaType {
			common.ReportConciseProblem(o.logger, c,
				http.StatusUnsupportedMediaType,
				fmt.Sprintf(
					"unknown media type: %v, expected: %v",
					mediaType,
					ELMQueryMediaType,
				),
			)
			return
		}

		// read body
		payload, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, o.MaxPayloadSize))
		if err != nil {
			common.ReportConciseProblem(o.logger, c,
				http.StatusBadRequest,
				fmt.Sprintf("error reading body: %s", err),
			)
			return
		}

		if len(payload) == 0 {
			common.ReportConciseProblem(o.logger, c,
				http.StatusBadRequest,
				"empty body",
			)
			return
		}

		err = o.Provisioner.SetEndorsementsState(tenantID, payload, setActive)
		if err != nil {
			action := "deactivate"
			if setActive {
				action = "activate"
			}

			o.logger.Errorw(action+" endorsement failed", "error", err)

			if errors.Is(err, fault.ErrNoConnection) {
				common.ReportProblem(o.logger, c,
					http.StatusInternalServerError,
					"something went wrong on our side, please try again",
				)
				return
			}

			// corimstore.ErrNoMatch type cannot be propagated beyond the
			// VTS gRPC boundary. So, match the error strings instead.
			if strings.Contains(err.Error(), store.ErrNoMatch.Error()) {
				common.ReportConciseProblem(o.logger, c,
					http.StatusNotFound,
					"no endorsements matched the selection",
				)
				return
			}

			common.ReportConciseProblem(o.logger, c,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
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

func getProvisioningEndpoints() map[string]string {
	return publicApiMap
}

func (o *Handler) GetWellKnownProvisioningInfo(c *gin.Context) {
	offered := c.NegotiateFormat(common.WellKnownMediaType)
	if offered != common.WellKnownMediaType && offered != gin.MIMEJSON {
		common.ReportProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", common.WellKnownMediaType),
		)
		return
	}

	// Get provisioning media types
	mediaTypes, err := o.Provisioner.SupportedMediaTypes()
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	// Get provisioning server version and state
	version, err := o.Provisioner.ServerVersion()
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	// Get provisioning server API endpoints
	endpoints := getProvisioningEndpoints()

	// Get final object with well known information
	obj, err := common.NewWellKnownInfoObj(nil, mediaTypes, nil, version, endpoints)
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", int64(o.WkCacheMaxAge.Seconds())))
	c.Header("Expires", time.Now().Add(o.WkCacheMaxAge).UTC().Format(time.RFC1123))
	c.Header("Content-Type", common.WellKnownMediaType)
	c.JSON(http.StatusOK, obj)
}
