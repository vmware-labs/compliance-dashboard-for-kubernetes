/*
Copyright 2023-2023 VMware Inc.
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
	"collie-api-server/config"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func RunApp(fnRun func(config.Config, *logrus.Entry, context.Context, chan error) error) {

	cfg := config.Get()

	logger := logrus.New()
	logger.SetLevel(logrus.Level(cfg.Log.Level))
	log := logger.WithField("version", "local")

	logBuildInfo(log)

	log.Printf("%v", cfg)

	ctx := signals.SetupSignalHandler()
	ctx, ctxCancel := context.WithCancel(ctx)
	defer ctxCancel()

	exitCh := make(chan error, 10)
	go watchExitErrors(ctx, log, exitCh, ctxCancel)
	closeHealthz := StartHealthz(cfg, log, exitCh)
	defer closeHealthz()

	if err := fnRun(cfg, log, ctx, exitCh); err != nil {
		log.Fatalf("agent failed: %v", err)
	}

	log.Println("Exit")
}

func logBuildInfo(log *logrus.Entry) {
	//   - vcs.revision: the revision identifier for the current commit or checkout
	//   - vcs.time: the modification time associated with vcs.revision, in RFC3339 format
	interestedFields := map[string]int{"vcs.revision": 1, "vcs.time": 1}
	if bi, ok := debug.ReadBuildInfo(); ok {
		log.Printf(bi.GoVersion)
		for _, v := range bi.Settings {
			if _, ok := interestedFields[v.Key]; ok {
				log.Println(v.Key, v.Value)
			}
		}
	}
}

// if any errors are observed on exitCh, context cancel is called, and all errors in the channel are logged
func watchExitErrors(ctx context.Context, log *logrus.Entry, exitCh chan error, ctxCancel func()) {
	select {
	case err := <-exitCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("Stopped with an error: %v", err)
		}
		ctxCancel()
	case <-ctx.Done():
		return
	}
}
