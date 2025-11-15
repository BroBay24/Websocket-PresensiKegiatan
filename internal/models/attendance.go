package models

import "time"

type Attendance struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Nama     string    `gorm:"size:100;not null" json:"nama"`
	Nim      string    `gorm:"size:20;not null;uniqueIndex" json:"nim"`
	Jurusan  string    `gorm:"size:100;not null" json:"jurusan"`
	Angkatan int       `gorm:"not null" json:"angkatan"`
	Waktu    time.Time `gorm:"autoCreateTime" json:"waktu"`
}

func (Attendance) TableName() string {
	return "kehadiran"
}

func (a Attendance) ToMap() map[string]any {
	return map[string]any{
		"id":       a.ID,
		"nama":     a.Nama,
		"nim":      a.Nim,
		"jurusan":  a.Jurusan,
		"angkatan": a.Angkatan,
		"waktu":    a.Waktu,
	}
}
