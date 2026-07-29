// Copyright 2004 Ryan Forte
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rdforte/gomaxecs/internal/client"
)

// taskMeta represents the ECS Task Metadata.
type taskMeta struct {
	Containers []container `json:"Containers"`
	Limits     limit       `json:"Limits"` // this is optional in the response
}

// container represents the ECS Container Metadata.
type container struct {
	//nolint:tagliatelle // ECS Agent inconsistency. All fields adhere to goPascal but this one.
	DockerID string `json:"DockerId"`
	Limits   limit  `json:"Limits"`
}

// limit contains the CPU limit.
type limit struct {
	CPU float64 `json:"CPU"`
}

// Grab the container metadata from the ECS Metadata endpoint.
// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4-examples.html
func (t *Task) getContainerMeta(ctx context.Context) (container, error) {
	t.log("Fetching container metadata from %s", t.containerMetadataURI)
	return getMeta[container](ctx, t.client, t.containerMetadataURI)
}

// Grab the task metadata from the ECS Metadata endpoint + `/task`
// This will also include the container metadata.
// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4-examples.html
// #task-metadata-endpoint-v4-example-task-metadata-response.
func (t *Task) getTaskMeta(ctx context.Context) (taskMeta, error) {
	t.log("Fetching task metadata from %s", t.taskMetadataURI)
	return getMeta[taskMeta](ctx, t.client, t.taskMetadataURI)
}

func getMeta[T any](ctx context.Context, client *client.Client, url string) (T, error) {
	var res T

	resp, err := client.Get(ctx, url)
	if err != nil {
		return res, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return res, newStatusError(resp.StatusCode)
	}

	err = json.Unmarshal(resp.Body, &res)
	if err != nil {
		return res, fmt.Errorf("unmarshal failed: %w", err)
	}

	return res, nil
}

func newStatusError(status int) error {
	return &statusError{status}
}

type statusError struct {
	status int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("request failed, status code: %d", e.status)
}
