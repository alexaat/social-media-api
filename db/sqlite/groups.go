package sqlite

import (
	"errors"
	"fmt"
	"my-social-network/types"
	util "my-social-network/util"
	"strconv"
	"strings"
)

func GetGroups() (*[]types.Group, error) {

	groups := []types.Group{}

	query := `
	SELECT
	id, creator_id, date, title, description
	FROM
	groups	
	ORDER BY
	date
	DESC`
	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		group := &types.Group{}
		err = rows.Scan(
			&(group.Id),
			&(group.Creator),
			&(group.Date),
			&(group.Title),
			&(group.Description))

		if err != nil {
			return nil, err
		}
		groups = append(groups, *group)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &groups, nil
}

func GetGroupById(id int) (*types.Group, error) {

	var group types.Group

	var creatorId int
	var nickName string
	var firstName string
	var lastName string
	var avatar string

	query := `
	SELECT
	groups.id, creator_id, nick_name, first_name, last_name, avatar, date, title, description, members
	FROM
	groups
	JOIN
	users
	ON
	creator_id = users.id
	WHERE groups.id = ?`

	rows, err := db.Query(query, id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(
			&(group.Id),
			&creatorId,
			&nickName,
			&firstName,
			&lastName,
			&avatar,
			&(group.Date),
			&(group.Title),
			&(group.Description),
			&(group.Members),
		)

		if err != nil {
			return nil, err
		}

		displayName := nickName
		if displayName == "" {
			displayName = firstName + " " + lastName
		}

		creator := types.UserBasicInfo{
			Id:          creatorId,
			DisplayName: displayName,
			Avatar:      avatar,
		}
		group.Creator = creator
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	//Get Members Info
	if group.Members != nil {

		str := fmt.Sprintf("%v", group.Members)

		stripped := strings.TrimSuffix(strings.TrimPrefix(str, "["), "]")
		arr := strings.Split(stripped, ",")
		arrAny := make([]interface{}, len(arr))
		for i, v := range arr {
			arrAny[i] = v
		}

		placeholders := ""

		for i := 0; i < len(arr); i++ {
			placeholders += "?,"
		}

		placeholders = strings.TrimSuffix(placeholders, ",")

		query = `
		SELECT
		id, nick_name, first_name, last_name, avatar	
		FROM
		users
		WHERE
		id IN (` + placeholders + ")"

		rows, err = db.Query(query, arrAny...)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		members := []types.UserBasicInfo{}

		for rows.Next() {

			member := types.UserBasicInfo{}
			err = rows.Scan(
				&(member.Id),
				&nickName,
				&firstName,
				&lastName,
				&(member.Avatar))

			if err != nil {
				return nil, err
			}

			displayName := nickName
			if displayName == "" {
				displayName = firstName + " " + lastName
			}
			member.DisplayName = displayName
			members = append(members, member)
		}
		group.Members = members
	} else {
		group.Members = []types.UserBasicInfo{}
	}

	return &group, nil
}

func SaveGroup(creatorId int, title string, description string) (*int64, error) {

	statement, err := db.Prepare("INSERT INTO groups (creator_id, date, title, description) VALUES(?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	date := util.GetCurrentMilli()

	res, err := statement.Exec(creatorId, date, title, description)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func AddToGroup(groupId int, member int) (*int64, error) {

	//1. Get Group Members
	query := `SELECT creator_id, members FROM groups WHERE id = ?`

	rows, err := db.Query(query, groupId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvedMembersInterface interface{}
	var creator_id int

	for rows.Next() {
		err = rows.Scan(&creator_id, &approvedMembersInterface)
		if err != nil {
			return nil, err
		}
	}

	fmt.Println("approvedMembersInterface", approvedMembersInterface)

	approvedMembersStr := ""
	if approvedMembersInterface != nil {
		approvedMembersStr = approvedMembersInterface.(string)
	}
	fmt.Println("approvedMembersStr", approvedMembersStr)

	ids := []int{}

	if approvedMembersStr != "" && approvedMembersStr != "[]" {
		approvedMembersStr = strings.TrimPrefix(strings.TrimSuffix(approvedMembersStr, "]"), "[")
		approvedMembersStrArr := strings.Split(approvedMembersStr, ",")
		for _, val := range approvedMembersStrArr {
			m_id, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			ids = append(ids, m_id)
		}
	}

	fmt.Println("ids ", ids)

	//2. Verify if members are already in the db
	if member == creator_id {
		e := errors.New("error. Cannot add creator as a group member")
		fmt.Println(e)
		return nil, e
	}

	for j := 0; j < len(ids); j++ {
		id := ids[j]
		if id == member {
			e := errors.New("error. Member already exists")
			fmt.Println(e)
			return nil, e
		}
	}

	//3. Insert new Members

	ids = append(ids, member)

	membersStr := ""
	for _, val := range ids {
		membersStr += strconv.Itoa(val) + ","
	}
	membersStr = strings.TrimSuffix(membersStr, ",")
	membersStr = "[" + membersStr + "]"

	query =
		`UPDATE 
	groups
	SET 
	members = ?
	WHERE
	id = ?
	`
	statement, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(membersStr, groupId)
	if err != nil {
		return nil, err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &num, nil
}

func DeleteGroup(groupId int) (*int64, error) {
	query := `
	DELETE FROM
	groups
	WHERE
	id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(groupId)
	if err != nil {
		return nil, err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	query = `	
	DELETE FROM 
	posts
	WHERE 
	group_id = ?`

	statement, err = db.Prepare(query)

	if err != nil {
		return nil, err
	}

	res, err = statement.Exec(groupId)
	if err != nil {
		return nil, err
	}

	num1, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	numTotal := num + num1

	return &numTotal, nil
}

func SaveInvitationToGroup(inviterId int, groupId int, memberId int, date int64) (*int64, error) {
	statement, err := db.Prepare("INSERT INTO group_invites (date, inviter_id, group_id, member_id) VALUES(?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(date, inviterId, groupId, memberId)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func SaveJoinGroupRequest(groupId int, memberId int, date int64) (*int64, error) {

	statement, err := db.Prepare("INSERT INTO join_group_requests (date, group_id, member_id) VALUES(?,?,?)")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(date, groupId, memberId)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func GetInvites(userId int) (*[]types.InviteToJoinGroup, error) {

	invites := []types.InviteToJoinGroup{}

	query := `
	SELECT
	date, group_id, users.id, users.nick_name, users.first_name, users.last_name, users.avatar
	FROM 
	group_invites
	JOIN
	users
	ON
	inviter_id = users.id
	WHERE
	member_id = ? 	
	`

	rows, err := db.Query(query, userId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupId int
	var nickName string
	var lastName string
	var firstName string

	for rows.Next() {
		invite := types.InviteToJoinGroup{}
		invite.Inviter = &types.UserBasicInfo{}
		err = rows.Scan(
			&invite.Date,
			&groupId,
			&invite.Inviter.Id,
			&nickName,
			&firstName,
			&lastName,
			&invite.Inviter.Avatar,
		)

		displayName := nickName
		if strings.TrimSpace(displayName) == "" {
			displayName = firstName + " " + lastName
		}

		invite.Inviter.DisplayName = displayName

		if err != nil {
			return nil, err
		}

		group, err := GetGroupById(groupId)
		if err != nil {
			return nil, err
		}

		invite.Group = *group

		invites = append(invites, invite)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &invites, nil
}

func SaveMembersToGroup(groupId int, members string) (*int64, error) {

	query :=
		`UPDATE 
	groups
	SET 
	members = ?
	WHERE
	id = ?
	`
	statement, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(members, groupId)
	if err != nil {
		return nil, err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &num, nil
}

func DeleteInvites(groupId int, memberId int) (*int64, error) {

	query := `
	DELETE FROM  
	group_invites
	WHERE 
	group_id = ?
	AND
	member_id = ?
	`
	statement, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(groupId, memberId)
	if err != nil {
		return nil, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &num, nil
}

func GetGroupsByCreatorId(creator_id int) (*[]types.Group, error) {
	groups := []types.Group{}

	var creatorId int
	var nickName string
	var firstName string
	var lastName string
	var avatar string

	query := `
	SELECT
	groups.id, creator_id, nick_name, first_name, last_name, avatar, date, title, description, members
	FROM
	groups
	JOIN
	users
	ON
	creator_id = users.id
	WHERE creator_id = ?`

	rows, err := db.Query(query, creator_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var group types.Group
		err = rows.Scan(
			&(group.Id),
			&creatorId,
			&nickName,
			&firstName,
			&lastName,
			&avatar,
			&(group.Date),
			&(group.Title),
			&(group.Description),
			&(group.Members),
		)

		if err != nil {
			return nil, err
		}

		displayName := nickName
		if displayName == "" {
			displayName = firstName + " " + lastName
		}

		creator := types.UserBasicInfo{
			Id:          creatorId,
			DisplayName: displayName,
			Avatar:      avatar,
		}
		group.Creator = creator

		groups = append(groups, group)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &groups, nil
}

func GetJoinRequestByGroupId(groupId int) (*[]types.JoinRequest, error) {

	joinRequests := []types.JoinRequest{}

	query := `
	SELECT
	id, date, group_id, member_id
	FROM
	join_group_requests
	WHERE 
	group_id = ?	
	ORDER BY
	date
	DESC`
	rows, err := db.Query(query, groupId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		joinRequest := types.JoinRequest{}
		err = rows.Scan(
			&(joinRequest.Id),
			&(joinRequest.Date),
			&(joinRequest.GroupId),
			&(joinRequest.MemberId))

		if err != nil {
			return nil, err
		}
		joinRequests = append(joinRequests, joinRequest)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &joinRequests, nil
}

func DeleteGroupJoinRequest(groupId int, memberId int) (*int64, error) {
	query := `
	DELETE FROM
	join_group_requests
	WHERE
	group_id = ?
	AND
	member_id = ?
	`
	statement, err := db.Prepare(query)

	if err != nil {
		return nil, err
	}
	defer statement.Close()

	res, err := statement.Exec(groupId, memberId)
	if err != nil {
		return nil, err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &num, nil
}

func GetJoinRequestByGroupIdAndMemberId(groupId int, memberId int) ([]types.JoinRequest, error) {
	joinRequests := []types.JoinRequest{}

	query := `
	SELECT
	id, date, group_id, member_id
	FROM
	join_group_requests	
	WHERE 
	group_id = ?
	AND
	member_id = ?
	ORDER BY
	date
	DESC`
	rows, err := db.Query(query, groupId, memberId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		joinRequest := types.JoinRequest{}
		err = rows.Scan(
			&(joinRequest.Id),
			&(joinRequest.Date),
			&(joinRequest.GroupId),
			&(joinRequest.MemberId))

		if err != nil {
			return nil, err
		}
		joinRequests = append(joinRequests, joinRequest)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return joinRequests, nil
}

func GetInvited(groupId int) ([]int, error) {
	query :=
		`
	SELECT member_id
	FROM group_invites
	WHERE
	group_id = ?		`

	rows, err := db.Query(query, groupId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id int

	invited := []int{}

	for rows.Next() {

		err = rows.Scan(&id)

		if err != nil {
			return nil, err
		}

		invited = append(invited, id)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return invited, nil
}
