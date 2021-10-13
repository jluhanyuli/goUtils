package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	MongoClient "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

// connection pool
var (
	connectPool         = make(map[string]*MongoClient.Collection, 0)
	connectMasterPool   = make(map[string]*MongoClient.Collection, 0)
	reloadMgoLock       = &sync.Mutex{}
	reloadMgoMasterLock = &sync.Mutex{}
)

const DefaultLimit = int64(1000)

const (
	StatusInvalid     = iota // 删除
	StatusValidWrite         // 可读可写
	StatusValidRead          // 只可读
	StatusDraftAdd           // 草稿增加态
	StatusDraftRemove        // 草稿删除态
)

const (
	_                    = iota
	FilterStatusGetField //获取字段为1
)

type CommonModel struct {
	CreateTime time.Time `bson:"create_time" json:"create_time"` // 创建时间
	ModifyTime time.Time `bson:"modify_time" json:"modify_time"` // 最后更新时间
	TenantId   string    `bson:"tenant_id" json:"tenant_id"`     // 租户id
	Status     uint16    `bson:"status" json:"status"`           // 记录是否有效
}

type CommonModelFilter interface {
	ToMongoFilter() interface{}
}

type CommonModelOption interface {
	ToMongoOption() []*options.FindOptions
}

type CommonModelObj interface {
	GetTableName() string
	GetId() string
	ToString() string
	GetCollection() *MongoClient.Collection
	GetInstance() interface{}
	GetInstanceList() interface{}
}

type CommonBatchUpdate interface {
	GetFilter(ctx context.Context) bson.M
	GetUpdate(ctx context.Context) bson.M
	GetCollection() *MongoClient.Collection
	IsVaildUpdate(ctx context.Context) bool
	InitModifyTime(ctx context.Context) CommonBatchUpdate
	SetFilter(string, interface{}) CommonBatchUpdate
	SetUpdate(key string, value interface{}) CommonBatchUpdate
}



func (m *CommonModel) IsSoftDeleted() bool {
	return m != nil && m.Status == StatusInvalid
}


//// 获取写连接
//func (m *CommonModel) GetMasterCollection(tableName string) *MongoClient.Collection {
//	collection, ok := connectMasterPool[tableName]
//	if !ok {
//		reloadMgoMasterLock.Lock()
//		defer reloadMgoMasterLock.Unlock()
//		if collectionTry, okTry := connectMasterPool[tableName]; okTry {
//			return collectionTry
//		}
//		collection = mongo.GetMgoMasterInstance().Collection(tableName)
//		connectMasterPool[tableName] = collection
//		return collection
//	}
//	return collection
//}
//
//// 批量插入数据
//func (m *CommonModel) Insert(ctx context.Context, commonObj CommonModelObj, documents []interface{}) ([]interface{}, error) {
//	if len(documents) == 0 {
//		return nil, nil
//	}
//	var insertIdList []interface{}
//	result, err := m.GetCollection(commonObj.GetTableName()).InsertMany(ctx, documents)
//	if err != nil {
//		return insertIdList, err
//	}
//	return result.InsertedIDs, nil
//}
//
//// 批量修改数据
//func (m *CommonModel) Update(ctx context.Context, commonObj CommonModelObj,
//	filter interface{}, update interface{}) (int64, error) {
//	updateResult, err := m.GetCollection(commonObj.GetTableName()).UpdateMany(ctx, filter, update)
//	if err != nil {
//		logrus.WithContext(ctx).Errorf("Update table:%s error:%v", commonObj.GetTableName(), err)
//		return 0, err
//	}
//	return updateResult.ModifiedCount, nil
//}
//
//// 批量删除数据
//func (m *CommonModel) Delete(ctx context.Context, commonObj CommonModelObj, filter interface{}) (int64, error) {
//	result, err := m.GetCollection(commonObj.GetTableName()).DeleteMany(ctx, filter)
//	if err != nil {
//		logrus.WithContext(ctx).Errorf("Delete table:%s error:%v", commonObj.GetTableName(), err)
//		return 0, err
//	}
//	return result.DeletedCount, nil
//}
//
//// 批量软删除数据
//func (m *CommonModel) SoftDelete(ctx context.Context, commonObj CommonModelObj, filter interface{}) (int64, error) {
//	values := bson.D{
//		{"commonmodel.status", StatusInvalid},
//		{"commonmodel.modify_time", time.Now()},
//	}
//	updateResult, err := m.GetCollection(commonObj.GetTableName()).UpdateMany(ctx, filter, bson.D{{"$set", values}})
//	if err != nil {
//		return 0, err
//	}
//	return updateResult.ModifiedCount, nil
//}
//
//// 自增
//func (m *CommonModel) Incr(ctx context.Context, commonObj CommonModelObj, filter interface{}, field string, incrBy int64) (
//	int64, error) {
//	update := bson.D{{"$inc", bson.M{field: incrBy}}}
//	// singleResult 是更新前的文档
//	singleResult := m.GetCollection(commonObj.GetTableName()).FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().
//		SetUpsert(true))
//	if err := singleResult.Err(); err != nil {
//		if err == MongoClient.ErrNoDocuments {
//			return incrBy, nil
//		} else {
//			return 0, err
//		}
//	}
//	var val int64
//	items := primitive.D{}
//	err := singleResult.Decode(&items)
//	if err != nil {
//		return 0, err
//	}
//	for _, item := range items {
//		if item.Key == field {
//			val = item.Value.(int64)
//			break
//		}
//	}
//	return val + incrBy, nil
//}

// 格式化
func (m *CommonModel) ToString() string {
	return fmt.Sprintf("create_time:%s modify_time:%s tenant_id:%s status:%d", m.CreateTime.String(),
		m.ModifyTime.String(), m.TenantId, m.Status)
}

func (m *CommonModel) CreateOrModify(ctx context.Context, projects []CommonModelObj) error {
	if len(projects) == 0 {
		return nil
	}
	models := make([]MongoClient.WriteModel, 0, len(projects))
	for _, project := range projects {
		filter := bson.M{"_id": project.GetId()}
		models = append(models, MongoClient.NewReplaceOneModel().SetFilter(filter).SetReplacement(project).SetUpsert(
			true))
	}
	resp, err := projects[0].GetCollection().BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).
			Errorf("[model.CreateOrModify.BulkWrite] BulkWrite error, resp:%v", resp)
	}
	return err
}

func (m *CommonModel) Modify(ctx context.Context, projects []CommonModelObj) error {
	if len(projects) == 0 {
		return nil
	}
	models := make([]MongoClient.WriteModel, 0, len(projects))
	for _, project := range projects {
		filter := bson.M{"_id": project.GetId()}
		models = append(models, MongoClient.NewReplaceOneModel().SetFilter(filter).SetReplacement(project).SetUpsert(
			false))
	}
	resp, err := projects[0].GetCollection().BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).
			Errorf("[model.Modify.BulkWrite] BulkWrite error,resp:%v", resp)
	}
	return err
}

func BatchUpdate(ctx context.Context, objects []CommonBatchUpdate) error {
	if len(objects) == 0 {
		return nil
	}
	models := make([]MongoClient.WriteModel, 0, len(objects))
	for _, object := range objects {
		if !object.IsVaildUpdate(ctx) {
			logrus.WithContext(ctx).Errorf("[model.BatchUpdate.BulkWrite] BulkWrite error filter or update is nil")
			return errors.New("BulkWrite error filter or update is nil")
		}
		object.InitModifyTime(ctx)
		filter := object.GetFilter(ctx)
		models = append(models, MongoClient.NewUpdateManyModel().SetFilter(filter).SetUpdate(object.GetUpdate(ctx)).SetUpsert(
			false))
	}
	resp, err := objects[0].GetCollection().BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).
			Errorf("[model.BatchUpdate.BulkWrite] BulkWrite error,resp:%v", resp)
	}
	return err
}

func (m *CommonModel) GetByFilterOne(ctx context.Context, commonObj CommonModelObj, filter CommonModelFilter) (
	interface{}, error) {
	flt := filter.ToMongoFilter()
	singleResult := commonObj.GetCollection().FindOne(ctx, flt, options.FindOne().SetSort(bson.D{{"create_time", 1}}))
	if err := singleResult.Err(); err != nil {
		if err == MongoClient.ErrNoDocuments {
			return nil, nil
		} else {
			return nil, err
		}
	}
	resp := commonObj.GetInstance()
	err := singleResult.Decode(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *CommonModel) GetByOption(ctx context.Context, commonObj CommonModelObj,
	filter CommonModelFilter, option CommonModelOption) (interface{}, error) {
	flt := filter.ToMongoFilter()
	opt := option.ToMongoOption()
	cursor, err := commonObj.GetCollection().Find(ctx, flt, opt...)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("[model.GetByFilterList] err:%v", err)
		return nil, err
	}
	if err = cursor.Err(); err != nil {
		if err == MongoClient.ErrNoDocuments {
			return nil, nil
		} else {
			return nil, err
		}
	}
	var records = commonObj.GetInstanceList()
	err = cursor.All(ctx, records)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("[model.GetByFilterList] err:%v", err)
		return nil, err
	}
	return records, nil
}

func (m *CommonModel) GetById(ctx context.Context, commonObj CommonModelObj, id string) (interface{}, error) {
	singleResult := commonObj.GetCollection().FindOne(ctx, bson.M{"_id": id})
	if err := singleResult.Err(); err != nil {
		if err == MongoClient.ErrNoDocuments {
			return nil, nil
		} else {
			logrus.WithContext(ctx).Errorf("GetById error: %v tablename:%s", err, commonObj.GetTableName())
			return nil, err
		}
	}
	resp := commonObj.GetInstance()
	err := singleResult.Decode(resp)
	if err != nil {
		logrus.WithContext(ctx).Errorf("GetById error: %v tablename:%s", err, commonObj.GetTableName())
		return nil, err
	}
	return resp, nil
}

//func (m *CommonModel) Invalid(ctx context.Context, commonObj CommonModelObj, filter CommonModelFilter) error {
//
//	flt := filter.ToMongoFilter()
//	update := bson.M{"$set": bson.M{"commonmodel.status": StatusInvalid}}
//	resp, err := m.Update(ctx, commonObj, flt, update)
//	if err != nil {
//		logrus.WithContext(ctx).WithError(err).Errorf("[model.Invalid] error resp:%v err:%v", resp, err)
//	}
//	return err
//}

func (m *CommonModel) GetByFilterList(ctx context.Context, commonObj CommonModelObj, filter CommonModelFilter) (
	interface{}, error) {
	flt := filter.ToMongoFilter()
	cursor, err := commonObj.GetCollection().Find(ctx, flt, options.Find().SetSort(bson.D{{"create_time", 1}}))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("[model.GetByFilterList] err:%v", err)
		return nil, err
	}
	var records = commonObj.GetInstanceList()
	err = cursor.All(ctx, records)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("[model.GetByFilterList] err:%v", err)
		return nil, err
	}
	return records, nil
}

func (m *CommonModel) Aggregate(ctx context.Context, commonObj CommonModelObj,
	filter []bson.D) (interface{}, error) {
	cursor, err := commonObj.GetCollection().Aggregate(ctx, filter)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("[model.Aggregate] err:%v", err)
		return nil,err
	}
	var records = commonObj.GetInstanceList()
	err = cursor.All(ctx, records)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("[model.Aggregate] err:%v", err)
		return nil, err
	}
	return records, nil
}

func (m *CommonModel) CountDocuments(ctx context.Context, commonObj CommonModelObj, filter CommonModelFilter) (int64, error) {
	num, err := commonObj.GetCollection().CountDocuments(ctx, filter.ToMongoFilter())
	if err != nil {
		logrus.WithContext(ctx).Errorf("CountDocuments error:%v table:%s", err, commonObj.GetTableName())
	}
	return num, err
}

func (m *CommonModel) IsInvalid() bool {
	if m == nil {
		return false
	}

	return m.Status == StatusInvalid
}

func (m *CommonModel) IsValidWrite() bool {
	if m == nil {
		return false
	}

	return m.Status == StatusValidWrite
}

func (m *CommonModel) IsValidRead() bool {
	if m == nil {
		return false
	}

	return m.Status == StatusValidRead
}

func (m *CommonModel) IsDraftAdd() bool {
	if m == nil {
		return false
	}

	return m.Status == StatusDraftAdd
}

func (m *CommonModel) IsDraftRemove() bool {
	if m == nil {
		return false
	}

	return m.Status == StatusDraftRemove
}

type BatchUpdateModel struct {
	Obj    CommonModelObj
	Filter bson.M
	Update bson.M
}

func (obj *BatchUpdateModel) GetFilter(ctx context.Context) bson.M {
	return obj.Filter
}
func (obj *BatchUpdateModel) GetUpdate(ctx context.Context) bson.M {
	return bson.M{"$set": obj.Update}
}

func (obj *BatchUpdateModel) GetCollection() *MongoClient.Collection {
	return obj.Obj.GetCollection()
}

func NewBatchUpdateModel(ctx context.Context, cobj CommonModelObj) CommonBatchUpdate {
	return &BatchUpdateModel{
		Obj:    cobj,
		Filter: bson.M{},
		Update: bson.M{},
	}
}

func (obj *BatchUpdateModel) SetFilter(key string, value interface{}) CommonBatchUpdate {
	obj.Filter[key] = value
	return obj
}
func (obj *BatchUpdateModel) SetUpdate(key string, value interface{}) CommonBatchUpdate {
	obj.Update[key] = value
	return obj
}

func (obj *BatchUpdateModel) IsVaildUpdate(ctx context.Context) bool {
	if len(obj.Update) == 0 || len(obj.Filter) == 0 {
		return false
	}
	return true
}

func (obj *BatchUpdateModel) InitModifyTime(ctx context.Context) CommonBatchUpdate {
	obj.Update["commonmodel.modify_time"] = time.Now()
	return obj
}
