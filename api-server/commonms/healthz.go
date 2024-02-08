/*
Copyright 2023-2024 VMware Inc.
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commonms

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"collie-api-server/config"
)

func newHealthzProvider(cfg config.Config, log logrus.FieldLogger) *HealthzProvider {
	return &HealthzProvider{
		cfg:             cfg,
		log:             log,
		initHardTimeout: cfg.Controller.PrepTimeout + cfg.Controller.InitialSleepDuration + cfg.Controller.InitializationTimeoutExtension,
	}
}

type HealthzProvider struct {
	cfg             config.Config
	log             logrus.FieldLogger
	initHardTimeout time.Duration

	initializeStartedAt *time.Time
	lastHealthyActionAt *time.Time
}

func (h *HealthzProvider) Check(_ *http.Request) (err error) {
	defer func() {
		if err != nil {
			h.log.Warnf("Health check failed due to: %v", err)
		}
	}()

	if h.lastHealthyActionAt != nil {
		if time.Since(*h.lastHealthyActionAt) > h.cfg.Controller.HealthySnapshotIntervalLimit {
			return fmt.Errorf("time since initialization or last snapshot sent is over the considered healthy limit of %s", h.cfg.Controller.HealthySnapshotIntervalLimit)
		}
		return nil
	}

	if h.initializeStartedAt != nil {
		if time.Since(*h.initializeStartedAt) > h.initHardTimeout {
			return fmt.Errorf("controller initialization is taking longer than the hard timeout of %s", h.initHardTimeout)
		}
		return nil
	}

	return nil
}

func (h *HealthzProvider) Initializing() {
	if h.initializeStartedAt == nil {
		h.initializeStartedAt = nowPtr()
		h.lastHealthyActionAt = nil
	}
}

func (h *HealthzProvider) Initialized() {
	h.healthyAction()
}

func (h *HealthzProvider) SnapshotSent() {
	h.healthyAction()
}

func (h *HealthzProvider) healthyAction() {
	h.initializeStartedAt = nil
	h.lastHealthyActionAt = nowPtr()
}

func nowPtr() *time.Time {
	now := time.Now()
	return &now
}

func runHealthzEndpoints(cfg config.Config, log *logrus.Entry, controllerCheck healthz.Checker, exitCh chan error) func() {
	log.Infof("starting healthz on port: %d", cfg.HealthzPort)
	healthzSrv := &http.Server{Addr: portToServerAddr(cfg.HealthzPort), Handler: &healthz.Handler{Checks: map[string]healthz.Checker{
		"server":     healthz.Ping,
		"controller": controllerCheck,
	}}}
	closeFunc := func() {
		if err := healthzSrv.Close(); err != nil {
			log.Errorf("closing healthz server: %v", err)
		}
	}

	go func() {
		exitCh <- fmt.Errorf("healthz server: %w", healthzSrv.ListenAndServe())
	}()
	return closeFunc
}

func portToServerAddr(port int) string {
	return fmt.Sprintf(":%d", port)
}

func StartHealthz(cfg config.Config, log *logrus.Entry, exitCh chan error) func() {
	ctrlHealthz := newHealthzProvider(cfg, log)
	closeHealthz := runHealthzEndpoints(cfg, log, ctrlHealthz.Check, exitCh)
	return closeHealthz
}
