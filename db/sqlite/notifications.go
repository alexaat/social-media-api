package sqlite

import (
	"fmt"

	types "my-social-network/types"
	"time"
)

func SaveNotification(n types.Notification) error {

	query := `
	INSERT INTO notifications
	(date, type, content, sender_id, recipient_id, is_read, group_id, event_id)
	VALUES(?,?,?,?,?,?,?,?)	
	`
	statement, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer statement.Close()

	date := time.Now().UnixNano() / 1000000

	senderId := -1
	recipientId := -1
	var groupId interface{}
	var eventId interface{}

	//Check Sender Type
	if _, ok := n.Sender.(types.User); ok {
		senderId = n.Sender.(types.User).Id
	}
	if _, ok := n.Sender.(*types.User); ok {
		senderId = n.Sender.(*types.User).Id
	}
	if _, ok := n.Sender.(types.UserBasicInfo); ok {
		senderId = n.Sender.(types.UserBasicInfo).Id
	}
	if _, ok := n.Sender.(int); ok {
		senderId = n.Sender.(int)
	}

	//Check Recipient Type
	if _, ok := n.Recipient.(types.User); ok {
		recipientId = n.Recipient.(types.User).Id
	}
	if _, ok := n.Recipient.(*types.User); ok {
		recipientId = n.Recipient.(*types.User).Id
	}
	if _, ok := n.Recipient.(types.UserBasicInfo); ok {
		recipientId = n.Recipient.(types.UserBasicInfo).Id
	}
	if _, ok := n.Recipient.(int); ok {
		recipientId = n.Recipient.(int)
	}

	//Check Group
	if _, ok := n.Group.(types.Group); ok {
		groupId = int(n.Group.(types.Group).Id)
	}
	if _, ok := n.Group.(int); ok {
		groupId = n.Group.(int)
	}

	//Check Event
	if _, ok := n.Event.(types.Event); ok {
		eventId = n.Event.(types.Event).Id
	}
	if _, ok := n.Event.(int); ok {
		eventId = n.Event.(int)
	}

	/*

		if _, ok := n.Sender.(*types.User); ok {
			senderType = "user"
			senderId = n.Sender.(*types.User).Id
		}

		if _, ok := n.Sender.(*types.UserBasicInfo); ok {
			senderType = "user"
			senderId = n.Sender.(*types.UserBasicInfo).Id
		}

		if _, ok := n.Sender.(*types.Group); ok {
			senderType = "group"
			senderId = n.Sender.(*types.Group).Id
		}

		if _, ok := n.Recipient.(*types.User); ok {
			recipientType = "user"
			recipientId = n.Recipient.(*types.User).Id
		}

		if _, ok := n.Recipient.(*types.Group); ok {
			recipientType = "group"
			recipientId = n.Recipient.(*types.Group).Id
		}
	*/

	_, err = statement.Exec(date, n.Type, n.Content, senderId, recipientId, false, groupId, eventId)

	if err != nil {
		return err
	}

	return nil
}

func GetNotifications(recipient interface{}) (*[]types.Notification, error) {

	notifications := []types.Notification{}

	recipientAsUser := recipient.(*types.User)

	sql := `
			SELECT
			notifications.id,
			notifications.date,
			notifications.type,
			notifications.content,
			notifications.sender_id,
			notifications.recipient_id,
			notifications.group_id,
			notifications.event_id,
			notifications.is_read
			FROM
			notifications
			WHERE
			recipient_id = ?
			ORDER BY date DESC
		`
	rows, err := db.Query(sql, recipientAsUser.Id)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for rows.Next() {
		notification := types.Notification{}
		err = rows.Scan(
			&(notification.Id),
			&(notification.Date),
			&(notification.Type),
			&(notification.Content),
			&(notification.Sender),
			&(notification.Recipient),
			&(notification.Group),
			&(notification.Event),
			&(notification.IsRead))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if notification.Group != nil {
			notification.Group = int(notification.Group.(int64))
		}
		if notification.Event != nil {
			notification.Event = int(notification.Event.(int64))
		}

		notifications = append(notifications, notification)
	}

	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &notifications, nil
}

func DeleteNotifications() error {
	statement, err := db.Prepare("DELETE FROM notifications")

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec()

	if err != nil {
		return err
	}

	return nil
}

func DropNotifications() error {
	statement, err := db.Prepare("DROP TABLE notifications")

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec()

	if err != nil {
		return err
	}

	return nil
}

func NotificationsSetRead(id string) error {
	statement, err := db.Prepare("UPDATE notifications SET is_read = true WHERE id = ?")

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(id)

	if err != nil {
		return err
	}
	return nil
}
