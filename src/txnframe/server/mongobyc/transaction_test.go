/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongobyc

import (
	"context"
	"fmt"
	"testing"
)

func TestTransaction(t *testing.T) {

	InitMongoc()
	defer CleanupMongoc()

	mongo := NewClient("mongodb://test.mongoc:27017", "db_name_uri")
	if err := mongo.Open(); nil != err {
		fmt.Println("failed  open:", err)
		return
	}
	defer mongo.Close()

	cliSession := mongo.TransactionOperation().CreateSession()
	if err := cliSession.Start(); nil != err {
		t.Errorf("failed to  start session: %s", err.Error())
		return
	}
	defer cliSession.Close()

	txn := cliSession.CreateTransaction()
	if err := txn.Start(); nil != err {
		t.Errorf("failed to  start txn: %s", err.Error())
		return
	}

	txnCol := cliSession.CollectionWithSession("txn_uri")

	if err := txnCol.InsertOne(context.Background(), `{"txn":"txn_uri_val"}`); nil != err {
		t.Logf("err:%s", err.Error())
		return
	}
	if err := txn.Commit(); nil != err {
		t.Errorf("failed to  commit coll: %s", err.Error())
		return
	}

}
