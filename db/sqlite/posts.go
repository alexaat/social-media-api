package sqlite

import (
	"database/sql"
	"fmt"
	types "my-social-network/types"

	"time"
)

func SavePost(post types.Post) (*int64, error) {

	statement, err := db.Prepare("INSERT INTO posts (date, user_id, content, privacy, image) VALUES(?,?,?,?,?)")
	if post.Group != nil {
		statement, err = db.Prepare("INSERT INTO posts (date, user_id, content, privacy, image, group_id) VALUES(?,?,?,?,?,?)")
	}

	if err != nil {
		return nil, err
	}

	defer statement.Close()

	date := time.Now().UnixNano() / 1000000

	var res sql.Result
	if post.Group == nil {
		res, err = statement.Exec(date, post.User.Id, post.Content, post.Privacy, post.Image)
	} else {
		res, err = statement.Exec(date, post.User.Id, post.Content, post.Privacy, post.Image, post.Group.(int))
	}

	if err != nil {
		return nil, err
	}

	lid, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &lid, nil
}

func GetAllPostsRespectPrivacy(userId int) (*[]types.Post, error) {
	posts := []types.Post{}

	postsByUserIdAndPublicPostsExcludeLoggedUserAndSpecificFriendsExcludeLoggedUserAndPrivateExcluseLoggedUser, err := GetPostsByUserIdAndPublicPostsExcludeLoggedUserAndSpecificFriendsExcludeLoggedUserAndPrivatePostsExcludeLoggedUser(userId)
	if err != nil {
		fmt.Println("err12 ", err)
		return nil, err
	}
	posts = append(posts, *postsByUserIdAndPublicPostsExcludeLoggedUserAndSpecificFriendsExcludeLoggedUserAndPrivateExcluseLoggedUser...)
	return &posts, nil
}

func GetPostsByUserIdNoPrivacy(userId int) (*[]types.Post, error) {
	fmt.Println("Own Profile")
	//Show all posts by logged in user
	postOfLoggedUser, err := GetPostsByUserId(userId)
	if err != nil {
		fmt.Println("err1 ", err)
		return nil, err
	}
	fmt.Println("Own Posts: ", postOfLoggedUser)
	return postOfLoggedUser, nil
}

func GetPostsByUserIdRespectPrivacy(personId int, currentUserId int) (*[]types.Post, error) {
	posts := []types.Post{}
	fmt.Println("Profile Of ", personId)

	publicPostsByUserIdAndPrivateSpecificPostsByUserIdAndPrivatePosts, err := GetPublicPostsByUserIdAndPrivateSpecificPostsByUserIdAndPrivatePosts(personId, currentUserId)
	if err != nil {
		fmt.Println("err21 ", err)
		return nil, err
	}
	posts = append(posts, *publicPostsByUserIdAndPrivateSpecificPostsByUserIdAndPrivatePosts...)

	return &posts, nil
}

func GetPostsByUserId(id int) (*[]types.Post, error) {
	posts := []types.Post{}

	sql := `
	SELECT posts.id, date, user_id, users.nick_name, users.first_name, users.last_name, users.avatar, users.privacy, content, posts.privacy, image
	FROM posts
	INNER JOIN users
	ON user_id = users.id
	WHERE user_id = ?
	ORDER BY date DESC`

	rows, err := db.Query(sql, id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&(post.Id), &(post.Date), &(post.User.Id), &(post.User.NickName), &(post.User.FirstName), &(post.User.LastName), &(post.User.Avatar), &(post.User.Privacy), &(post.Content), &(post.Privacy), &(post.Image))
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &posts, nil
}

func GetPostsByUserIdAndPublicPostsExcludeLoggedUserAndSpecificFriendsExcludeLoggedUserAndPrivatePostsExcludeLoggedUser(userId int) (*[]types.Post, error) {
	posts := []types.Post{}

	sql := `
	SELECT 
		posts.id,
		posts.date,
		user_id,
		users.nick_name,
		users.first_name,
		users.last_name,
		users.avatar,
		users.privacy,
		content,
		posts.privacy,
		image,
		group_id
	FROM
		posts
	JOIN
		users
	ON
		user_id = users.id
	WHERE
		user_id = ?
	OR	
		user_id = (SELECT creator_id FROM groups WHERE groups.id = group_id)
	OR
		user_id IN (SELECT members FROM groups WHERE  groups.id = group_id)	
	OR
		(posts.privacy = 'public' AND user_id != ?)
	OR
		posts.privacy LIKE '[' || ? || ']' 
	OR
		posts.privacy LIKE '[%,' || ? || ']' 
	OR
		posts.privacy LIKE '[' || ? || ',%]'
	OR
		posts.privacy LIKE '[%,' || ? || ',%]'
	OR
	(
		posts.privacy = 'private'
		AND
		user_id != ?
		AND
		user_id IN (SELECT users.id FROM followers INNER JOIN users ON users.id = followee WHERE approved = true AND follower = ?)
	) 
	ORDER BY
		date
	DESC`

	/*
		sql := `
		SELECT
			posts.id,
			posts.date,
			user_id,
			users.nick_name,
			users.first_name,
			users.last_name,
			users.avatar,
			users.privacy,
			content,
			posts.privacy,
			image,
			group_id
		FROM
			posts
		JOIN
			users
		ON
			user_id = users.id
		WHERE
			user_id = ?
		OR
			(posts.privacy = 'public' AND user_id != ?)
		OR
			posts.privacy LIKE '[' || ? || ']'
		OR
			posts.privacy LIKE '[%,' || ? || ']'
		OR
			posts.privacy LIKE '[' || ? || ',%]'
		OR
			posts.privacy LIKE '[%,' || ? || ',%]'
		OR
		(
			posts.privacy = 'private'
			AND
			user_id != ?
			AND
			user_id IN (SELECT users.id FROM followers INNER JOIN users ON users.id = followee WHERE approved = true AND follower = ?)
		)
		ORDER BY
			date
		DESC`
	*/

	rows, err := db.Query(sql, userId, userId, userId, userId, userId, userId, userId, userId)

	if err != nil {
		return nil, err
	}

	var groupId interface{}

	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(
			&(post.Id),
			&(post.Date),
			&(post.User.Id),
			&(post.User.NickName),
			&(post.User.FirstName),
			&(post.User.LastName),
			&(post.User.Avatar),
			&(post.User.Privacy),
			&(post.Content),
			&(post.Privacy),
			&(post.Image),
			&groupId)
		if err != nil {
			return nil, err
		}

		//Filter Group Posts
		if groupId == nil {
			post.Group = nil
			posts = append(posts, post)
		} else {

			group, err := GetGroupById(int(groupId.(int64)))
			if err != nil {
				return nil, err
			}
			if group.Creator.(types.UserBasicInfo).Id == userId {
				post.Group = group
				posts = append(posts, post)
			}
			if group.Members != nil {
				for _, m := range group.Members.([]types.UserBasicInfo) {
					if m.Id == userId {
						post.Group = group
						posts = append(posts, post)
						break
					}
				}
			}

		}

	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &posts, nil
}

func GetPublicPostsByUserIdAndPrivateSpecificPostsByUserIdAndPrivatePosts(personId int, currentUserId int) (*[]types.Post, error) {
	posts := []types.Post{}

	sql := `
	SELECT posts.id, date, user_id, users.nick_name, users.first_name, users.last_name, users.avatar, users.privacy, content, posts.privacy, image
	FROM posts
	INNER JOIN users
	ON user_id = users.id
	WHERE
	(posts.privacy = 'public' AND user_id = ?)
	OR
	posts.privacy LIKE '[' || ? || ']' 
	OR
	posts.privacy LIKE '[%,' || ? || ']' 
	OR
	posts.privacy LIKE '[' || ? || ',%]'
	OR
	posts.privacy LIKE '[%,' || ? || ',%]'
	OR
		(
			posts.privacy = 'private'
			AND
			user_id = ?
			AND
			user_id IN (SELECT users.id FROM followers INNER JOIN users ON users.id = followee WHERE approved = true AND follower = ?)
		)  
	ORDER BY date DESC
	`

	rows, err := db.Query(sql, personId, currentUserId, currentUserId, currentUserId, currentUserId, personId, currentUserId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&(post.Id), &(post.Date), &(post.User.Id), &(post.User.NickName), &(post.User.FirstName), &(post.User.LastName), &(post.User.Avatar), &(post.User.Privacy), &(post.Content), &(post.Privacy), &(post.Image))
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &posts, nil
}

func GetPostsByGroupId(groupId int) (*[]types.GroupPost, error) {
	posts := []types.GroupPost{}

	sql := `
	SELECT
	  posts.id,
	  group_id,
	  date,
	  user_id,
	  users.nick_name,
	  users.first_name,
	  users.last_name,
	  users.avatar,
	  content,
	  posts.image
	FROM posts
	INNER JOIN users
	ON user_id = users.id
	WHERE group_id = ?
	ORDER BY date DESC`

	rows, err := db.Query(sql, groupId)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for rows.Next() {
		post := types.GroupPost{}
		err = rows.Scan(&(post.Id), &(post.GroupId), &(post.Date), &(post.User.Id), &(post.User.NickName), &(post.User.FirstName), &(post.User.LastName), &(post.User.Avatar), &(post.Content), &(post.Image))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		posts = append(posts, post)
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &posts, nil
}
