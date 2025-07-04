package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/kelvinzer0/imail-ipv6/internal/tools/mail"
)

type Mail struct {
	Id                int64  `gorm:"primaryKey"`
	Uid               int64  `gorm:"index;comment:UserID"`
	Type              int    `gorm:"index;comment:0:Send;1:Received"`
	MailFrom          string `gorm:"size:50;comment:Mail From"`
	MailFromInContent string `gorm:"text;comment:Mail From Name In Content"`
	MailTo            string `gorm:"size:50;comment:Receive mail"`
	Subject           string `gorm:"size:191;comment:Subject"`
	SubjectIndex      string `gorm:"index;size:191;comment:Index Subject"`
	Size              int    `gorm:"size:10;comment:Message content size"`
	Status            int    `gorm:"size:2;comment:'0:准备发送;1:发送成功;2:发送失败;3:已接收'"`

	IsRead   bool `gorm:"index;default:0;comment:是否已读"`
	IsDelete bool `gorm:"index;default:0;comment:是否删除"`
	IsFlags  bool `gorm:"index;default:0;comment:是否星标"`
	IsJunk   bool `gorm:"index;default:0;comment:是否无用"`
	IsDraft  bool `gorm:"index;default:0;comment:是否草稿"`

	IsCheck bool `gorm:"default:0;comment:是否通过检查"`

	Created time.Time `gorm:"autoCreateTime;comment:创建时间"`
	Updated time.Time `gorm:"autoCreateTime;comment:更新时间"`
}

const (
	MailSearchOptionsTypeSend = iota
	MailSearchOptionsTypeInbox
	MailSearchOptionsTypeDraft
	MailSearchOptionsTypeDeleted
	MailSearchOptionsTypeFlags
	MailSearchOptionsTypeJunk
	MailSearchOptionsTypeUnread
)

type DIRType int32

const (
	DIR_DELETED DIRType = 0
	DIR_JUNK    DIRType = 1
	DIR_READ    DIRType = 2
	DIR_FLAGS   DIRType = 3
)

func (*Mail) TableName() string {
	return TablePrefix("mail")
}

func MailCount() int64 {
	var count int64
	db.Model(&Mail{}).Count(&count)
	return count
}

func MailCountWithOpts(opts *MailSearchOptions) (int64, error) {
	var count int64
	dbm := db.Model(&Mail{})
	dbm = MailSearchByNameCond(opts, dbm)
	err := dbm.Where("uid=?", opts.Uid).Count(&count).Error
	return count, err
}

func MailList(mtype int, page, pageSize int, keyword string) ([]Mail, error) {
	mails := make([]Mail, 0, pageSize)
	dbm := db.Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc")

	opts := &MailSearchOptions{
		Type:    mtype,
		Keyword: keyword,
	}
	dbm = MailSearchByNameCond(opts, dbm)

	err := dbm.Find(&mails).Error
	return mails, err
}

type MailSearchOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int
	Type     int
	Uid      int64
}

func MailSearchByNameCond(opts *MailSearchOptions, dbm *gorm.DB) *gorm.DB {
	if opts.Type == MailSearchOptionsTypeSend {
		dbm = dbm.Where("type = ?", 0).
			Where("is_junk = ?", 0).
			Where("is_delete = ?", 0).
			Where("is_flags = ?", 0).
			Where("is_draft = ?", 0)
	}

	if opts.Type == MailSearchOptionsTypeInbox {
		dbm = dbm.Where("type = ?", 1).
			Where("is_junk = ?", 0).
			Where("is_delete = ?", 0).
			Where("is_flags = ?", 0)
	}

	if opts.Type == MailSearchOptionsTypeDraft {
		dbm = dbm.Where("is_draft = ?", 1)
	}

	if opts.Type == MailSearchOptionsTypeDeleted {
		dbm = dbm.Where("is_delete = ?", 1).Where("is_draft = ?", 0)
	}

	if opts.Type == MailSearchOptionsTypeJunk {
		dbm = dbm.Where("is_junk = ?", 1).Where("is_draft = ?", 0)
	}

	if opts.Type == MailSearchOptionsTypeFlags {
		dbm = dbm.Where("is_flags = ?", 1).Where("is_draft = ?", 0)
	}

	return dbm
}

func MailSearchByName(opts *MailSearchOptions) ([]Mail, int64, error) {
	if len(opts.Keyword) == 0 {
		return nil, 0, nil
	}

	opts.Keyword = strings.ToLower(opts.Keyword)

	if opts.PageSize <= 0 || opts.PageSize > 20 {
		opts.PageSize = 10
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	searchQuery := "idx_%" + opts.Keyword + "%"
	email := make([]Mail, 0, opts.PageSize)

	dbm := db.Model(&Mail{}).Where("LOWER(subject_index) LIKE ?", searchQuery)
	dbm = MailSearchByNameCond(opts, dbm)
	err := dbm.Where("uid=?", opts.Uid).Find(&email).Error
	count, countErr := MailCountWithOpts(opts)
	if err != nil {
		return nil, 0, err
	}
	if countErr != nil {
		return nil, 0, countErr
	}
	return email, count, nil
}

func MailStatInfoForImap(uid int64, mtype int64) (int64, int64, error) {
	return MailStatInfo(uid, mtype)
}

func MailStatInfoForPop(uid int64) (int64, int64, error) {
	return MailStatInfo(uid, 0)
}

func MailStatInfo(uid int64, mtype int64) (int64, int64, error) {
	type Result struct {
		Count int64
		Size  int64
	}
	var result Result
	sql := fmt.Sprintf("SELECT count(uid) as count, sum(size) as size FROM `%s` WHERE uid=? and type=%d", TablePrefix("mail"), mtype)
	res := db.Raw(sql, uid).Scan(&result)

	if res.Error != nil {
		return 0, 0, res.Error
	}

	return result.Count, result.Size, nil
}

func MailListForPop(uid int64) ([]Mail, error) {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? and type=1 order by created_unix desc", TablePrefix("mail"))
	err := db.Raw(sql, uid).Find(&result).Error
	return result, err
}

func MailListForImap(uid int64) ([]Mail, error) {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? order by created_unix desc", TablePrefix("mail"))
	err := db.Raw(sql, uid).Find(&result).Error
	return result, err
}

func MailSendListForStatus(status int64, limit int64) ([]Mail, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE status=%d and type=0 and is_draft=0 order by created_unix limit %d", TablePrefix("mail"), status, limit)
	err := db.Raw(sql).Find(&result).Error
	return result, err
}

func MailListPosForPop(uid int64, pos int64) ([]Mail, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? and type=1 order by id limit %d,%d", TablePrefix("mail"), pos-1, 1)
	ret := db.Raw(sql, uid).Scan(&result)

	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailListForRspamd(limit int64) ([]Mail, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE type=1 and is_check=0 order by id desc limit %d", TablePrefix("mail"), limit)
	err := db.Raw(sql).Find(&result).Error
	return result, err
}

func MailListAllForPop(uid int64) ([]Mail, error) {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? and type=1 order by id limit 100", TablePrefix("mail"))
	ret := db.Raw(sql, uid).Scan(&result)
	// fmt.Println(sql, result)
	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailDeletedListAllForImap(uid int64) ([]Mail, error) {

	var result []Mail
	sql := fmt.Sprintf("SELECT id FROM `%s` WHERE uid=? and is_delete=1 order by id limit 10", TablePrefix("mail"))
	ret := db.Raw(sql, uid).Scan(&result)
	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailPosContentForPop(uid int64, pos int64) (string, int, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT id, uid, size FROM `%s` WHERE uid=? and type=1 order by id limit %d,%d", TablePrefix("mail"), pos-1, 1)
	ret := db.Raw(sql, uid).Scan(&result)

	if ret.Error != nil {
		return "", 0, ret.Error
	}

	content, err := MailContentRead(result[0].Uid, result[0].Id)
	if err != nil {
		return "", 0, err
	}

	return content, result[0].Size, nil
}

func MailDeleteById(id int64, status int64) error {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,uid FROM `%s` WHERE is_delete=1 and id='%d' order by id limit 1", TablePrefix("mail"), id)
	ret := db.Raw(sql).Scan(&result)
	if ret.Error == nil {
		if len(result) > 0 && status == 1 {
			return MailHardDeleteById(result[0].Uid, id)
		}
	}

	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_delete", status).Error
	return err
}

// 用过ID获取邮件的全部信息
func MailById(id int64) (Mail, error) {
	var m Mail
	result := db.Model(&Mail{}).Where("id=?", id).Take(&m)
	return m, result.Error
}

func MailSoftDeleteById(id int64) error {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_delete", 1).Error
	return err
}

func MailSoftDeleteByIds(ids []int64) error {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_delete", 1).Error
	return err
}

// TODO:批量硬删除邮件数据
func MailHardDeleteByIds(ids []int64) error {
	for _, id := range ids {

		mList, err := MailById(id)
		if err != nil {
			return err
		}

		if mList.IsDelete {
			if err := MailHardDeleteById(mList.Uid, id); err != nil {
				return err
			}
		}

	}
	return nil
}

func MailHardDeleteById(uid, mid int64) error {
	err := db.Where("id = ?", mid).Delete(&Mail{}).Error
	if err != nil {
		return err
	}
	return MailContentDelete(uid, mid)
}

func MailSeenById(id int64) error {
	ids := []int64{id}
	return MailSeenByIds(ids)
}

func MailSeenByIds(ids []int64) error {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_read", 1).Error
	return err
}

func MailUnSeenById(id int64) error {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_read", 0).Error
	return err
}

func MailUnSeenByIds(ids []int64) error {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_read", 0).Error
	return err
}

func MailSetFlagsById(id int64, status int64) error {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_flags", status).Error
	return err
}

func MailSetJunkById(id int64, status int64) error {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_junk", status).Error
	return err
}

func MailSetJunkByIds(ids []int64, status int64) error {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_junk", status).Error
	return err
}

func MailSetIsCheckById(id int64, status int64) error {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_check", status).Error
	return err
}

func MailSetStatusById(id int64, status int64) error {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("status", status).Error
	return err
}

func MailPushSend(uid int64, mail_from string, mail_to string, content string, is_draft bool) (int64, error) {
	return MailPush(uid, 0, mail_from, mail_to, content, 0, is_draft)
}

func MailPushReceive(uid int64, mail_from string, mail_to string, content string) (int64, error) {
	return MailPush(uid, 1, mail_from, mail_to, content, 3, false)
}

func MailUpdate(id int64, uid int64, mtype int, mail_from string, mail_to string, content string, status int, is_draft bool) (int64, error) {
	if id == 0 {
		return 0, errors.New("id is error!")
	}

	tx := db.Begin()

	subject := mail.GetMailSubject(content)
	subjectIndex := fmt.Sprintf("idx_%s", subject)
	mail_from_in_content := mail.GetMailFromInContent(content)

	m := Mail{
		Id:                id,
		Uid:               uid,
		Type:              mtype,
		MailFrom:          mail_from,
		MailFromInContent: mail_from_in_content,
		MailTo:            mail_to,
		Subject:           subject,
		SubjectIndex:      subjectIndex,
		Size:              len(content),
		Status:            status,
		IsDraft:           is_draft,
	}

	m.Updated = time.Now()
	m.Created = time.Now()
	result := tx.Save(&m)

	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	err := MailContentWrite(uid, m.Id, content)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return m.Id, nil
}

func MailPush(uid int64, mtype int, mail_from string, mail_to string, content string, status int, is_draft bool) (int64, error) {
	if uid == 0 {
		return 0, errors.New("user id is error!")
	}

	tx := db.Begin()

	subject := mail.GetMailSubject(content)
	subjectIndex := fmt.Sprintf("idx_%s", subject)
	mail_from_in_content := mail.GetMailFromInContent(content)

	m := Mail{
		Uid:               uid,
		Type:              mtype,
		MailFrom:          mail_from,
		MailFromInContent: mail_from_in_content,
		MailTo:            mail_to,
		Subject:           subject,
		SubjectIndex:      subjectIndex,
		Size:              len(content),
		Status:            status,
		IsDraft:           is_draft,
	}

	m.Updated = time.Now()
	m.Created = time.Now()
	result := tx.Create(&m)

	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	err := MailContentWrite(m.Uid, m.Id, content)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()
	return m.Id, nil
}
