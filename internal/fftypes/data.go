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

package fftypes

import (
	"context"
	"crypto/sha256"
	"encoding/json"

	"github.com/google/uuid"
)

type DataType string

const (
	DataTypeDefinition DataType = "definition"
	DataTypeJSON       DataType = "json"
	DataTypeBLOB       DataType = "blob"
)

type DataRef struct {
	ID   *uuid.UUID `json:"id,omitempty"`
	Hash *Bytes32   `json:"hash,omitempty"`
}

type Data struct {
	ID        *uuid.UUID `json:"id,omitempty"`
	Type      DataType   `json:"type"`
	Namespace string     `json:"namespace,omitempty"`
	Hash      *Bytes32   `json:"hash,omitempty"`
	Created   int64      `json:"created,omitempty"`
	Schema    *SchemaRef `json:"schema,omitempty"`
	Value     JSONData   `json:"value,omitempty"`
}

type SchemaRef struct {
	Entity  string `json:"entity,omitempty"`
	Version string `json:"version,omitempty"`
}

type DataRefs []DataRef

func (d DataRefs) Hash(ctx context.Context) *Bytes32 {
	b, _ := json.Marshal(&d)
	var b32 Bytes32 = sha256.Sum256(b)
	return &b32
}