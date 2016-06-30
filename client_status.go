// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mendersoftware/log"
	"github.com/pkg/errors"
)

const (
	statusInstalling  = "installing"
	statusDownloading = "downloading"
	statusRebooting   = "rebooting"
	statusSuccess     = "success"
	statusFailure     = "failure"
)

type StatusReporter interface {
	Report(api ApiRequester, server string, report StatusReport) error
}

type StatusReport struct {
	deviceID     string
	deploymentID string
	Status       string `json:"status"`
}

type StatusClient struct {
}

func NewStatusClient() StatusReporter {
	return &StatusClient{}
}

// Report status information to the backend
func (u *StatusClient) Report(api ApiRequester, url string, report StatusReport) error {
	req, err := makeStatusReportRequest(url, report)
	if err != nil {
		return errors.Wrapf(err, "failed to prepare status report request")
	}

	r, err := api.Do(req)
	if err != nil {
		log.Error("failed to report status: ", err)
		return errors.Wrapf(err, "reporting status failed")
	}

	// HTTP 204 No Content
	if r.StatusCode != http.StatusNoContent {
		log.Errorf("got unexpected HTTP status when reporting status: %v", r.StatusCode)
		return errors.Errorf("reporting status failed, bad status %v", r.StatusCode)
	}
	log.Debugf("status reported, response %v", r)

	return nil
}

func makeStatusReportRequest(server string, report StatusReport) (*http.Request, error) {
	path := fmt.Sprintf("/deployments/devices/%s/deployments/%s/status",
		report.deviceID, report.deploymentID)
	url := buildApiURL(server, path)

	out := &bytes.Buffer{}
	enc := json.NewEncoder(out)
	enc.Encode(&report)

	return http.NewRequest(http.MethodPut, url, out)
}
