package main

import (
	"os"
	"strings"
	types "my-social-network/types"

	db "my-social-network/db/sqlite"
)

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}

func ToUserBasicInfo(id int) *types.UserBasicInfo {
	user, err := db.GetUserById(id)
	if err != nil || user == nil {
		return nil
	}
	displayName := strings.TrimSpace(user.NickName)
	if displayName == "" {
		displayName = user.FirstName + " " + user.LastName
	}

	userBasicInfo := types.UserBasicInfo{
		Id:          user.Id,
		DisplayName: displayName,
		Avatar:      user.Avatar,
	}
	return &userBasicInfo
}

