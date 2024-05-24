package sqlite

func SaveSession(userId int, sessionId string) error {

	query := `
	INSERT INTO session(user_id, session_id)
  	VALUES(?, ?)
  	ON CONFLICT(user_id) DO UPDATE SET
	session_id = ?   
  	WHERE user_id = ?`

	statement, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(userId, sessionId, sessionId, userId)

	if err != nil {
		return err
	}

	/*
		query := `UPDATE session SET session_id = ? WHERE user_id = ?`

		statement, err := db.Prepare(query)

		if err != nil {
			return err
		}
		defer statement.Close()

		result, err := statement.Exec(sessionId, userId)

		if err != nil {
			return err
		}

		id, err := result.LastInsertId()

		if err != nil {
			return err
		}

		fmt.Println("Update")
		fmt.Println(id)

		if id == 0 {
			query = `INSERT INTO session (user_id, session_id) VALUES(?, ?)`

			statement, err = db.Prepare(query)

			if err != nil {
				return err
			}

			defer statement.Close()

			result, err = statement.Exec(userId, sessionId)

			if err != nil {
				return err
			}

			id, err := result.LastInsertId()

			if err != nil {
				return err
			}

			if id != 1 {
				return fmt.Errorf("session_id inset incorrect. Insert number: %v", id)
			}

			fmt.Println("Inset")
			fmt.Println(id)

		}
	*/

	return nil
}

func DeleteSession(sessionId string) error {
	query := `
		UPDATE session SET session_id = '' WHERE session_id = ?
	`

	statement, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(sessionId)

	if err != nil {
		return err
	}
	return nil
}
