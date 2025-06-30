package db

import (
	// "fmt"
	"time"

	"gorm.io/gorm"
)

type Domain struct {
	Id        int64     `gorm:"primaryKey"`
	Domain    string    `gorm:"unique;comment:顶级域名"`
	Mx        bool      `gorm:"comment:MX记录"`
	A         bool      `gorm:"comment:A记录"`
	AAAA      bool      `gorm:"comment:AAAA记录"`
	Spf       bool      `gorm:"comment:Spf记录"`
	Dkim      bool      `gorm:"comment:Dkim记录"`
	Dmarc     bool      `gorm:"comment:DMARC记录"`
	IsDefault bool      `gorm:"comment:是否默认"`
	Created   time.Time `gorm:"autoCreateTime;comment:创建时间"`
	Updated   time.Time `gorm:"autoCreateTime;comment:更新时间"`
}

func (*Domain) TableName() string {
	return TablePrefix("domain")
}

func DomainCreate(d *Domain) (err error) {
	data := db.First(d, "domain = ?", d.Domain)
	if data.Error == gorm.ErrRecordNotFound {
		result := db.Create(d)
		return result.Error
	}
	return data.Error
}

func DomainCount() int64 {
	var count int64
	db.Model(&Domain{}).Count(&count)
	return count
}

func DomainVaildList(page, pageSize int) ([]Domain, error) {
	domain := make([]Domain, 0, pageSize)
	dbm := db.Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc")
	err := dbm.Where("a=?", 1).
		Where("mx=?", 1).
		Where("spf=?", 1).
		Where("dkim=?", 1).
		Where("dmarc=?", 1).
		Find(&domain).Error
	return domain, err
}

func DomainVaild(name string) (bool, error) {
	var d Domain
	result := db.Model(&Domain{}).Where("domain=?", name).Where("a=?", 1).
		Where("mx=?", 1).
		Where("spf=?", 1).
		Where("dkim=?", 1).
		Where("dmarc=?", 1).
		Find(&d)

	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	if result.Error != nil {
		return false, result.Error
	}
	return d.Id > 0, nil
}

func DomainList(page, pageSize int) ([]Domain, error) {
	domain := make([]Domain, 0, pageSize)
	dbm := db.Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc")
	err := dbm.Find(&domain).Error
	return domain, err
}

func DomainDeleteByName(name string) error {
	var d Domain
	return db.Where("domain = ?", name).Delete(&d).Error
}

func DomainDeleteById(id int64) error {
	var d Domain
	return db.Where("id = ?", id).Delete(&d).Error
}

func DomainGetById(id int64) (Domain, error) {
	var d Domain
	err := db.First(&d, "id = ?", id).Error
	return d, err
}

func DomainUpdateById(id int64, d Domain) error {
	err := db.Where("id = ?", id).Save(d).Error
	return err
}

func DomainSetDefaultOnlyOne(id int64) error {
	result := db.Model(&Domain{}).Where("1 = ?", 1).Update("is_default", 0)
	if result.Error != nil {
		return result.Error
	}

	result = db.Model(&Domain{}).Where("id = ?", id).Update("is_default", 1)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DomainGetMain() (Domain, error) {
	var d Domain
	err := db.Model(&Domain{}).
		Where("a=?", 1).
		Where("mx=?", 1).
		Where("spf=?", 1).
		Where("dkim=?", 1).
		Where("dmarc=?", 1).
		Where("is_default=?", 1).
		First(&d).Error
	return d, err
}

func DomainGetMainForDomain() (string, error) {
	d, err := DomainGetMain()
	return d.Domain, err
}
