package sqlite

import (
	"my-social-network/types"
)

func GetPrivateMessages(id1 int, id2 int) (*[]types.ChatMessage, error) {

	messages := []types.ChatMessage{}

	sql := `
	SELECT
	messages.id, sender_id, recipient_id, date, content, is_read
	FROM
	messages
	WHERE
	sender_id = ? AND recipient_id = ?
	UNION
	SELECT
	messages.id, sender_id, recipient_id, date, content, is_read
	FROM
	messages
	WHERE
	sender_id = ? AND recipient_id = ?
	ORDER BY
	date
	DESC
	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer statement.Close()

	rows, err := db.Query(sql, id1, id2, id2, id1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		message := types.ChatMessage{}
		err = rows.Scan(
			&message.Id,
			&message.Sender,
			&message.Recipient,
			&message.Date,
			&message.Content,
			&message.IsRead)
		if err != nil {
			return nil, err
		}
		message.Sender = int(message.Sender.(int64))
		message.Recipient = int(message.Recipient.(int64))
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &messages, nil
}

func GetPrivateMessagesByUserId(userId int) (*[]types.ChatMessage, error) {
	messages := []types.ChatMessage{}

	sql := `
	SELECT
	messages.id, sender_id, recipient_id, date, content, is_read
	FROM
	messages
	WHERE	
	(sender_id = ? OR recipient_id = ?)
	AND
	chat_group_id IS NULL
	ORDER BY
	date
	DESC
	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer statement.Close()

	rows, err := db.Query(sql, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		message := types.ChatMessage{}
		err = rows.Scan(
			&message.Id,
			&message.Sender,
			&message.Recipient,
			&message.Date,
			&message.Content,
			&message.IsRead)
		if err != nil {
			return nil, err
		}
		message.Sender = int(message.Sender.(int64))
		message.Recipient = int(message.Recipient.(int64))
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &messages, nil
}

func SavePrivateMessage(senderId int, recipientId int, date int64, content string) (*int, error) {
	statement, err := db.Prepare("INSERT INTO messages (sender_id, recipient_id, date, content, is_read) VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(senderId, recipientId, date, content, false)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	idInt := int(id)

	return &idInt, nil
}

func SaveChatGroup(date int64, title string, image string, members string) (*int, error) {
	query := "INSERT INTO chat_groups (date, title, image, members) VALUES(?,?,?,?)"

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(date, title, image, members)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	idInt := int(id)

	return &idInt, nil
}

func GetChatGroups() (*[]types.ChatGroup, error) {
	chatGroups := []types.ChatGroup{}

	sql := `
	SELECT id, date, title, image, members
	FROM
	chat_groups
	ORDER BY
	date
	DESC
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		chatRoom := types.ChatGroup{}

		err = rows.Scan(
			&chatRoom.Id,
			&chatRoom.Date,
			&chatRoom.Title,
			&chatRoom.Image,
			&chatRoom.Members)

		if err != nil {
			return nil, err
		}
		chatGroups = append(chatGroups, chatRoom)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &chatGroups, nil
}

func GetChatGroupById(chatGroupId int) (*types.ChatGroup, error) {
	chatRoom := types.ChatGroup{}

	sql := `
	SELECT id, date, title, image, members
	FROM
	chat_groups
	WHERE
	id = ?
	ORDER BY
	date
	DESC
	LIMIT 1
	`

	rows, err := db.Query(sql, chatGroupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(
			&chatRoom.Id,
			&chatRoom.Date,
			&chatRoom.Title,
			&chatRoom.Image,
			&chatRoom.Members)

		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &chatRoom, nil
}

func GetChatGroupMessages(chatGroupId int) (*[]types.ChatMessage, error) {
	messages := []types.ChatMessage{}

	sql := `
	SELECT id, sender_id, chat_group_id, date, content, is_read, read_by
	FROM
	messages
	WHERE
	chat_group_id = ?
	ORDER BY
	date
	DESC
	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer statement.Close()

	rows, err := db.Query(sql, chatGroupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		message := types.ChatMessage{}
		err = rows.Scan(
			&message.Id,
			&message.Sender,
			&message.ChatGroup,
			&message.Date,
			&message.Content,
			&message.IsRead,
			&message.ReadBy)

		if err != nil {
			return nil, err
		}
		message.Sender = int(message.Sender.(int64))
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &messages, nil
}

func GetChatGroupMessagesByMemberId(userId int) (*[]types.ChatMessage, error) {
	messages := []types.ChatMessage{}

	sql := `
	SELECT messages.id, sender_id, chat_group_id, messages.date, content, is_read, read_by
	FROM
	messages
	JOIN
	chat_groups
	ON
	chat_group_id = chat_groups.id	
	WHERE
	members	LIKE '[' || ? || ']'
	OR
	members	LIKE '[%' || ? || ']'
	OR
	members	LIKE '[' || ? || '%]'
	OR
	members	LIKE '[%' || ? || '%]'
	ORDER BY
	messages.date
	DESC
	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer statement.Close()

	rows, err := db.Query(sql, userId, userId, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		message := types.ChatMessage{}
		err = rows.Scan(
			&message.Id,
			&message.Sender,
			&message.ChatGroup,
			&message.Date,
			&message.Content,
			&message.IsRead,
			&message.ReadBy)

		if err != nil {
			return nil, err
		}
		message.Sender = int(message.Sender.(int64))
		message.ChatGroup = int(message.ChatGroup.(int64))
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &messages, nil
}

func SaveChatGroupMessage(senderId int, chatGroupId int, date int64, content string) (*int, error) {
	statement, err := db.Prepare("INSERT INTO messages (sender_id, chat_group_id, date, content, is_read) VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(senderId, chatGroupId, date, content, false)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	idInt := int(id)
	return &idInt, nil
}

func SaveChatGroupMembers(chatGroupId int, members string) (*int, error) {
	query := `
	UPDATE
	chat_groups
	SET
	members = ?
	WHERE
	id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer statement.Close()

	result, err := statement.Exec(members, chatGroupId)

	if err != nil {
		return nil, err
	}

	num, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	numInt := int(num)

	return &numInt, nil
}

func UpdateIsRead(value bool, userId int, personId int) (*int, error) {
	query := `
	UPDATE
	messages
	SET
	is_read = ?
	WHERE
	is_read <> ?
	AND
	sender_id = ?
	AND
	recipient_id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer statement.Close()

	result, err := statement.Exec(value, value, personId, userId)

	if err != nil {
		return nil, err
	}

	num, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	numInt := int(num)

	return &numInt, nil
}

func UpdateReadBy(messageId int, readBys string) (*int, error) {
	query := `
	UPDATE
	messages
	SET
	read_by = ?
	WHERE
	id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer statement.Close()

	result, err := statement.Exec(readBys, messageId)

	if err != nil {
		return nil, err
	}

	num, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	numInt := int(num)

	return &numInt, nil
}

/*
func GetChatMessages(userId int, chatMateId int) (*[]types.ChatMessage, error) {

	messages := []types.ChatMessage{}

	sql := `
	SELECT messages.id, sender_id, nick_name, first_name, last_name, recipient_id, date, content, is_read
	FROM
	messages
	JOIN
	users
	ON
	sender_id = users.id
	WHERE sender_id = ? AND recipient_id = ?
	AND
	is_room = false
	UNION
	SELECT messages.id, sender_id, nick_name, first_name, last_name, recipient_id, date, content, is_read
	FROM
	messages
	JOIN
	users
	ON
	sender_id = users.id
	WHERE sender_id = ? AND recipient_id = ?
	AND
	is_room = false
	ORDER BY date DESC
	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer statement.Close()

	rows, err := db.Query(sql, userId, chatMateId, chatMateId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nickName := ""
	firstName := ""
	lastName := ""

	for rows.Next() {
		message := types.ChatMessage{}
		err = rows.Scan(
			&message.Id,
			&message.Sender,
			&nickName,
			&firstName,
			&lastName,
			&message.Recipient,
			&message.Date,
			&message.Content,
			&message.IsRead)

		if nickName == "" {
			message.SenderDisplayName = firstName + " " + lastName
		} else {
			message.SenderDisplayName = nickName
		}

		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &messages, nil
}

func GetChatMessagesByUserId(userId int) (*[]types.ChatMessage, error) {
	messages := []types.ChatMessage{}

	sql := `
	SELECT messages.id, sender_id, nick_name, first_name, last_name, recipient_id, date, content, is_read
	FROM
	messages
	JOIN
	users
	ON
	sender_id = users.id
	WHERE sender_id = ? OR recipient_id = ?
	AND
	is_room = false
	ORDER BY date DESC
	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer statement.Close()

	rows, err := db.Query(sql, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nickName := ""
	firstName := ""
	lastName := ""

	for rows.Next() {
		message := types.ChatMessage{}
		err = rows.Scan(
			&message.Id,
			&message.Sender,
			&nickName,
			&firstName,
			&lastName,
			&message.Recipient,
			&message.Date,
			&message.Content,
			&message.IsRead)

		if nickName == "" {
			message.SenderDisplayName = firstName + " " + lastName
		} else {
			message.SenderDisplayName = nickName
		}

		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &messages, nil
}

func SaveChatMessage(senderId int, recipientId int, content string) (*types.ChatMessage, error) {

	statement, err := db.Prepare("INSERT INTO messages (sender_id, recipient_id, date, content, is_read, is_room) VALUES(?,?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	date := util.GetCurrentMilli()

	res, err := statement.Exec(senderId, recipientId, date, content, false, false)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	message := types.ChatMessage{
		Id:        int(id),
		Sender:    senderId,
		Recipient: recipientId,
		Date:      date,
		Content:   content,
		IsRead:    false,
	}

	return &message, nil
}

func UpdateIsRead(value bool, userId int, personId int) (int64, error) {

	query := `
	UPDATE
	messages
	SET
	is_read = ?
	WHERE
	is_read <> ?
	AND
	sender_id = ?
	AND
	recipient_id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return -1, err
	}

	defer statement.Close()

	result, err := statement.Exec(value, value, personId, userId)

	if err != nil {
		return -1, err
	}

	num, err := result.RowsAffected()

	if err != nil {
		return -1, err
	}

	return num, nil
}

func SaveChatRoom(date int64, title string, image string, members string) (*int, error) {
	query:="INSERT INTO chat_rooms (date, title, image, members) VALUES(?,?,?,?)"

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(date, title, image, members)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	idInt := int(id)

	return &idInt, nil
}

func GetChatRooms() (*[]types.ChatRoom, error) {
	chatRooms := []types.ChatRoom{}

	sql := `
	SELECT id, date, title, image, members
	FROM
	chat_rooms
	ORDER BY
	date
	DESC
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		chatRoom := types.ChatRoom{}

		err = rows.Scan(
			&chatRoom.Id,
			&chatRoom.Date,
			&chatRoom.Title,
			&chatRoom.Image,
			&chatRoom.Members)

		if err != nil {
			return nil, err
		}
		chatRooms = append(chatRooms, chatRoom)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &chatRooms, nil
}

func SaveChatMessageIntoRoom(date int64, content string, senderId int, chatRoomId int) (*int, error) {
	statement, err := db.Prepare("INSERT INTO messages (sender_id, recipient_id, date, content, is_read, is_room) VALUES(?,?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(senderId, chatRoomId, date, content, false, true)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	idInt := int(id)
	return &idInt, nil
}

func GetChatMessagesByChatRoom(chatRoomId int) (*[]types.ChatMessage, error) {
	messages := []types.ChatMessage{}

	sql := `
	SELECT id, sender_id, recipient_id, date, content, is_read
	FROM
	messages
	WHERE
	is_room = true
	AND
	recipient_id = ?
	ORDER BY
	date
	DESC
	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer statement.Close()

	rows, err := db.Query(sql, chatRoomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		message := types.ChatMessage{}
		err = rows.Scan(
			&message.Id,
			&message.Sender,
			&message.Recipient,
			&message.Date,
			&message.Content,
			&message.IsRead)

		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &messages, nil
}

func GetChatRoomById(id int) (*types.ChatRoom, error) {
	chatRoom := types.ChatRoom{}

	sql := `
	SELECT id, date, title, image, members
	FROM
	chat_rooms
	WHERE
	id = ?
	ORDER BY
	date
	DESC
	LIMIT 1
	`

	rows, err := db.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(
			&chatRoom.Id,
			&chatRoom.Date,
			&chatRoom.Title,
			&chatRoom.Image,
			&chatRoom.Members)

		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &chatRoom, nil
}

func SaveChatRoomMembers(id int, members string) (*int, error) {
	query := `
	UPDATE
	chat_rooms
	SET
	members = ?
	WHERE
	id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer statement.Close()

	result, err := statement.Exec(members, id)

	if err != nil {
		return nil, err
	}

	num, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	numInt := int(num)

	return &numInt, nil
}
*/
