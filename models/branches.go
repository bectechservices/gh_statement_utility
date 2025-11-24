package models

import (
	"time"

	"gh-statement-app/pagination"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// Branch is used by pop to map your branches database table to your go code.
type Branch struct {
	ID         uuid.UUID `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"column:name"`
	Code       string    `json:"code" gorm:"column:code"`
	BankName   string    `json:"bank_name" gorm:"column:bank_name"`
	StreetName string    `json:"street_name" gorm:"column:street_name"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// Branches is not required by pop and may be deleted
type Branches []Branch

// LoadAllBranches loads all branches
func LoadAllBranches(tx *gorm.DB) Branches {
	branches := make(Branches, 0)
	tx.Order("code asc").Find(&branches)
	return branches
}

// CountOnBoardedBranches counts all branches
func CountOnBoardedBranches(tx *gorm.DB) int {
	var count int64
	tx.Model(&Branches{}).Count(&count)
	return int(count)
}

// GetBranchByCode gets a branch by code
func GetBranchByCode(code string, tx *gorm.DB) Branch {
	branch := Branch{}
	if err := tx.Where("code=?", code).First(&branch); err != nil {
		panic(err)
	}
	return branch
}

// PaginateBranches pagniates the records
func PaginateBranches(search string, tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	branches := make(Branches, 0)
	query := "%" + search + "%"
	tx.Scopes(Paginate(branches, &pagination, tx)).Where("name like ? or code like ?", query, query).Order("created_at desc").Find(&branches)
	pagination.Rows = branches
	return &pagination
}

// CreateBranch creates a new branch
func CreateBranch(name, code, bank_name, street_name string, tx *gorm.DB) Branch {
	branch := Branch{}
	branch.ID = NewUUID()
	branch.Name = name
	branch.Code = code
	branch.BankName = bank_name
	branch.StreetName = street_name
	tx.Create(&branch)
	return branch
}

// EditBranch edits an existing branch
func (b Branch) EditBranch(name, code, bank_name, street_name string, tx *gorm.DB) {
	b.Name = name
	b.Code = code
	b.BankName = bank_name
	b.StreetName = street_name
	tx.Save(&b)
}

// GetBranchByID gets a branch by ID
func GetBranchByID(id uuid.UUID, tx *gorm.DB) Branch {
	branch := Branch{}
	tx.Where("id=?", id).First(&branch)
	return branch
}

// GetBranchByID gets a branch by name
func GetBranchByName(name string, tx *gorm.DB) Branch {
	branch := Branch{}
	tx.Where("name=?", name).First(&branch)
	return branch
}

// GetBranchByUserID gets a branch by ID
func GetBranchByUserID(id uuid.UUID, tx *gorm.DB) Branch {
	branch := Branch{}

	tx.Where("id=?", id).First(&branch)
	return branch
}
