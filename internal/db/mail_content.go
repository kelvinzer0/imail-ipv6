package db

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/kelvinzer0/imail-ipv6/internal/conf"
	"github.com/kelvinzer0/imail-ipv6/internal/tools"
)

type MailContent struct {
	Id      int64  `gorm:"primaryKey"`
	Mid     int64  `gorm:"uniqueIndex;comment:MID"`
	Content string `gorm:"comment:Content"`
}

func (*MailContent) TableName() string {
	return TablePrefix("mail_content")
}

func MailContentDir(uid, mid int64) string {
	dataPath := path.Join(conf.Web.AppDataPath, "mail", "user"+strconv.FormatInt(uid, 10), string(strconv.FormatInt(mid, 10)[0]))
	return dataPath
}

func MailContentFilename(uid, mid int64) string {
	dataPath := MailContentDir(uid, mid)
	emailFile := fmt.Sprintf("%s/%d.eml", dataPath, mid)
	return emailFile
}

func MailContentRead(uid, mid int64) (string, error) {
	mode := conf.Web.MailSaveMode
	if strings.EqualFold(mode, "hard_disk") {
		return MailContentReadHardDisk(uid, mid)
	} else {
		return MailContentReadDb(mid)
	}
}

func MailContentReadDb(mid int64) (string, error) {
	var r MailContent
	err := db.Model(&MailContent{}).Where("mid = ?", mid).First(&r).Error
	if err != nil {
		return "", err
	}
	return r.Content, nil
}

func MailContentReadHardDisk(uid int64, mid int64) (string, error) {
	dataPath := path.Join(conf.Web.AppDataPath, "mail", "user"+strconv.FormatInt(uid, 10), string(strconv.FormatInt(mid, 10)[0]))

	if !tools.IsExist(dataPath) {
		err := os.MkdirAll(dataPath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	emailFile := fmt.Sprintf("%s/%d.eml", dataPath, mid)
	content, err := tools.ReadFile(emailFile)
	if err != nil {
		return "", err
	}
	return content, nil
}

func MailContentWrite(uid int64, mid int64, content string) error {
	mode := conf.Web.MailSaveMode
	if strings.EqualFold(mode, "hard_disk") {
		return MailContentWriteHardDisk(uid, mid, content)
	} else {
		return MailContentWriteDb(mid, content)
	}
}

func MailContentWriteDb(mid int64, content string) error {
	user := MailContent{Mid: mid, Content: content}
	result := db.Create(&user)
	return result.Error
}

func MailContentWriteHardDisk(uid int64, mid int64, content string) error {

	dataPath := MailContentDir(uid, mid)

	if !tools.IsExist(dataPath) {
		err := os.MkdirAll(dataPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	emailFile := fmt.Sprintf("%s/%d.eml", dataPath, mid)
	err := tools.WriteFile(emailFile, content)
	if err != nil {
		return err
	}
	return nil

}

func MailContentDelete(uid int64, mid int64) error {
	mode := conf.Web.MailSaveMode
	if strings.EqualFold(mode, "hard_disk") {
		return MailContentDeleteHardDisk(uid, mid)
	} else {
		return MailContentDeleteDb(mid)
	}
}

func MailContentDeleteDb(mid int64) error {
	err := db.Where("mid = ?", mid).Delete(&MailContent{}).Error
	return err
}

func MailContentDeleteHardDisk(uid int64, mid int64) error {
	dataPath := MailContentDir(uid, mid)

	emailFile := fmt.Sprintf("%s/%d.eml", dataPath, mid)
	if tools.IsExist(emailFile) {
		err := os.RemoveAll(emailFile)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
