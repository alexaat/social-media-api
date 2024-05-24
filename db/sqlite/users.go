package sqlite

import (
	types "my-social-network/types"
	util "my-social-network/util"
	"strings"
)

func SaveUser(user *types.User) (int64, error) {

	statement, err := db.Prepare("INSERT INTO users (first_name, last_name, date_of_birth, nick_name, email, password, about_me, avatar, privacy) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return -1, err
	}
	defer statement.Close()

	result, err := statement.Exec(
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		user.NickName,
		strings.TrimSpace(strings.ToLower(user.Email)),
		util.Encrypt(user.Password),
		user.AboutMe,
		user.Avatar,
		strings.TrimSpace(strings.ToLower(user.Privacy)))

	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func GetUserBySessionId(session_id string) (*types.User, error) {

	if strings.TrimSpace(session_id) == "" {
		return nil, nil
	}

	query := `
	SELECT * FROM users
	WHERE
	id = 
		(SELECT user_id FROM session WHERE session_id = ? LIMIT 1)
	LIMIT 1
	`
	rows, err := db.Query(query, session_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var user *types.User = nil
	for rows.Next() {
		user = &types.User{}
		err = rows.Scan(
			&(user.Id),
			&(user.FirstName),
			&(user.LastName),
			&(user.DateOfBirth),
			&(user.NickName),
			&(user.Email),
			&(user.Password),
			&(user.AboutMe),
			&(user.Avatar),
			&(user.Privacy))

		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByEmailOrNickNameAndPassword(user types.User) (*types.User, error) {
	u := types.User{}

	// Get By Email
	rows, err := db.Query("SELECT * FROM users WHERE email = ?", strings.ToLower(strings.TrimSpace(user.NickName)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&(u.Id),
			&(u.FirstName),
			&(u.LastName),
			&(u.DateOfBirth),
			&(u.NickName),
			&(u.Email),
			&(u.Password),
			&(u.AboutMe),
			&(u.Avatar),
			&(u.Privacy))

		if err != nil {
			return nil, err
		}
		if util.CompairPasswords(u.Password, user.Password) {
			return &u, nil
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	// Get By Nick Name
	rows, err = db.Query("SELECT * FROM users WHERE nick_name = ?", strings.TrimSpace(user.NickName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&(u.Id),
			&(u.FirstName),
			&(u.LastName),
			&(u.DateOfBirth),
			&(u.NickName),
			&(u.Email),
			&(u.Password),
			&(u.AboutMe),
			&(u.Avatar),
			&(u.Privacy))

		if err != nil {
			return nil, err
		}
		if util.CompairPasswords(u.Password, user.Password) {
			return &u, nil
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func UpdateUser(user *types.User) error {
	query := `
	UPDATE users
	SET
	first_name = ?,
	last_name = ?,
	date_of_birth = ?,
	nick_name = ?,
	email = ?,
	password = ?,
	about_me = ?,
	avatar = ?,
	privacy = ?
	WHERE id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		user.NickName,
		user.Email,
		user.Password,
		user.AboutMe,
		user.Avatar,
		user.Privacy,
		user.Id)

	if err != nil {
		return err
	}

	return nil
}

func GetUserById(id int) (*types.User, error) {

	query := `
	SELECT * FROM users
	WHERE
	id = ?		
	LIMIT 1
	`
	rows, err := db.Query(query, id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var user *types.User = nil
	for rows.Next() {
		user = &types.User{}
		err = rows.Scan(
			&(user.Id),
			&(user.FirstName),
			&(user.LastName),
			&(user.DateOfBirth),
			&(user.NickName),
			&(user.Email),
			&(user.Password),
			&(user.AboutMe),
			&(user.Avatar),
			&(user.Privacy))

		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUsers(currentUserId int) (*[]interface{}, error) {
	users := []interface{}{}

	query := `
	SELECT id, nick_name, first_name, last_name, avatar FROM users`

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nickName string
	var firstName string
	var lastName string

	for rows.Next() {
		user := types.UserBasicInfo{}
		err := rows.Scan(
			&(user.Id),
			&nickName,
			&firstName,
			&lastName,
			&(user.Avatar),
		)

		if err != nil {
			return nil, err
		}

		if nickName != "" {
			user.DisplayName = nickName
		} else {
			user.DisplayName = firstName + " " + lastName
		}

		users = append(users, user)

	}

	return &users, nil
}
