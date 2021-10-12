package recode

import (
	"encoding/json"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type JobStatus string

const (
	New     JobStatus = "new"
	Running JobStatus = "running"
	Done    JobStatus = "done"
	Failed  JobStatus = "failed"
)

type Job struct {
	ID         uint          `json:"id" gorm:"id;primaryKey"`
	Status     JobStatus     `json:"status" gorm:"status"`
	SourceUrl  string        `json:"sourceUrl" gorm:"source_url"`
	Formats    pq.Int32Array `json:"formats"  gorm:"formats;type:integer[]"`
	JobResults []JobResult   `json:"results" gorm:"foreignKey:JobId"`
	TargetPath string        `json:"targetPath" gorm:"target_path"`
	Publish    bool          `json:"publish" gorm:"publish"`
}

type JobData struct {
	Job     Job
	Formats []MediaFormat
}

func (w JobData) Serialize() []byte {
	bytes, err := json.Marshal(w)
	if err != nil {
		panic(err)
	}
	return bytes
}

type JobResult struct {
	ID       uint      `json:"id" gorm:"id;primaryKey"`
	JobId    uint      `json:"jobId" gorm:"job_id"`
	FormatId uint      `json:"formatId" gorm:"format_id"`
	Status   JobStatus `json:"status" gorm:"status"`
	Resource string    `json:"resource" gorm:"resource"`
	Details  string    `json:"details" gorm:"details"`
}

type JobStorage struct {
	db *gorm.DB
}

func AutomigrateJob(db *gorm.DB) error {
	return db.AutoMigrate(&Job{}, &JobResult{})
}

func NewJobFormatStorage(db *gorm.DB) *JobStorage {
	return &JobStorage{
		db: db,
	}
}

func (s *JobStorage) create(job *Job) {
	s.db.Create(job)
}
func (s *JobStorage) update(job *Job) {
	s.db.Model(job).Updates(job)
}

func (s *JobStorage) find(id int) *Job {
	job := &Job{}
	s.db.Preload("JobResults").First(job, id)
	return job
}
func (s *JobStorage) list(page int) []Job {
	var job []Job
	s.db.Preload("JobResults").Find(&job).Limit(100).Offset(page * 100)
	return job
}
func (s *JobStorage) delete(id uint) {
	s.db.Delete(&Job{}, id)
}

func (s *JobStorage) createJobResults(jobResults []*JobResult) {
	s.db.Create(jobResults)
}
