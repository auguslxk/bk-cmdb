/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */
package instances

import (
	"fmt"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/errors"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/source_controller/coreservice/core"
	"configcenter/src/storage/dal"
)

type synchronizeInstanceAdapterError struct {
	err      errors.CCError
	instInfo *metadata.SynchronizeItem
}

type synchronizeInstanceAdapterDBParameter struct {
	tableName   string
	InstIDField string
}

type synchronizeInstanceAdapter struct {
	dbProxy    dal.RDB
	syncData   *metadata.SynchronizeInstanceParameter
	errorArray map[int64]synchronizeInstanceAdapterError
}

func newSynchronizeInstanceAdapter(syncData *metadata.SynchronizeInstanceParameter, dbProxy dal.RDB) *synchronizeInstanceAdapter {
	return &synchronizeInstanceAdapter{
		syncData: syncData,
		dbProxy:  dbProxy,
	}
}

func (s *synchronizeInstanceAdapter) PreSynchronizeFilter(ctx core.ContextParams) errors.CCError {
	if s.syncData.SynchronizeSign == "" {
		// TODO  return error not synchronize sign
		return ctx.Error.Errorf(common.CCErrCommParamsNeedSet, "sync_sign")
	}
	if s.syncData.InstacneInfoArray == nil {
		// TODO return error not found synchroize data
		return ctx.Error.Errorf(common.CCErrCommParamsNeedSet, "instance_info_array")
	}
	var syncDataInstArr []*metadata.SynchronizeItem
	for _, item := range s.syncData.InstacneInfoArray {
		if !item.Info.IsEmpty() {
			syncDataInstArr = append(syncDataInstArr, item)
		}
	}
	s.syncData.InstacneInfoArray = syncDataInstArr
	// synchronize data need to write data,append synchronize sign to metada
	if s.syncData.OperateType != metadata.SynchronizeOperateTypeUpdate {
		// set synchroize sign to instance metadata
		for _, item := range s.syncData.InstacneInfoArray {
			if item.Info.Exists(common.MetadataField) {
				metadata, err := item.Info.MapStr(common.MetaDataSynchronizeSignField)
				if err != nil {
					// TODO addd error  get metadata error
					blog.Errorf("preSynchronizeFilter get %s field error, inst info:%#v,rid:%s", common.MetaDataSynchronizeSignField, item, ctx.ReqID)
					s.errorArray[item.ID] = synchronizeInstanceAdapterError{
						instInfo: item,
						err:      ctx.Error.Errorf(common.CCErrCommInstFieldConvFail, s.syncData.ObjectID, common.MetaDataSynchronizeSignField, "mapstr", err.Error()),
					}
					continue
				}
				metadata.Set(common.MetaDataSynchronizeSignField, s.syncData.SynchronizeSign)
			} else {
				item.Info.Set(common.MetadataField, mapstr.MapStr{common.MetaDataSynchronizeSignField: s.syncData.SynchronizeSign})
			}
		}
	}

	return nil
}

func (s *synchronizeInstanceAdapter) SaveSynchronize(ctx core.ContextParams) {
	// Each model is written separately for subsequent expansion,
	// each model may be processed differently.
	switch s.syncData.ObjectID {
	case common.BKInnerObjIDApp:
		s.saveSynchronizeAppInstance(ctx)
	case common.BKInnerObjIDSet:
		s.saveSynchronizeSetInstance(ctx)
	case common.BKInnerObjIDModule:
		s.saveSynchronizeModuleInstance(ctx)
	case common.BKInnerObjIDProc:
		s.saveSynchronizeModuleInstance(ctx)
	case common.BKInnerObjIDPlat:
		s.saveSynchronizePlatInstance(ctx)
	case common.BKInnerObjIDHost:
		s.saveSynchronizeHostInstance(ctx)
	default:
		s.saveSynchronizeObjectInstance(ctx)

	}
}

func (s *synchronizeInstanceAdapter) GetErrorStringArr(ctx core.ContextParams) ([]string, errors.CCError) {
	if len(s.errorArray) == 0 {
		return make([]string, 0), nil
	}
	var errStrArr []string
	for _, err := range s.errorArray {
		errMsg := fmt.Sprintf("module[%s] instID:[%d] error:%s", s.syncData.ObjectID, err.instInfo.ID, err.instInfo.ID)
		errStrArr = append(errStrArr, errMsg)
	}
	return errStrArr, ctx.Error.Error(common.CCErrCoreServiceSyncInstError)
}

func (s *synchronizeInstanceAdapter) saveSynchronizeAppInstance(ctx core.ContextParams) {
	var dbParam synchronizeInstanceAdapterDBParameter
	dbParam.tableName = common.BKTableNameBaseApp
	dbParam.InstIDField = common.BKAppIDField
	s.saveSynchronizeInstance(ctx, dbParam)
}

func (s *synchronizeInstanceAdapter) saveSynchronizeSetInstance(ctx core.ContextParams) {
	var dbParam synchronizeInstanceAdapterDBParameter
	dbParam.tableName = common.BKTableNameBaseSet
	dbParam.InstIDField = common.BKSetIDField
	s.saveSynchronizeInstance(ctx, dbParam)
}

func (s *synchronizeInstanceAdapter) saveSynchronizeModuleInstance(ctx core.ContextParams) {
	var dbParam synchronizeInstanceAdapterDBParameter
	dbParam.tableName = common.BKTableNameBaseModule
	dbParam.InstIDField = common.BKModuleIDField
	s.saveSynchronizeInstance(ctx, dbParam)
}

func (s *synchronizeInstanceAdapter) saveSynchronizeProcessInstance(ctx core.ContextParams) {
	var dbParam synchronizeInstanceAdapterDBParameter
	dbParam.tableName = common.BKTableNameBaseProcess
	dbParam.InstIDField = common.BKProcIDField
	s.saveSynchronizeInstance(ctx, dbParam)
}

func (s *synchronizeInstanceAdapter) saveSynchronizePlatInstance(ctx core.ContextParams) {
	var dbParam synchronizeInstanceAdapterDBParameter
	dbParam.tableName = common.BKTableNameBasePlat
	dbParam.InstIDField = common.BKCloudIDField
	s.saveSynchronizeInstance(ctx, dbParam)
}

func (s *synchronizeInstanceAdapter) saveSynchronizeHostInstance(ctx core.ContextParams) {
	var dbParam synchronizeInstanceAdapterDBParameter
	dbParam.tableName = common.BKTableNameBaseHost
	dbParam.InstIDField = common.BKHostIDField
	s.saveSynchronizeInstance(ctx, dbParam)
}

func (s *synchronizeInstanceAdapter) saveSynchronizeObjectInstance(ctx core.ContextParams) {
	var dbParam synchronizeInstanceAdapterDBParameter
	dbParam.tableName = common.BKTableNameBaseInst
	dbParam.InstIDField = common.BKInstIDField
	s.saveSynchronizeInstance(ctx, dbParam)
}

func (s *synchronizeInstanceAdapter) saveSynchronizeInstance(ctx core.ContextParams, dbParam synchronizeInstanceAdapterDBParameter) {
	switch s.syncData.OperateType {
	case metadata.SynchronizeOperateTypeDelete:
		s.deleteSynchronizeInstacne(ctx, dbParam)
	case metadata.SynchronizeOperateTypeUpdate, metadata.SynchronizeOperateTypeAdd, metadata.SynchronizeOperateTypeRepalce:
		s.replaceSynchronizeInstance(ctx, dbParam)

	}
}

func (s *synchronizeInstanceAdapter) replaceSynchronizeInstance(ctx core.ContextParams, dbParam synchronizeInstanceAdapterDBParameter) {
	for _, item := range s.syncData.InstacneInfoArray {
		_, ok := s.errorArray[item.ID]
		if ok {
			continue
		}
		conds := mapstr.MapStr{dbParam.InstIDField: item.ID}
		exist, err := s.existSynchronizeInstanceID(ctx, dbParam.tableName, conds)
		if err != nil {
			blog.Errorf("replaceSynchronizeInstance existSynchronizeInstanceID error.objID:%s,info:%#v,rid:%s", s.syncData.ObjectID, item, ctx.ReqID)
			s.errorArray[item.ID] = synchronizeInstanceAdapterError{
				instInfo: item,
				err:      err,
			}
			continue
		}
		if exist {
			err := s.dbProxy.Table(dbParam.tableName).Update(ctx, conds, item.Info)
			if err != nil {
				blog.Errorf("replaceSynchronizeInstance update info error,err:%s.objID:%s,condition:%#v,info:%#v,rid:%s", err.Error(), s.syncData.ObjectID, conds, item, ctx.ReqID)
				s.errorArray[item.ID] = synchronizeInstanceAdapterError{
					instInfo: item,
					err:      ctx.Error.Error(common.CCErrCommDBUpdateFailed),
				}
				continue
			}
		} else {
			err := s.dbProxy.Table(dbParam.tableName).Insert(ctx, item.Info)
			if err != nil {
				blog.Errorf("replaceSynchronizeInstance insert info error,err:%s.objID:%s,info:%#v,rid:%s", err.Error(), s.syncData.ObjectID, item, ctx.ReqID)
				s.errorArray[item.ID] = synchronizeInstanceAdapterError{
					instInfo: item,
					err:      ctx.Error.Error(common.CCErrCommDBInsertFailed),
				}
				continue
			}
		}
	}
}

func (s *synchronizeInstanceAdapter) deleteSynchronizeInstacne(ctx core.ContextParams, dbParam synchronizeInstanceAdapterDBParameter) {
	var instIDArr []int64
	for _, item := range s.syncData.InstacneInfoArray {
		instIDArr = append(instIDArr, item.InstanceID)
	}
	err := s.dbProxy.Table(dbParam.tableName).Delete(ctx, mapstr.MapStr{dbParam.InstIDField: mapstr.MapStr{common.BKDBIN: instIDArr}})
	if err != nil {
		blog.Errorf("deleteSynchronizeInstacne delete info error,err:%s.objID:%s,instIDArr:%#v,rid:%s", err.Error(), s.syncData.ObjectID, instIDArr, ctx.ReqID)
		for _, item := range s.syncData.InstacneInfoArray {
			s.errorArray[item.InstanceID] = synchronizeInstanceAdapterError{
				instInfo: item,
				err:      ctx.Error.Error(common.CCErrCommDBDeleteFailed),
			}
		}
	}
}

func (s *synchronizeInstanceAdapter) existSynchronizeInstanceID(ctx core.ContextParams, tableName string, conds mapstr.MapStr) (bool, errors.CCError) {
	cnt, err := s.dbProxy.Table(tableName).Find(conds).Count(ctx)
	if err != nil {
		blog.Errorf("existSynchronizeInstanceID error. objectID:%s,conds:%#v,rid:%s", s.syncData.ObjectID, conds, ctx.ReqID)
		return false, ctx.Error.Error(common.CCErrCommDBSelectFailed)
	}
	if cnt > 0 {
		return true, nil
	}
	return false, nil

}
