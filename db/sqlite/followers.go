package sqlite

import (
	"errors"
	"fmt"
	types "my-social-network/types"
	"time"
)

func UpdateFollowers(follower int, followee int, privacy string) error {

	query := `
	SELECT id, approved 
	FROM followers
	WHERE
	follower = ? AND followee = ?
	LIMIT 1
	`
	row, err := db.Query(query, follower, followee)

	if err != nil {
		return err
	}

	defer row.Close()

	var id int
	var approved bool

	for row.Next() {
		err = row.Scan(&id, &approved)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	if id != 0 && privacy == "private" && !approved {
		return errors.New("awaiting approval")
	}

	approved = true
	if privacy == "private" {
		approved = false
	}

	//New Approval request
	if id == 0 {
		query = `INSERT INTO followers (date, follower ,followee, approved) VALUES(?, ?, ?, ?)`

		fmt.Println("query = ", query)

		statement, err := db.Prepare(query)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer statement.Close()
		date := time.Now().UnixNano() / 1000000

		_, err = statement.Exec(date, follower, followee, approved)

		if err != nil {
			fmt.Println(err)
			return err
		}

		return nil
	}

	//Existing request
	query = `UPDATE followers SET approved = ? WHERE id = ?`
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}

	defer statement.Close()
	_, err = statement.Exec(approved, id)

	if err != nil {
		return err
	}

	return nil
}

func DeleteFollower(follower int, followee int) error {
	statement, err := db.Prepare("DELETE FROM followers WHERE follower = ? AND followee = ?")

	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(follower, followee)

	if err != nil {
		return err
	}

	numOfRows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if numOfRows == 0 {
		return errors.New("no active follow requests found")
	}

	return nil
}

func GetFollowing(userId int) (*[]types.Following, error) {
	if userId < 1 {
		return nil, errors.New("invalid user id")
	}

	query := `
	SELECT
	users.id, users.first_name, users.last_name, users.nick_name, users.avatar, users.privacy, approved, date FROM followers
	INNER JOIN
	users
	ON
	users.id = followee
	WHERE
	follower = ?
	ORDER BY
	date
	DESC`

	rows, err := db.Query(query, userId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	followings := []types.Following{}

	for rows.Next() {
		user := types.User{}
		following := types.Following{}
		err = rows.Scan(&(user.Id), &(user.FirstName), &(user.LastName), &(user.NickName), &(user.Avatar), &(user.Privacy), &(following.Approved), &(following.Date))
		if err != nil {
			return nil, err
		}
		following.Following = &user
		followings = append(followings, following)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &followings, nil
}

func GetFollowers(userId int) (*[]types.Follower, error) {
	if userId < 1 {
		return nil, errors.New("invalid user id")
	}

	query := `
	SELECT
	users.id, users.first_name, users.last_name, users.nick_name, users.avatar, users.privacy, approved, date FROM followers
	INNER JOIN
	users
	ON
	users.id = follower
	WHERE
	followee = ?
	ORDER BY
	date
	DESC
	`

	rows, err := db.Query(query, userId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	followers := []types.Follower{}

	for rows.Next() {
		user := types.User{}
		follower := types.Follower{}
		err = rows.Scan(&(user.Id), &(user.FirstName), &(user.LastName), &(user.NickName), &(user.Avatar), &(user.Privacy), &(follower.Approved), &(follower.Date))
		if err != nil {
			return nil, err
		}
		follower.Follower = &user
		followers = append(followers, follower)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &followers, nil
}

func ApproveFollower(follower int, followee int) error {

	fmt.Println("Approve: ", follower, followee)
	sql := `
	UPDATE followers
	SET approved = true
	WHERE follower = ? AND followee = ?;	
	`

	statement, err := db.Prepare(sql)

	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(follower, followee)

	if err != nil {
		return err
	}

	numOfRows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if numOfRows == 0 {
		return errors.New("no active follow requests found")
	}

	return nil
}
