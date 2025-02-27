// Copyright © 2022 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
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

package sqlcommon

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly/internal/coremsgs"
	"github.com/hyperledger/firefly/pkg/core"
	"github.com/hyperledger/firefly/pkg/database"
)

var (
	nextpinColumns = []string{
		"namespace",
		"context",
		"identity",
		"hash",
		"nonce",
	}
)

const nextpinsTable = "nextpins"

func (s *SQLCommon) InsertNextPin(ctx context.Context, nextpin *core.NextPin) (err error) {
	ctx, tx, autoCommit, err := s.beginOrUseTx(ctx)
	if err != nil {
		return err
	}
	defer s.rollbackTx(ctx, tx, autoCommit)

	sequence, err := s.insertTx(ctx, nextpinsTable, tx,
		sq.Insert(nextpinsTable).
			Columns(nextpinColumns...).
			Values(
				nextpin.Namespace,
				nextpin.Context,
				nextpin.Identity,
				nextpin.Hash,
				nextpin.Nonce,
			),
		nil, // no change events for next pins
	)
	if err != nil {
		return err
	}
	nextpin.Sequence = sequence

	return s.commitTx(ctx, tx, autoCommit)
}

func (s *SQLCommon) nextpinResult(ctx context.Context, row *sql.Rows) (*core.NextPin, error) {
	nextpin := core.NextPin{}
	err := row.Scan(
		&nextpin.Namespace,
		&nextpin.Context,
		&nextpin.Identity,
		&nextpin.Hash,
		&nextpin.Nonce,
		&nextpin.Sequence,
	)
	if err != nil {
		return nil, i18n.WrapError(ctx, err, coremsgs.MsgDBReadErr, nextpinsTable)
	}
	return &nextpin, nil
}

func (s *SQLCommon) GetNextPinsForContext(ctx context.Context, namespace string, context *fftypes.Bytes32) (message []*core.NextPin, err error) {

	cols := append([]string{}, nextpinColumns...)
	cols = append(cols, sequenceColumn)

	rows, _, err := s.query(ctx, nextpinsTable, sq.Select(cols...).From(nextpinsTable).
		Where(sq.And{
			sq.Eq{"context": context},
			sq.Or{sq.Eq{"namespace": namespace}, sq.Eq{"namespace": nil}}, // old entries will have NULL for namespace
		}))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nextpin := []*core.NextPin{}
	for rows.Next() {
		d, err := s.nextpinResult(ctx, rows)
		if err != nil {
			return nil, err
		}
		nextpin = append(nextpin, d)
	}

	return nextpin, err

}

func (s *SQLCommon) UpdateNextPin(ctx context.Context, namespace string, sequence int64, update database.Update) (err error) {

	ctx, tx, autoCommit, err := s.beginOrUseTx(ctx)
	if err != nil {
		return err
	}
	defer s.rollbackTx(ctx, tx, autoCommit)

	query, err := s.buildUpdate(sq.Update(nextpinsTable), update, pinFilterFieldMap)
	if err != nil {
		return err
	}
	query = query.Set("namespace", namespace) // always populate namespace (to migrate the table over time)
	query = query.Where(sq.Eq{sequenceColumn: sequence})

	_, err = s.updateTx(ctx, nextpinsTable, tx, query, nil /* no change events for next pins */)
	if err != nil {
		return err
	}

	return s.commitTx(ctx, tx, autoCommit)
}
