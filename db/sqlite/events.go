package sqlite

import (
	"my-social-network/types"
)

func GetEvents(groupId int) (*[]types.Event, error) {
	events := []types.Event{}
	query := `
	SELECT
	 events.id,
	 users.nick_name,
	 users.first_name,
	 users.last_name,
	 users.avatar,
	 create_date,
	 event_date,
	 image,
	 title,
	 description,
	 members,
	 group_id
	 FROM
	 events
	 JOIN
	 users
	 ON
	 users.id = creator_id
	 WHERE
	 group_id = ?
	 ORDER BY
	 create_date
	 DESC
	`

	rows, err := db.Query(query, groupId)
	if err != nil {
		return nil, err
	}

	var nickName string
	var firstName string
	var lastName string

	for rows.Next() {
		event := types.Event{}
		basicUserInfo := types.UserBasicInfo{}

		err = rows.Scan(
			&(event.Id),
			&nickName,
			&firstName,
			&lastName,
			&(basicUserInfo.Avatar),
			&(event.CreateDate),
			&(event.EventDate),
			&(event.Image),
			&(event.Title),
			&(event.Description),
			&(event.Members),
			&(event.GroupId))

		if err != nil {
			return nil, err
		}

		displayName := nickName
		if nickName == "" {
			displayName = firstName + " " + lastName
		}
		basicUserInfo.DisplayName = displayName
		event.Creator = basicUserInfo
		events = append(events, event)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &events, nil
}

func GetEventById(eventId int) (*types.Event, error) {
	var event types.Event
	query := `
	SELECT
	 events.id,
	 users.nick_name,
	 users.first_name,
	 users.last_name,
	 users.avatar,
	 create_date,
	 event_date,
	 image,
	 title,
	 description,
	 members,
	 group_id
	 FROM
	 events
	 JOIN
	 users
	 ON
	 users.id = creator_id
	 WHERE
	 events.id = ?
	 ORDER BY
	 create_date
	 DESC
	 LIMIT 1	`

	rows, err := db.Query(query, eventId)
	if err != nil {
		return nil, err
	}

	var nickName string
	var firstName string
	var lastName string

	for rows.Next() {
		event = types.Event{}
		basicUserInfo := types.UserBasicInfo{}

		err = rows.Scan(
			&(event.Id),
			&nickName,
			&firstName,
			&lastName,
			&(basicUserInfo.Avatar),
			&(event.CreateDate),
			&(event.EventDate),
			&(event.Image),
			&(event.Title),
			&(event.Description),
			&(event.Members),
			&(event.GroupId))

		if err != nil {
			return nil, err
		}

		displayName := nickName
		if nickName == "" {
			displayName = firstName + " " + lastName
		}
		basicUserInfo.DisplayName = displayName
		event.Creator = basicUserInfo

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func SaveEvent(event types.Event) (*int64, error) {

	query := `
	INSERT INTO
	events
	(creator_id, create_date, event_date, image, title, description, members, group_id)
	VALUES
	(?,?,?,?,?,?,?,?)`

	statement, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	res, err := statement.Exec(
		event.Creator.(int),
		event.CreateDate,
		event.EventDate,
		event.Image,
		event.Title,
		event.Description,
		event.Members,
		event.GroupId)

	if err != nil {
		return nil, err
	}

	row, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func SaveEventMembers(eventId int, members string) (*int64, error) {
	query := `
	UPDATE
	events
	SET
	members = ?
	WHERE
	id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer statement.Close()

	result, err := statement.Exec(members, eventId)

	if err != nil {
		return nil, err
	}

	num, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	return &num, nil
}

func DeleteEventsByGroupId(groupId int) (*int64, error) {
	query := `
	DELETE FROM
	events
	WHERE
	group_id = ?	
	`

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer statement.Close()

	result, err := statement.Exec(groupId)

	if err != nil {
		return nil, err
	}

	num, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	return &num, nil
}
