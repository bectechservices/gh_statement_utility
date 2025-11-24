package models

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// PasswordExpiry is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type UserStampDetails struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	Position          string    `json:"position" db:"position"`
	BranchStamp       bool      `json:"branch_stamp" db:"branch_stamp"`
	BranchID          uuid.UUID `json:"branch_id" db:"branch_id"`
	SignatureLocation string    `json:"signature_location" db:"signature_location"`
	User              User      //`gorm:"foreignKey:UserID;references:user_id"`
	Branch            Branch    //`gorm:"foreignKey:BranchID;references:ID"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// PasswordExpiries is not required by pop and may be deleted
type UserStampDetail []UserStampDetails

func GetStampifyUserID(ID uuid.UUID, tx *gorm.DB) UserStampDetails {
	stampify := UserStampDetails{}
	go UserBranchIDUpdate(ID, tx)
	tx.Preload("User").Preload("Branch").Where("user_id=?", ID).First(&stampify)
	fmt.Println("############ Branch User ##################",
		stampify.User.FirstName,
		stampify.User.Branch.BankName,
		stampify.User.Branch.StreetName)
	return stampify
}

func (usl UserStampDetails) UpdateStampifyUserDetails(position string, branch_stamp bool, signature_location string, tx *gorm.DB) {

	// fmt.Println("############ <<<position >>>>################", position)
	usl.Position = position
	usl.BranchStamp = branch_stamp
	// fmt.Println("#############<<<<USER ID >>>>###############", usl.UserID)
	usl.BranchID = usl.User.BranchID
	usl.SignatureLocation = SaveImage(usl.UserID.String(), signature_location) //saveImageAndPath(signature_location)
	if err := tx.Save(&usl); err != nil {
		//panic(err)
	}
}

func SaveImage(ID, data string) string {
	filename := fmt.Sprintf("%s_%s.%s", "signature", strings.ReplaceAll(ID, "-", ""), "jpg")
	fmt.Println("############# filename ###############", filename)
	fmt.Println("############# filename ###############", data)
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic(err)
	}

	imageFileLocation := filepath.Join(envy.Get("BYTE_PATH", ""), filename)
	f, err := os.Create(imageFileLocation)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(decoded); err != nil {
		panic(err)
	}
	if err := f.Sync(); err != nil {
		panic(err)
	}
	return imageFileLocation
}

// Update user branch details
func UserBranchIDUpdate(ID uuid.UUID, tx *gorm.DB) UserStampDetail {
	stampuser := UserStampDetail{}
	tmpbranch := UserBranchID(tx, ID)
	fmt.Println("##### Branch ID Test #######", tmpbranch, ID)
	tx.Raw("update user_stamp_details set branch_id = ? where user_id = ?", tmpbranch, ID).Scan(&stampuser)
	return stampuser
}
