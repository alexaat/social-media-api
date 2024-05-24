package main

import (
	"fmt"
	db "my-social-network/db/sqlite"
	types "my-social-network/types"
	"net/http"
)

func getUserFromRequest(r *http.Request) (*types.User, *types.Error) {
	keys, ok := r.URL.Query()["session_id"]
	if !ok || len(keys[0]) < 1 {
		return nil, &types.Error{Type: MISSING_PARAM, Message: "Error: missing request parameter: session_id"}
	}
	session_id := keys[0]

	user, err := db.GetUserBySessionId(session_id)

	if err != nil {
		return nil, &types.Error{Type: DATABASE_ERROR, Message: "Error: could not get user from database"}
	}

	if user == nil {
		return nil, &types.Error{Type: NO_USER_FOUND, Message: "Could not find user"}
	}

	return user, nil
}

func isAccessRestricted(userId int, personId int) *types.Error {

	person, err := db.GetUserById(personId)
	if err != nil {
		return &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get user from database: %v", personId)}
	}
	privacy := person.Privacy
	if privacy == "private" {
		followings, err := db.GetFollowing(userId)
		if err != nil {
			return &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get followings from database for user: %v", personId)}
		}
		isFollowing := false
		for i := 0; i < len(*followings); i++ {
			following := (*followings)[i]
			if personId == following.Following.Id && following.Approved {
				isFollowing = true
				break
			}
		}
		if !isFollowing {
			return &types.Error{Type: AUTHORIZATION, Message: "Private accont. Access is restricted"}
		}
	}
	return nil
}
