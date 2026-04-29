package data

import (
	"context"
	"errors"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	dataModel "github.com/hobbyGG/Dawnix/internal/workflow/data/model"
	domain "github.com/hobbyGG/Dawnix/internal/workflow/domain"
	//"gorm.io/gorm"
)

type RecordRepo struct {
	db *Data
}

// 调用gorm框架写出对应的历史记录数据并且返回
func NewRecordRepo(db *Data) biz.RecordRepo {
	return &RecordRepo{
		db: db,
	}
}

// Create 新建流程记录
func (repo *RecordRepo) Create(ctx context.Context, record *domain.Record) (int64, error) { //要返回审批记录的ID吗？不返回ID 有什么必要返回的数据与需要的功能对应?
	if record == nil {
		return -1, errors.New("repo: creation model cannot be nil")
	}

	recordPO := recordTOPO(record)
	// repo := recordTOPO(record) // domain dataModel
	if err := repo.db.DB(ctx).WithContext(ctx).Create(recordPO).Error; err != nil {
		return -1, err
	}
	record.ID = recordPO.ID
	return recordPO.ID, nil
	//return r.db.WithContext(ctx).Create(repo).Error
}

func (repo *RecordRepo) List(ctx context.Context, instanceID int64) ([]*domain.Record, error) {
	var records []dataModel.Record

	//query:= repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.Record{})

	err := repo.db.DB(ctx).WithContext(ctx). //这个写法好复杂看不懂
							Where("instance_id = ?", instanceID).
							Order("created_at desc").
							Find(&records).Error
	if err != nil {
		return nil, err
	}

	//将数据库模型 -> domain实体
	result := make([]*domain.Record, 0, len(records))
	for _, record := range records {
		result = append(result, record.ToDomain())
	}
	return result, nil
	// var pdList []dataModel.ProcessDefinition
	// query := repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.ProcessDefinition{})

	// res := query.Offset(params.Page - 1).Limit(params.Size).Find(&pdList)
	// if err := res.Error; err != nil {
	// 	return nil, err
	// }
	// result := make([]domain.ProcessDefinition, 0, len(pdList))
	// for i := range pdList {
	// 	if item := pdList[i].ToDomain(); item != nil {
	// 		result = append(result, *item)
	// 	}
	// }
	// return result, nil

}

func (repo *RecordRepo) ListAll(ctx context.Context) ([]*domain.Record, error) {
	var records []dataModel.Record

	err := repo.db.DB(ctx).WithContext(ctx).
		Order("created_at desc").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Record, 0, len(records))
	for _, record := range records {
		result = append(result, record.ToDomain())
	}
	return result, nil
}

// 工具方法：domain → dataModel
// -------------------------------------------------------
func recordTOPO(d *domain.Record) *dataModel.Record {
	if d == nil {
		return nil //如果传入的domain对象为nil 就返回nil 这样就不会创建一条空记录了
	}
	return &dataModel.Record{
		InstanceID:   d.InstanceID,
		TaskID:       d.TaskID,
		NodeID:       d.NodeID,
		NodeName:     d.NodeName,
		ApproverUID:  d.ApproverUID,
		ApproverName: d.ApproverName,
		Action:       d.Action,
		Comment:      d.Comment, //这个作用是什么？
		CreatedAt:    d.CreatedAt,
	}
}
