package recode

import (
	"github.com/cmazx/recode/pkg/convert"
	"gorm.io/gorm"
	"log"
	"strconv"
)

func AutoMigrateFormats(db *gorm.DB) error {
	return db.AutoMigrate(&MediaFormat{})
}

type MediaFormat struct {
	ID       uint                  `json:"id" gorm:"id;primary"`
	Name     string                `json:"name" gorm:"name"`
	Width    uint                  `json:"width" gorm:"width"`
	Height   uint                  `json:"height" gorm:"height"`
	Quality  uint                  `json:"quality" gorm:"quality"`
	Encoding convert.MediaEncoding `json:"encoding" gorm:"encoding"`
	Method   convert.ResizeMethod  `json:"method" gorm:"method"`
}

type MediaFormatStorage struct {
	db *gorm.DB
}

func NewMediaFormatStorage(conn *gorm.DB) *MediaFormatStorage {
	return &MediaFormatStorage{
		db: conn,
	}
}

func (s *MediaFormatStorage) create(format *MediaFormat) {
	s.db.Create(format)
}
func (s *MediaFormatStorage) update(format *MediaFormat) {
	s.db.Model(format).Updates(format)
}
func (s *MediaFormatStorage) find(id int) *MediaFormat {
	format := &MediaFormat{}
	s.db.First(format, id)
	return format
}
func (s *MediaFormatStorage) list() []MediaFormat {
	log.Println("List of media formats")
	var formats []MediaFormat
	s.db.Find(&formats)
	log.Println("Found " + strconv.Itoa(len(formats)))
	return formats
}
func (s *MediaFormatStorage) listByIds(ids []int32) []MediaFormat {
	var formats []MediaFormat
	s.db.Where("id IN ?", ids).Find(&formats)
	return formats
}
func (s *MediaFormatStorage) delete(id int) {
	s.db.Delete(&MediaFormat{}, id)
}
