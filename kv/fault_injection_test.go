// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package kv_test

import (
	"context"

	. "github.com/pingcap/check"
	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/parser/terror"
)

type testFaultInjectionSuite struct{}

var _ = Suite(testFaultInjectionSuite{})

func (s testFaultInjectionSuite) TestFaultInjectionBasic(c *C) {
	var cfg kv.InjectionConfig
	err1 := errors.New("foo")
	cfg.SetGetError(err1)
	cfg.SetCommitError(err1)

	storage := kv.NewInjectedStore(kv.NewMockStorage(), &cfg)
	txn, err := storage.Begin()
	c.Assert(err, IsNil)
	_, err = storage.BeginWithStartTS(0)
	c.Assert(err, IsNil)
	ver := kv.Version{Ver: 1}
	snap, err := storage.GetSnapshot(ver)
	c.Assert(err, IsNil)
	b, err := txn.Get(context.TODO(), []byte{'a'})
	c.Assert(err.Error(), Equals, err1.Error())
	c.Assert(b, IsNil)
	b, err = snap.Get(context.TODO(), []byte{'a'})
	c.Assert(err.Error(), Equals, err1.Error())
	c.Assert(b, IsNil)

	err = txn.Commit(context.Background())
	c.Assert(err.Error(), Equals, err1.Error())

	cfg.SetGetError(nil)
	cfg.SetCommitError(nil)

	storage = kv.NewInjectedStore(kv.NewMockStorage(), &cfg)
	txn, err = storage.Begin()
	c.Assert(err, IsNil)
	snap, err = storage.GetSnapshot(ver)
	c.Assert(err, IsNil)

	b, err = txn.Get(context.TODO(), []byte{'a'})
	c.Assert(err, IsNil)
	c.Assert(b, IsNil)

	err = txn.Commit(context.Background())
	c.Assert(err, NotNil)
	c.Assert(terror.ErrorEqual(err, kv.ErrTxnRetryable), IsTrue)

}
