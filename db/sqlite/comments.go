package sqlite

import "my-social-network/types"

func SaveComment(comment types.Comment) (*int64, error) {
	statement, err := db.Prepare("INSERT INTO comments (date, user_id, post_id, content, image) VALUES (?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	res, err := statement.Exec(comment.Date, comment.User.(int), comment.PostId, comment.Content, comment.Image)
	if err != nil {
		return nil, err
	}

	row, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func GetComments(postId int) ([]types.Comment, error) {
	comments := []types.Comment{}
	sql :=
		`SELECT
			comments.id,
			comments.date,
			comments.user_id,
			users.nick_name,
			users.first_name,
			users.last_name,
			users.avatar,
			comments.post_id,
			comments.content,
			comments.image
		 FROM
		 	comments
		 JOIN
		 	users
		 ON
		 	comments.user_id = users.id
		 WHERE
		    comments.post_id = ?
		 ORDER BY
		 	date
		 DESC`

	rows, err := db.Query(sql, postId)
	if err != nil {
		return comments, err
	}

	var nickName string
	var firstName string
	var lastName string
	var image interface{}

	for rows.Next() {
		comment := types.Comment{}
		basicUserInfo := types.UserBasicInfo{}

		err = rows.Scan(
			&(comment.Id),
			&(comment.Date),
			&(basicUserInfo.Id),
			&nickName,
			&firstName,
			&lastName,
			&(basicUserInfo.Avatar),
			&(comment.PostId),
			&(comment.Content),
			&image)

		if err != nil {
			return nil, err
		}

		displayName := nickName
		if nickName == "" {
			displayName = firstName + " " + lastName
		}
		basicUserInfo.DisplayName = displayName
		comment.User = basicUserInfo

		if image != nil {
			comment.Image = image.(string)
		}

		comments = append(comments, comment)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return comments, nil
}
