package seeders

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"user-service/constants"
	"user-service/domain/models"
)

func RunUserSeeder(db *gorm.DB) {
	password, _ := bcrypt.GenerateFromPassword([]byte("admin1234"), bcrypt.DefaultCost)
	user := models.User{
		UUID:        uuid.New(),
		Name:        "Administrator",
		Username:    "admin",
		Password:    string(password),
		PhoneNumber: "0896123123",
		Email:       "admin@admin.com",
		RoleID:      constants.Admin,
	}

	err := db.FirstOrCreate(&user, models.User{Username: user.Username}).Error
	if err != nil {
		logrus.Errorf("failed to seed user: %v", err)
		panic(err)
	}
	logrus.Infof("role %s successfully seeded", user.Username)
}
