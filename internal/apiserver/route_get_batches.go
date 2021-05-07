// Copyright © 2021 Kaleido, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiserver

import (
	"net/http"

	"github.com/kaleido-io/firefly/internal/apispec"
	"github.com/kaleido-io/firefly/internal/fftypes"
	"github.com/kaleido-io/firefly/internal/i18n"
	"github.com/kaleido-io/firefly/internal/persistence"
)

var getBatches = &apispec.Route{
	Name:   "getBatches",
	Path:   "ns/{ns}/batches",
	Method: http.MethodGet,
	PathParams: []apispec.PathParam{
		{Name: "ns", Description: i18n.MsgTBD},
	},
	QueryParams:     nil,
	FilterFactory:   persistence.BatchFilterBuilder,
	Description:     i18n.MsgTBD,
	JSONInputValue:  func() interface{} { return nil },
	JSONOutputValue: func() interface{} { return []*fftypes.Batch{} },
	JSONOutputCode:  http.StatusOK,
	JSONHandler: func(r apispec.APIRequest) (output interface{}, err error) {
		output, err = r.E.GetBatches(r.Ctx, r.PP["ns"], r.Filter)
		return output, err
	},
}