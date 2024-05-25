package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http"

	types "my-social-network/types"

	db "my-social-network/db/sqlite"

	util "my-social-network/util"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to social media api!!!")
}

func signinHandler(w http.ResponseWriter, r *http.Request) {

	resp := types.Response{Payload: nil, Error: nil}

	if r.Method != "POST" {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	} else {

		user := strings.TrimSpace(r.FormValue("user"))
		password := strings.TrimSpace(r.FormValue("password"))

		if len(user) < 2 || len(user) > 50 {
			resp.Error = &types.Error{Type: INVALID_USER_FORMAT, Message: "Error: username should be between 2 and 50 characters long"}
		} else {

			user, err := db.GetUserByEmailOrNickNameAndPassword(types.User{NickName: user, Password: password})

			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: cannot get user from database"}
			} else if user == nil {
				resp.Error = &types.Error{Type: NO_USER_FOUND, Message: "Error: no user found"}
			} else {

				sessionId := generateSessionId()

				err = db.SaveSession(user.Id, sessionId)

				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: cannot save session"}
				} else {
					resp.Payload = types.Session{SessionId: sessionId}
				}
			}
		}
	}
	sendResponse(w, resp)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {

	resp := types.Response{Payload: nil, Error: nil}

	if r.Method != "POST" {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	} else {

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while parse multipart form %v", err)}
			sendResponse(w, resp)
			return
		}

		firstName := strings.TrimSpace(r.FormValue("firstName"))
		lastName := strings.TrimSpace(r.FormValue("lastName"))
		nickName := strings.TrimSpace(r.FormValue("nickName"))
		dateOfBirth := strings.TrimSpace(r.FormValue("dateOfBirth"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := strings.TrimSpace(r.FormValue("password"))
		about := strings.TrimSpace(r.FormValue("about"))

		//Validate image

		fileName := ""

		file, _, err := r.FormFile("image")

		if err == nil {

			defer file.Close()

			err = makeDirectoryIfNotExists(IMAGES_DIRECTORY)

			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating directory %v", err)}
				sendResponse(w, resp)
				return
			}

			uuid := generateSessionId()
			tempFile, err := os.CreateTemp(IMAGES_DIRECTORY, fmt.Sprintf("%v-*.gif", uuid))

			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating temp file %v", err)}
				sendResponse(w, resp)
				return
			}

			defer tempFile.Close()

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while reading file %v", err)}
				sendResponse(w, resp)
				return
			}

			_, err = tempFile.Write(fileBytes)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while writing file %v", err)}
				sendResponse(w, resp)
				return
			}

			fileName = filepath.Base(tempFile.Name())

		} else {
			if !strings.Contains(err.Error(), "no such file") {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: image error %v", err)}
				sendResponse(w, resp)
				return
			}
		}

		//Validate input
		if len(firstName) < 1 || len(firstName) > 50 {
			resp.Error = &types.Error{Type: INVALID_FIRST_NAME_FORMAT, Message: "Error: First Name should be between 1 and 50 characters long"}
			sendResponse(w, resp)
			return
		}
		if len(lastName) < 1 || len(lastName) > 50 {
			resp.Error = &types.Error{Type: INVALID_LAST_NAME_FORMAT, Message: "Error: Last Name should be between 1 and 50 characters long"}
			sendResponse(w, resp)
			return
		}
		if len(nickName) != 0 && (len(nickName) < 2 || len(nickName) > 50) {
			resp.Error = &types.Error{Type: INVALID_NICK_NAME_FORMAT, Message: "Error: Nickname should be between 2 and 50 characters long"}
			sendResponse(w, resp)
			return
		}

		parseTime, err := time.Parse("2006-01-02 15:04:05", dateOfBirth+" 00:00:00")
		if err != nil {
			resp.Error = &types.Error{Type: INVALID_DATE_FORMAT, Message: "Error: Invalid date format"}
			sendResponse(w, resp)
			return
		}
		milli := parseTime.UnixNano() / 1000000
		unixMilli := time.Now().UnixMilli()

		if milli > unixMilli {
			resp.Error = &types.Error{Type: INVALID_DATE_FORMAT, Message: "Error: Invalid date"}
			sendResponse(w, resp)
			return
		}

		reg := `^[^@\s]+@[^@\s]+\.[^@\s]+$`
		match, err := regexp.MatchString(reg, email)
		if err != nil || !match {
			resp.Error = &types.Error{Type: INVALID_EMAIL, Message: "Error: invalid email"}
			sendResponse(w, resp)
			return
		}

		if len(password) < 6 || len(password) > 50 {
			resp.Error = &types.Error{Type: INVALID_PASSWORD, Message: "Error: password should be between 6 and 50 charachters long"}
			sendResponse(w, resp)
			return
		}

		if len(about) > 5000 {
			resp.Error = &types.Error{Type: INVALID_ABOUT_ME, Message: "Error: about me should be less than 5000 charachters"}
			sendResponse(w, resp)
			return
		}

		id, err := db.SaveUser(

			&types.User{
				FirstName:   firstName,
				LastName:    lastName,
				NickName:    nickName,
				DateOfBirth: milli,
				Email:       email,
				Password:    password,
				AboutMe:     about,
				Avatar:      fileName,
				Privacy:     "public",
			},
		)

		if err != nil {

			errorStr := fmt.Sprintf("%v", err)

			if strings.Contains(errorStr, "UNIQUE constraint") {
				if strings.Contains(errorStr, "email") {
					resp.Error = &types.Error{Type: INVALID_EMAIL, Message: "Error: email already in use"}
				} else {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot save user. %v", err)}
				}
			} else {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot save user. %v", err)}
			}
			sendResponse(w, resp)
			return
		}

		//Generate session id
		sessionId := generateSessionId()

		//Save session id to db
		err = db.SaveSession(int(id), sessionId)

		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: cannot save session"}
		}

		resp.Payload = types.Session{SessionId: sessionId}

	}
	sendResponse(w, resp)
}

func signoutHandler(w http.ResponseWriter, r *http.Request) {

	resp := types.Response{Payload: nil, Error: nil}

	if r.Method != "GET" {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
		sendResponse(w, resp)
		return
	}

	keys, ok := r.URL.Query()["session_id"]
	if !ok || len(keys[0]) < 1 {
		resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing request parameter: session_id"}
		sendResponse(w, resp)
		return
	}
	session_id := keys[0]

	user, err := db.GetUserBySessionId(session_id)

	if err != nil || user == nil {
		resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: could not get user from database"}
		sendResponse(w, resp)
		return
	}
	removeClient(user.Id)

	user.Password = ""
	user.DateOfBirth = 0
	resp.Payload = user
	resp.Payload = user.Id
	sendResponse(w, resp)
}

func sendResponse(w http.ResponseWriter, resp types.Response) {
	w.Header().Set("Access-Control-Allow-Origin", clientOrigin)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	json.NewEncoder(w).Encode(resp)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {

	resp := types.Response{Payload: nil, Error: nil}

	if r.Method == "GET" {
		keys, ok := r.URL.Query()["id"]
		if !ok || len(keys[0]) < 1 {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing request image parameter: id"}
			json.NewEncoder(w).Encode(resp)
			return
		}
		id := keys[0]
		path := "images/" + id
		w.Header().Set("Access-Control-Allow-Origin", clientOrigin)
		http.ServeFile(w, r, path)

	} else {
		sendResponse(w, resp)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	if r.Method == "GET" {
		user, err := getUserFromRequest(r)
		if err != nil {
			resp.Error = err
		} else {

			person_id_str := ""
			if strings.Contains(r.URL.Path, "/user/") {
				person_id_str = strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/user/"))
			}

			if person_id_str != "" {

				person_id, err := strconv.Atoi(person_id_str)
				if err != nil {
					resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", person_id_str)}
					sendResponse(w, resp)
					return
				}
				person, err := db.GetUserById(person_id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: could not get user from database"}
					sendResponse(w, resp)
					return
				}

				//Check Privacy
				e := isAccessRestricted(user.Id, person_id)
				if e == nil {
					resp.Payload = person
				} else {
					if e.Type == AUTHORIZATION {
						userRestricted := types.User{}
						userRestricted.Id = person.Id
						userRestricted.Avatar = person.Avatar
						if person.NickName != "" {
							userRestricted.NickName = person.NickName
						} else {
							userRestricted.FirstName = person.FirstName
							userRestricted.LastName = person.LastName
						}
						resp.Payload = userRestricted
					} else {
						resp.Error = e
					}
				}

			} else {
				resp.Payload = user
			}

		}

	} else if r.Method == "PUT" {

		user, e := getUserFromRequest(r)

		if e != nil {
			resp.Error = e
			sendResponse(w, resp)
			return
		}

		privacy := strings.TrimSpace(r.FormValue("privacy"))
		firstName := strings.TrimSpace(r.FormValue("first_name"))
		lastName := strings.TrimSpace(r.FormValue("last_name"))

		if privacy != "" {
			user.Privacy = privacy
		}
		if firstName != "" {
			user.FirstName = firstName
		}
		if lastName != "" {
			user.LastName = lastName
		}

		err := db.UpdateUser(user)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: could not update user in database"}
		}

		resp.Payload = user

	} else if r.Method == "PATCH" {
		user, e := getUserFromRequest(r)

		if e != nil {
			resp.Error = e
			sendResponse(w, resp)
			return
		}

		m := make(map[string]string)
		err := r.ParseForm()
		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse form: %v", err)}
			sendResponse(w, resp)
			return
		}

		for key, values := range r.PostForm {

			val := values[0]

			//Validation
			if key == "nick_name" {
				if len(val) != 0 && (len(val) < 2 || len(val) > 50) {
					resp.Error = &types.Error{Type: INVALID_NICK_NAME_FORMAT, Message: "Error: Nickname should be between 2 and 50 characters long"}
					sendResponse(w, resp)
					return
				}
			}

			if key == "first_name" {
				if len(val) < 1 || len(val) > 50 {
					resp.Error = &types.Error{Type: INVALID_FIRST_NAME_FORMAT, Message: "Error: First Name should be between 1 and 50 characters long"}
					sendResponse(w, resp)
					return
				}
			}
			if key == "last_name" {
				if len(val) < 1 || len(val) > 50 {
					resp.Error = &types.Error{Type: INVALID_LAST_NAME_FORMAT, Message: "Error: Last Name should be between 1 and 50 characters long"}
					sendResponse(w, resp)
					return
				}
			}

			if key == "email" {
				reg := `^[^@\s]+@[^@\s]+\.[^@\s]+$`
				match, err := regexp.MatchString(reg, val)
				if err != nil || !match {
					resp.Error = &types.Error{Type: INVALID_EMAIL, Message: "Error: invalid email"}
					sendResponse(w, resp)
					return
				}
			}

			if key == "about_me" {
				if len(val) > 5000 {
					resp.Error = &types.Error{Type: INVALID_ABOUT_ME, Message: "Error: about me should be less than 5000 charachters"}
					sendResponse(w, resp)
					return
				}
			}

			m[key] = val
		}

		num, err := db.UpdateUserDetails(user.Id, m)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not update user in database. %v", err)}
			sendResponse(w, resp)
			return
		}

		resp.Payload = types.Updated{Updated: int(*num)}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}

	sendResponse(w, resp)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}

	if r.Method == "GET" {

		users, err := db.GetUsers(user.Id)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get users from database. %v", err)}
			sendResponse(w, resp)
			return
		}

		resp.Payload = users

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}
	sendResponse(w, resp)
}

func postsHandler(w http.ResponseWriter, r *http.Request) {

	resp := types.Response{Payload: nil, Error: nil}

	//Check valid methods
	if r.Method != "GET" && r.Method != "POST" {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
		sendResponse(w, resp)
		return
	}

	user, e := getUserFromRequest(r)

	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}

	//New Post
	if r.Method == "POST" {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while parse multipart form %v", err)}
			sendResponse(w, resp)
			return
		}

		//Validate image
		fileName := ""

		file, _, err := r.FormFile("image")

		if err == nil {

			defer file.Close()

			err = makeDirectoryIfNotExists(IMAGES_DIRECTORY)

			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating directory %v", err)}
				sendResponse(w, resp)
				return
			}

			uuid := generateSessionId()

			tempFile, err := ioutil.TempFile(IMAGES_DIRECTORY, fmt.Sprintf("%v-*", uuid))

			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating temp file %v", err)}
				sendResponse(w, resp)
				return
			}

			defer tempFile.Close()

			fileBytes, err := ioutil.ReadAll(file)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while reading file %v", err)}
				sendResponse(w, resp)
				return
			}

			_, err = tempFile.Write(fileBytes)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while writing file %v", err)}
				sendResponse(w, resp)
				return
			}

			fileName = filepath.Base(tempFile.Name())

		} else {
			if !strings.Contains(err.Error(), "no such file") {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: image error %v", err)}
				sendResponse(w, resp)
				return
			}
		}

		content := strings.TrimSpace(r.FormValue("content"))
		privacy := strings.TrimSpace(r.FormValue("privacy"))
		groupIdStr := strings.TrimSpace(r.FormValue("group_id"))

		userId := user.Id

		post := types.Post{
			Content: content,
			Privacy: privacy,
			User:    types.User{Id: userId},
			Image:   fileName,
			//SpecificFriends: specificFriends,
		}

		if groupIdStr != "" {
			groupId, err := strconv.Atoi(groupIdStr)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", groupIdStr)}
				sendResponse(w, resp)
				return
			}
			// Check that user is a creator or a group member
			group, err := db.GetGroupById(groupId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			isPostAllowed := false
			if group.Creator.(types.UserBasicInfo).Id == user.Id {
				isPostAllowed = true
			} else if group.Members != nil {
				for _, m := range group.Members.([]types.UserBasicInfo) {
					if m.Id == userId {
						isPostAllowed = true
						break
					}
				}
			}
			if !isPostAllowed {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: could not get save post. Must to be a creator or a group member %v", err)}
				sendResponse(w, resp)
				return
			}
			post.Group = groupId
			post.Privacy = "public"
		}

		id, err := db.SavePost(post)

		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save post to database: %v", err)}
		} else {
			resp.Payload = types.PostId{PostId: int(*id)}
		}
	}

	//Get Posts
	if r.Method == "GET" {

		//Extract person_id from url
		person_id_str := ""
		params, ok := r.URL.Query()["person_id"]
		if ok && len(params[0]) > 0 {
			person_id_str = params[0]
		}

		//Extract group_id from url
		group_id_str := ""
		params, ok = r.URL.Query()["group_id"]
		if ok && len(params[0]) > 0 {
			group_id_str = params[0]
		}

		//url contains group_id
		if group_id_str != "" {
			//Convert id to int
			id, err := strconv.Atoi(group_id_str)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", group_id_str)}
				sendResponse(w, resp)
				return
			}

			//Check if user is a creator or a group member
			group, err := db.GetGroupById(id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			isGroupMember := false
			if group.Creator.(types.UserBasicInfo).Id == user.Id {
				isGroupMember = true
			} else if group.Members != nil {
				for _, m := range group.Members.([]types.UserBasicInfo) {
					if m.Id == user.Id {
						isGroupMember = true
						break
					}
				}
			}
			if !isGroupMember {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: could not get posts from database. Must be the group creator or a group member"}
				sendResponse(w, resp)
				return
			}

			posts, err := db.GetPostsByGroupId(id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get posts from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Get comments
			for index, post := range *posts {
				comments, err := db.GetComments(post.Id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get comments from database. %v", err)}
					sendResponse(w, resp)
					return
				}
				(*posts)[index].Comments = comments
			}

			resp.Payload = posts

		} else if person_id_str != "" {
			//url contains person_id
			//Convert id to int
			id, err := strconv.Atoi(person_id_str)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", person_id_str)}
				sendResponse(w, resp)
				return
			}

			if id == user.Id {
				//User Own Posts in Profile
				posts, err := db.GetPostsByUserIdNoPrivacy(id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get posts from database. %v", err)}
					sendResponse(w, resp)
					return
				}

				//Get Comments
				for index, post := range *posts {
					comments, err := db.GetComments(post.Id)
					if err != nil {
						resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get comments from database. %v", err)}
						sendResponse(w, resp)
						return
					}
					(*posts)[index].Comments = comments
				}

				resp.Payload = posts
			} else {
				//Someone else profile

				person, err := db.GetUserById(id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get person from database. %v", err)}
					sendResponse(w, resp)
					return
				}

				//Check Profile Privacy
				followings, err := db.GetFollowing(user.Id)
				isFollowing := false
				for _, f := range *followings {
					if f.Following.Id == id {
						isFollowing = true
						break
					}
				}

				if person.Privacy == "private" && !isFollowing {
					resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: no access - private account. %v", err)}
					sendResponse(w, resp)
					return
				}

				posts, err := db.GetPostsByUserIdRespectPrivacy(id, user.Id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get posts from database. %v", err)}
					sendResponse(w, resp)
					return
				}

				//Get Comments
				for index, post := range *posts {
					comments, err := db.GetComments(post.Id)
					if err != nil {
						resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get comments from database. %v", err)}
						sendResponse(w, resp)
						return
					}
					(*posts)[index].Comments = comments
				}
				resp.Payload = posts
			}

		} else {
			//No id in url. get posts for homepage
			posts, err := db.GetAllPostsRespectPrivacy(user.Id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get posts from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Add groups
			for index, post := range *posts {
				if post.Group != nil {
					group, err := db.GetGroupById(post.Group.(*types.Group).Id)
					if err != nil {
						resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
						sendResponse(w, resp)
						return
					}
					groupBasicInfo := types.GroupBasicInfo{
						Id:          group.Id,
						Title:       group.Title,
						Description: group.Description,
						Image:       group.Image}

					(*posts)[index].Group = groupBasicInfo
				}
			}

			//Get Comments
			for index, post := range *posts {
				comments, err := db.GetComments(post.Id)
				if err != nil {

					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get comments from database. %v", err)}
					sendResponse(w, resp)
					return
				}
				(*posts)[index].Comments = comments
			}

			resp.Payload = posts
		}
	}

	sendResponse(w, resp)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	resp := types.Response{Payload: nil, Error: nil}

	if r.Method != "GET" {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
		sendResponse(w, resp)
	} else {

		user, err := getUserFromRequest(r)
		if err != nil {
			resp.Error = err
			sendResponse(w, resp)
			return
		}

		addClient(*user, w, r)
	}
}

func followingHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	if r.Method != "GET" && r.Method != "POST" && r.Method != "OPTIONS" && r.Method != "DELETE" {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
		sendResponse(w, resp)
		return
	}

	user, err := getUserFromRequest(r)
	if err != nil {
		resp.Error = err
		sendResponse(w, resp)
		return
	}

	if r.Method == "GET" {

		person_id_str := ""
		params, ok := r.URL.Query()["person_id"]
		if ok && len(params[0]) > 0 {
			person_id_str = params[0]
		}

		if person_id_str != "" {

			person_id, err := strconv.Atoi(person_id_str)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", person_id_str)}
				sendResponse(w, resp)
				return
			}

			//Check Privacy
			e := isAccessRestricted(user.Id, person_id)
			if e != nil {
				fmt.Println("e: ", e)
				resp.Error = e
				sendResponse(w, resp)
				return
			}

			following, err := db.GetFollowing(person_id)

			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not get followers from database"}
				sendResponse(w, resp)
				return
			}

			resp.Payload = following

		} else {
			following, err := db.GetFollowing(user.Id)

			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: could not get followers from database"}
				sendResponse(w, resp)
				return
			}
			resp.Payload = following
		}

	} else if r.Method == "POST" {

		following, err := strconv.Atoi(strings.TrimSpace(r.FormValue("follow")))

		if err != nil {
			fmt.Println("err: ", err)
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: "Error: could not parse user id"}
			sendResponse(w, resp)
			return
		}

		followee, err := db.GetUserById(following)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not get following from database"}
		}

		if user.Id == followee.Id {
			resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: Followee id and user id are the same"}
			sendResponse(w, resp)
			return
		}

		err = db.UpdateFollowers(user.Id, following, followee.Privacy)

		if err != nil {
			if strings.Contains(fmt.Sprint(err), "awaiting approval") {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Awaiting approval"}
				sendResponse(w, resp)
				return
			}
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not get followers from database"}
			sendResponse(w, resp)
			return
		}

		if followee.Privacy == "private" {
			resp.Error = &types.Error{Type: AUTHORIZATION, Message: "User approval is required"}
		}

		//Notifications
		nick := user.FirstName + " " + user.LastName
		if user.NickName != "" {
			nick = user.NickName
		}
		followeeNick := followee.NickName
		if followeeNick == "" {
			followeeNick = followee.FirstName + " " + followee.LastName
		}

		userBasicInfo := types.UserBasicInfo{
			Id:          user.Id,
			DisplayName: nick,
			Avatar:      user.Avatar,
		}
		followeeBasicInfo := types.UserBasicInfo{
			Id:          followee.Id,
			DisplayName: followeeNick,
			Avatar:      followee.Avatar,
		}

		//1. Notification to recipient
		notificationType := ""
		content := ""
		if followee.Privacy == "private" {
			content = nick + " wants to be your follower. Approval needed."
			notificationType = NOTIFICATION_FOLLOW_ACTION_REQUEST
		} else if followee.Privacy == "public" {
			content = "You have a new follower: " + nick
			notificationType = NOTIFICATION_FOLLOW_INFO
		}

		n := types.Notification{
			Type:      notificationType,
			Content:   content,
			Sender:    userBasicInfo,
			Recipient: followeeBasicInfo}

		err = db.SaveNotification(n)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
			sendResponse(w, resp)
			return
		}

		//Send Message to recipient
		swMessage := types.WSMessage{
			Type:    NEW_NOTIFICATION,
			Payload: n,
		}
		b, err := json.Marshal(swMessage)
		if err == nil {
			notifyClient(followee.Id, b)
		} else {
			fmt.Println(err)
		}

		//2. Notification to sender
		content = ""
		notificationType = NOTIFICATION_FOLLOW_INFO
		if followee.Privacy == "private" {
			content = "Follow request has been sent to " + followeeNick
		} else if followee.Privacy == "public" {
			content = "You have are following: " + followeeNick
		}
		n = types.Notification{
			Type:      notificationType,
			Content:   content,
			Sender:    followeeBasicInfo,
			Recipient: userBasicInfo,
		}
		err = db.SaveNotification(n)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
			sendResponse(w, resp)
			return
		}

		//Send Message to sender
		swMessage = types.WSMessage{
			Type:    NEW_NOTIFICATION,
			Payload: n,
		}
		b, err = json.Marshal(swMessage)
		if err == nil {
			notifyClient(user.Id, b)
		} else {
			fmt.Println(err)
		}

		sendResponse(w, resp)
		return

	} else if r.Method == "DELETE" {

		keys, ok := r.URL.Query()["follow"]
		if !ok || len(keys) == 0 || len(keys[0]) == 0 {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing request parameter: follow"}
			sendResponse(w, resp)
			return
		}

		followingStr := keys[0]

		followingId, err := strconv.Atoi(strings.TrimSpace(followingStr))

		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: "Error: could not parse user follow"}
			sendResponse(w, resp)
			return
		}
		err = db.DeleteFollower(user.Id, followingId)

		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: could not delete follower from database"}
		}

		following, err := db.GetUserById(followingId)
		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: "Could not get follower from datatabase"}
			sendResponse(w, resp)
			return
		}

		//1. Notification to recipient(following)
		userNick := user.FirstName + " " + user.LastName
		if user.NickName != "" {
			userNick = user.NickName
		}
		followingNick := following.FirstName + " " + following.LastName
		if following.NickName != "" {
			followingNick = following.NickName
		}

		userBasicInfo := types.UserBasicInfo{
			Id:          user.Id,
			DisplayName: userNick,
			Avatar:      user.Avatar,
		}

		followingBasicInfo := types.UserBasicInfo{
			Id:          following.Id,
			DisplayName: followingNick,
			Avatar:      following.Avatar,
		}
		notificationType := NOTIFICATION_FOLLOW_INFO
		content := userNick + " has stopped following you"

		n := types.Notification{
			Type:      notificationType,
			Content:   content,
			Sender:    userBasicInfo,
			Recipient: followingBasicInfo,
		}
		err = db.SaveNotification(n)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
			sendResponse(w, resp)
			return
		}

		//Send Message to recipient
		swMessage := types.WSMessage{
			Type:    NEW_NOTIFICATION,
			Payload: n,
		}
		b, err := json.Marshal(swMessage)
		if err == nil {
			notifyClient(following.Id, b)
		} else {
			fmt.Println(err)
		}

		//Notification for sender (user)
		content = "You have stopped following: " + followingNick
		notificationType = NOTIFICATION_FOLLOW_INFO
		n = types.Notification{
			Type:      notificationType,
			Content:   content,
			Sender:    followingBasicInfo,
			Recipient: userBasicInfo,
		}
		err = db.SaveNotification(n)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
			sendResponse(w, resp)
			return
		}

		//Send Message to sender
		swMessage = types.WSMessage{
			Type:    NEW_NOTIFICATION,
			Payload: n,
		}
		b, err = json.Marshal(swMessage)
		if err == nil {
			notifyClient(user.Id, b)
		} else {
			fmt.Println(err)
		}
	}
	sendResponse(w, resp)
}

func followersHandler(w http.ResponseWriter, r *http.Request) {

	resp := types.Response{Payload: nil, Error: nil}

	user, err := getUserFromRequest(r)
	if err != nil {
		resp.Error = err
		sendResponse(w, resp)
		return
	}

	if r.Method == "GET" {

		person_id_str := ""
		params, ok := r.URL.Query()["person_id"]
		if ok && len(params[0]) > 0 {
			person_id_str = params[0]
		}

		if person_id_str != "" {
			person_id, err := strconv.Atoi(person_id_str)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", person_id_str)}
				sendResponse(w, resp)
				return
			}

			//Check Privacy
			e := isAccessRestricted(user.Id, person_id)
			if e != nil {
				fmt.Println("e: ", e)
				resp.Error = e
				sendResponse(w, resp)
				return
			}

			followers, err := db.GetFollowers(person_id)

			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not get followers from database"}
				sendResponse(w, resp)
				return
			}

			resp.Payload = followers

		} else {

			followers, err := db.GetFollowers(user.Id)

			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not get followers from database"}
				sendResponse(w, resp)
				return
			}

			resp.Payload = followers
		}

		sendResponse(w, resp)

	} else if r.Method == "POST" {

		//Approve Follower

		followeeId := user.Id

		followerId, err := strconv.Atoi(strings.TrimSpace(r.FormValue("follower_id")))

		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: "Could not parse follower id"}
			sendResponse(w, resp)
			return
		}
		follower, err := db.GetUserById(followerId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not get follower from database"}
			sendResponse(w, resp)
			return
		}

		//True or False
		approved := strings.TrimSpace(r.FormValue("approved"))

		if approved == "false" {
			err = db.DeleteFollower(followerId, followeeId)
			if err != nil {
				errAsText := fmt.Sprint(err)
				if strings.Contains(errAsText, "no active follow requests found") {
					resp.Error = &types.Error{Type: FOLLOW_REQUEST_NOT_FOUND, Message: errAsText}
				} else {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: could not get followers from database"}
				}
				sendResponse(w, resp)
				return
			}
		} else if approved == "true" {
			err = db.ApproveFollower(followerId, followeeId)
			if err != nil {
				errAsText := fmt.Sprint(err)
				if strings.Contains(errAsText, "no active follow requests found") {
					resp.Error = &types.Error{Type: FOLLOW_REQUEST_NOT_FOUND, Message: errAsText}
				} else {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Error: could not get followers from database"}
				}
				sendResponse(w, resp)
				return
			}

		}

		sendResponse(w, resp)

		userNick := user.FirstName + " " + user.LastName
		if user.NickName != "" {
			userNick = user.NickName
		}
		followerNick := follower.FirstName + " " + follower.LastName
		if follower.NickName != "" {
			followerNick = follower.NickName
		}
		userBasicInfo := types.UserBasicInfo{
			Id:          user.Id,
			DisplayName: userNick,
			Avatar:      user.Avatar,
		}
		followerBasicInfo := types.UserBasicInfo{
			Id:          follower.Id,
			DisplayName: followerNick,
			Avatar:      follower.Avatar,
		}

		//Notify approve result
		if approved == "true" {

			//Notification for follower
			notificationType := NOTIFICATION_FOLLOW_INFO
			content := "You are following: " + userNick
			n := types.Notification{
				Type:      notificationType,
				Content:   content,
				Sender:    userBasicInfo,
				Recipient: followerBasicInfo,
			}

			err = db.SaveNotification(n)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
				sendResponse(w, resp)
				return
			}

			//Send Message to follower
			swMessage := types.WSMessage{
				Type:    NEW_NOTIFICATION,
				Payload: n,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(follower.Id, b)
			} else {
				fmt.Println(err)
			}
			notificationType = NOTIFICATION_FOLLOW_INFO
			content = "You have a new follower: " + followerNick
			n = types.Notification{
				Type:      notificationType,
				Content:   content,
				Sender:    followerBasicInfo,
				Recipient: userBasicInfo,
			}
			err = db.SaveNotification(n)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
				sendResponse(w, resp)
				return
			}

			//Send Message to followee
			swMessage = types.WSMessage{
				Type:    NEW_NOTIFICATION,
				Payload: n,
			}
			b, err = json.Marshal(swMessage)
			if err == nil {
				notifyClient(user.Id, b)
			} else {
				fmt.Println(err)
			}
		}
		if approved == "false" {

			//Notification for follower
			//User is followee
			notificationType := NOTIFICATION_FOLLOW_INFO
			content := "Follow request has been rejected by: " + userNick
			n := types.Notification{
				Type:      notificationType,
				Content:   content,
				Sender:    userBasicInfo,
				Recipient: followerBasicInfo,
			}
			err = db.SaveNotification(n)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
				sendResponse(w, resp)
				return
			}

			//Send Message to follower
			swMessage := types.WSMessage{
				Type:    NEW_NOTIFICATION,
				Payload: n,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(follower.Id, b)
			} else {
				fmt.Println(err)
			}

			//Notification for followee
			notificationType = NOTIFICATION_FOLLOW_INFO
			content = "You have rejected a follow request from: " + followerNick
			n = types.Notification{
				Type:      notificationType,
				Content:   content,
				Sender:    followerBasicInfo,
				Recipient: userBasicInfo,
			}
			err = db.SaveNotification(n)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: "Could not save a notification to database"}
				sendResponse(w, resp)
				return
			}

			//Send Message to followee
			swMessage = types.WSMessage{
				Type:    NEW_NOTIFICATION,
				Payload: n,
			}
			b, err = json.Marshal(swMessage)
			if err == nil {
				notifyClient(user.Id, b)
			} else {
				fmt.Println(err)
			}
		}
	}
}

func notificationsHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, err := getUserFromRequest(r)
	if err != nil {
		resp.Error = err
		sendResponse(w, resp)
		return
	}

	if r.Method == "GET" {

		notifications, err := db.GetNotifications(user)

		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get notifications from database. %v", err)}
			sendResponse(w, resp)
			return
		}

		//Fill missing elements
		for index, n := range *notifications {
			//1. Get group
			if n.Group != nil {
				group, err := db.GetGroupById(n.Group.(int))
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
					sendResponse(w, resp)
					return
				}
				(*notifications)[index].Group = group
			}

			//2 Get Event
			if n.Event != nil {
				event, err := db.GetEventById(n.Event.(int))
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get event from database. %v", err)}
					sendResponse(w, resp)
					return
				}
				(*notifications)[index].Event = event
			}

			//3. Get Sender
			user, err := db.GetUserById(int(n.Sender.(int64)))
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get sender from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			displayName := user.NickName
			if displayName == "" {
				displayName = user.FirstName + " " + user.LastName
			}

			(*notifications)[index].Sender = types.UserBasicInfo{
				Id:          user.Id,
				DisplayName: displayName,
				Avatar:      user.Avatar,
			}

			//4. Get Recipient
			user, err = db.GetUserById(int(n.Recipient.(int64)))
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get recipient from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			displayName = user.NickName
			if displayName == "" {
				displayName = user.FirstName + " " + user.LastName
			}

			(*notifications)[index].Recipient = types.UserBasicInfo{
				Id:          user.Id,
				DisplayName: displayName,
				Avatar:      user.Avatar,
			}

		}

		resp.Payload = notifications

	} else if r.Method == "POST" {
		//Set read
		notificationId := strings.TrimSpace(r.FormValue("notification_id"))
		err := db.NotificationsSetRead(notificationId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not update notification in database. %v", err)}
			sendResponse(w, resp)
			return
		}
	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}
	sendResponse(w, resp)
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	if strings.Contains(r.URL.Path, "/groups/invites") {
		groupsInvitesHandler(w, r)
		return
	}

	if strings.Contains(r.URL.Path, "/groups/requests") {
		groupJoinRequestsHandler(w, r)
		return
	}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}

	//Extract group id
	group_id_str := ""
	if strings.Contains(r.URL.Path, "/groups/") {
		group_id_str = strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/groups/"))
	}
	group_id := -1
	if group_id_str != "" {
		var err error = nil
		group_id, err = strconv.Atoi(group_id_str)
		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", group_id_str)}
			sendResponse(w, resp)
			return
		}
	}

	if r.Method == "GET" {

		if group_id > 0 {
			group, err := db.GetGroupById(group_id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Check awating join approval
			joinRequests, err := db.GetJoinRequestByGroupIdAndMemberId(group_id, user.Id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get join requests from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			if len(joinRequests) > 0 {
				group.AwaitingJoinApproval = true
			} else {
				group.AwaitingJoinApproval = false
			}

			//Get invited members
			invited, err := db.GetInvited(group_id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get invited members from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			group.Invited = invited

			resp.Payload = group

		} else {
			groups, err := db.GetGroups()
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get groups from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			resp.Payload = groups
		}

	} else if r.Method == "POST" {
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))

		if len(title) < 2 || len(title) > 25 {
			resp.Error = &types.Error{Type: INVALID_GROUP_TITLE, Message: "Group title should be between 2 and 25 characters long"}
			sendResponse(w, resp)
			return
		}
		if len(description) < 2 || len(description) > 250 {
			resp.Error = &types.Error{Type: INVALID_GROUP_DESCRIPTION, Message: "Group description should be between 2 and 250 characters long"}
			sendResponse(w, resp)
			return
		}

		num, err := db.SaveGroup(user.Id, title, description)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save group to database. %v", err)}
			sendResponse(w, resp)
			return
		}

		resp.Payload = types.GroupId{GroupId: int(*num)}

	} else if r.Method == "DELETE" {
		if group_id > 0 {

			//Check that user is a group creator
			group, err := db.GetGroupById(group_id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			if group.Creator.(types.UserBasicInfo).Id == user.Id {
				num, err := db.DeleteGroup(group_id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not delete group from database. %v", err)}
					sendResponse(w, resp)
					return
				}

				_, err = db.DeleteEventsByGroupId(group_id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not delete events from database. %v", err)}
					sendResponse(w, resp)
					return
				}
				resp.Payload = types.RowsAffected{RowsAffected: int(*num)}
			} else {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: only creator is allowed to delete group"}
			}
		}

	} else if r.Method == "PATCH" {
		//URL example  /groups/1?action=invite&member=3&session_id=dbs-cvewf7cewfw-cew0vwev

		if group_id_str == "" || group_id < 1 {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter: group_id"}
			sendResponse(w, resp)
			return
		}

		action := strings.TrimSpace(r.FormValue("action"))
		if action == "" {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter: action"}
			sendResponse(w, resp)
			return
		}

		group, err := db.GetGroupById(group_id)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
			sendResponse(w, resp)
			return
		}

		if action == "invite" {
			memberStr := strings.TrimSpace(r.FormValue("member"))
			if memberStr == "" {
				resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter: members"}
				sendResponse(w, resp)
				return
			}

			member_id, err := strconv.Atoi(memberStr)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", memberStr)}
				sendResponse(w, resp)
				return
			}

			//Check that allowed to invite
			isAllowed := false
			if group.Creator.(types.UserBasicInfo).Id == user.Id {
				isAllowed = true
			} else if group.Members != nil && len(group.Members.([]types.UserBasicInfo)) > 0 {
				for _, u := range group.Members.([]types.UserBasicInfo) {
					if u.Id == user.Id {
						isAllowed = true
						break
					}
				}
			}
			if !isAllowed {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: only group members can send invitation"}
				sendResponse(w, resp)
				return
			}

			//Check that we don't invite members and creator
			inviteIsAllowed := true
			if member_id == group.Creator.(types.UserBasicInfo).Id {
				inviteIsAllowed = false
			} else if group.Members != nil && len(group.Members.([]types.UserBasicInfo)) > 0 {
				for _, u := range group.Members.([]types.UserBasicInfo) {
					if u.Id == member_id {
						inviteIsAllowed = false
						break
					}
				}
			}

			if !inviteIsAllowed {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: cannot invite group creator or group member"}
				sendResponse(w, resp)
				return
			}

			date := time.Now().UnixNano() / 1000000

			num, err := db.SaveInvitationToGroup(user.Id, group_id, member_id, date)

			if err != nil {
				errorStr := fmt.Sprintf("%v", err)
				if strings.Contains(errorStr, "UNIQUE constraint") {
					resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: invite aleady exists for member_id: %v and group_id: %v", member_id, group_id)}
				} else {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot add invite to database. %v", err)}
				}
				sendResponse(w, resp)
				return
			}

			resp.Payload = types.Inserted{Inserted: int(*num)}

			//Save to notification
			inviter := types.UserBasicInfo{
				Id:     user.Id,
				Avatar: user.Avatar,
			}
			displayName := user.NickName
			if displayName == "" {
				displayName = user.FirstName + " " + user.LastName
			}
			inviter.DisplayName = displayName

			n := types.Notification{
				Type:      INVITATION_TO_JOIN_GROUP,
				Content:   "Invitation to join group",
				Sender:    inviter,
				Recipient: member_id,
				Group:     group_id}
			err = db.SaveNotification(n)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save notification to database. %v", err)}
				sendResponse(w, resp)
				return
			}
			//Notify via ws
			inviteToJoinGroup := types.InviteToJoinGroup{
				Date:    date,
				Inviter: &inviter,
				Group:   *group,
			}
			swMessage := types.WSMessage{
				Type:    INVITATION_TO_JOIN_GROUP,
				Payload: inviteToJoinGroup,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(member_id, b)
			} else {
				fmt.Println(err)
			}

		}

		if action == "leave" {
			membersInterface := group.Members
			if membersInterface != nil && len(membersInterface.([]types.UserBasicInfo)) > 0 {
				//Filter members
				members := membersInterface.([]types.UserBasicInfo)

				filtered := []int{}
				for _, m := range members {
					if m.Id != user.Id {
						filtered = append(filtered, m.Id)
					}
				}
				//Convert to json
				b, err := json.Marshal(filtered)
				if err != nil {
					resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", filtered)}
					sendResponse(w, resp)
					return
				}
				//Save members
				rows, err := db.SaveMembersToGroup(group_id, string(b))
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save members from database. %v", fmt.Sprint(err))}
					sendResponse(w, resp)
					return
				}
				resp.Payload = types.RowsAffected{RowsAffected: int(*rows)}

				//Notify via ws
				member := types.UserBasicInfo{
					Id:     user.Id,
					Avatar: user.Avatar,
				}
				displayName := user.NickName
				if displayName == "" {
					displayName = user.FirstName + " " + user.LastName
				}
				member.DisplayName = displayName

				date := time.Now().UnixNano() / 1000000

				group, err = db.GetGroupById(group_id)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
					sendResponse(w, resp)
					return
				}
				leaveGroup := types.LeaveGroup{
					Date:   date,
					Member: &member,
					Group:  *group,
				}
				swMessage := types.WSMessage{
					Type:    LEAVE_GROUP,
					Payload: leaveGroup,
				}
				b, err = json.Marshal(swMessage)
				if err == nil {
					// Notify Creator
					notifyClient(group.Creator.(types.UserBasicInfo).Id, b)

					//Notify members
					if group.Members != nil {
						for _, m := range group.Members.([]types.UserBasicInfo) {
							notifyClient(m.Id, b)
						}
					}
				} else {
					fmt.Println(err)
				}

			} else {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: no group members found in database. %v", err)}
			}
		}

		if action == "join" {

			//1. Check that allowed to join (not a creator and not already a member)
			group, err := db.GetGroupById(group_id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			if group.Creator.(types.UserBasicInfo).Id == user.Id {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: group creator cannot join same group"}
				sendResponse(w, resp)
				return
			}
			if group.Members != nil && len(group.Members.([]types.UserBasicInfo)) > 0 {
				for _, m := range group.Members.([]types.UserBasicInfo) {
					if m.Id == user.Id {
						resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: already member, cannot join same group"}
						sendResponse(w, resp)
						return
					}
				}
			}

			//2. Add join request to database and handle 'request alredy exists' error
			date := time.Now().UnixNano() / 1000000
			num, err := db.SaveJoinGroupRequest(group_id, user.Id, date)
			if err != nil {
				errorStr := fmt.Sprintf("%v", err)
				if strings.Contains(errorStr, "UNIQUE constraint") {
					resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: join request aleady exists for user_id: %v and group_id: %v", user.Id, group_id)}
				} else {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot add join request to database. %v", err)}
				}
				sendResponse(w, resp)
				return
			}

			resp.Payload = types.Inserted{Inserted: int(*num)}

			//3. Send Notification via ws to creator
			//Save to notification
			n := types.Notification{
				Type:      REQUEST_TO_JOIN_GROUP,
				Content:   "Request to join group",
				Sender:    user.Id,
				Recipient: group.Creator.(types.UserBasicInfo).Id,
				Group:     group_id}
			err = db.SaveNotification(n)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save notification to database. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Notify via ws
			member := types.UserBasicInfo{
				Id:     user.Id,
				Avatar: user.Avatar,
			}
			displayName := user.NickName
			if displayName == "" {
				displayName = user.FirstName + " " + user.LastName
			}
			member.DisplayName = displayName

			group, err = db.GetGroupById(group_id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			requestToJoinGroup := types.RequestToJoinGroup{
				Date:   date,
				Member: &member,
				Group:  *group,
			}
			swMessage := types.WSMessage{
				Type:    REQUEST_TO_JOIN_GROUP,
				Payload: requestToJoinGroup,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(group.Creator.(types.UserBasicInfo).Id, b)
			} else {
				fmt.Println(err)
			}

		}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}

	sendResponse(w, resp)
}

func groupsInvitesHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}

	if r.Method == "GET" {
		invites, err := db.GetInvites(user.Id)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get invites from database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}
		resp.Payload = invites
	} else if r.Method == "PATCH" {

		group_id := strings.TrimSpace(r.FormValue("group_id"))
		if group_id == "" {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter: group_id"}
			sendResponse(w, resp)
			return
		}

		groupId, err := strconv.Atoi(group_id)
		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", group_id)}
			sendResponse(w, resp)
			return
		}

		action := strings.TrimSpace(r.FormValue("action"))

		if action == "accept" {

			//1. Add member
			// 1.1 Get current members of group
			group, err := db.GetGroupById(groupId)

			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group members from database. %v", fmt.Sprint(err))}
				sendResponse(w, resp)
				return
			}

			currentMembersInterace := group.Members
			currentMembers := []types.UserBasicInfo{}
			if currentMembersInterace != nil {
				currentMembers = currentMembersInterace.([]types.UserBasicInfo)
			}

			//1.2 Check if member already in array
			exist := false
			for _, m := range currentMembers {
				if m.Id == user.Id {
					exist = true
					break
				}
			}

			if exist {
				resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: group member already in database. %v", fmt.Sprint(err))}
			} else {
				//1.3 Append new member
				memberIds := []int{}
				for _, m := range currentMembers {
					memberIds = append(memberIds, m.Id)
				}

				memberIds = append(memberIds, user.Id)

				//1.4 Save updated array of members
				b, err := json.Marshal(memberIds)
				if err != nil {
					resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", memberIds)}
					sendResponse(w, resp)
					return
				}

				rows, err := db.SaveMembersToGroup(groupId, string(b))
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save group member to database. %v", fmt.Sprint(err))}
					sendResponse(w, resp)
					return
				}
				resp.Payload = types.Inserted{Inserted: int(*rows)}
			}

			//2. Delete invites
			_, err = db.DeleteInvites(groupId, user.Id)

			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could delete invites from database. %v", fmt.Sprint(err))}
				sendResponse(w, resp)
				return
			}

			//3. Notify Inviter
			member := types.UserBasicInfo{
				Id:     user.Id,
				Avatar: user.Avatar,
			}
			displayName := user.NickName
			if displayName == "" {
				displayName = user.FirstName + " " + user.LastName
			}
			member.DisplayName = displayName

			group, err = db.GetGroupById(groupId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			date := time.Now().UnixNano() / 1000000

			acceptJoinGroupInvite := types.AcceptJoinGroupInvite{
				Date:   date,
				Member: &member,
				Group:  *group,
			}
			swMessage := types.WSMessage{
				Type:    ACCEPT_JOIN_GROUP_INVITE,
				Payload: acceptJoinGroupInvite,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(group.Creator.(types.UserBasicInfo).Id, b)
			} else {
				fmt.Println(err)
			}

		}

		if action == "decline" {

			rows, err := db.DeleteInvites(groupId, user.Id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could delete invites from database. %v", fmt.Sprint(err))}
				sendResponse(w, resp)
				return
			}
			resp.Payload = types.Inserted{Inserted: int(*rows)}

			//3. Notify Inviter
			member := types.UserBasicInfo{
				Id:     user.Id,
				Avatar: user.Avatar,
			}
			displayName := user.NickName
			if displayName == "" {
				displayName = user.FirstName + " " + user.LastName
			}
			member.DisplayName = displayName

			group, err := db.GetGroupById(groupId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			date := time.Now().UnixNano() / 1000000

			declineJoinGroupInvite := types.DeclineJoinGroupInvite{
				Date:   date,
				Member: &member,
				Group:  *group,
			}
			swMessage := types.WSMessage{
				Type:    DECLINE_JOIN_GROUP_INVITE,
				Payload: declineJoinGroupInvite,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(group.Creator.(types.UserBasicInfo).Id, b)
			} else {
				fmt.Println(err)
			}

		}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}

	sendResponse(w, resp)
}

func groupJoinRequestsHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}

	//Extract group id
	group_id_str := ""
	if strings.Contains(r.URL.Path, "/groups/requests/") {
		group_id_str = strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/groups/requests/"))
	}
	groupId := -1
	if group_id_str != "" {
		var err error = nil
		groupId, err = strconv.Atoi(group_id_str)
		if err != nil || groupId < 1 {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", group_id_str)}
			sendResponse(w, resp)
			return
		}
	}

	if r.Method == "GET" {

		groups, err := db.GetGroupsByCreatorId(user.Id)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get groups from database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}

		requests := []types.RequestToJoinGroup{}

		for _, g := range *groups {
			joinRequests, err := db.GetJoinRequestByGroupId(g.Id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get join requests from database. %v", fmt.Sprint(err))}
				sendResponse(w, resp)
				return
			}
			for _, jr := range *joinRequests {
				user, err := db.GetUserById(jr.MemberId)
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get user from database. %v", fmt.Sprint(err))}
					sendResponse(w, resp)
					return
				}
				u := types.UserBasicInfo{
					Id:     user.Id,
					Avatar: user.Avatar,
				}
				displayName := user.NickName
				if displayName == "" {
					displayName = user.FirstName + " " + user.LastName
				}
				u.DisplayName = displayName

				request := types.RequestToJoinGroup{
					Date:   jr.Date,
					Member: &u,
					Group:  g,
				}

				requests = append(requests, request)
			}
		}
		resp.Payload = requests

	} else if r.Method == "PATCH" {

		action := strings.TrimSpace(r.FormValue("action"))
		if action == "" {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter: action"}
			sendResponse(w, resp)
			return
		}

		memberIdStr := strings.TrimSpace(r.FormValue("member_id"))
		if memberIdStr == "" {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter: member_id"}
			sendResponse(w, resp)
			return
		}

		memberId, err := strconv.Atoi(memberIdStr)
		if err != nil || memberId < 1 {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", memberIdStr)}
			sendResponse(w, resp)
			return
		}

		if action != "approve" && action != "decline" {
			resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: invalid action: %v", action)}
			sendResponse(w, resp)
			return
		}

		//1. Check that action is allowed (user is a creator of group)
		group, err := db.GetGroupById(groupId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}
		if group.Creator.(types.UserBasicInfo).Id != user.Id {
			resp.Error = &types.Error{Type: AUTHORIZATION, Message: fmt.Sprintf("Error: only creator is authorized to perform action: %v", action)}
			sendResponse(w, resp)
			return
		}

		//2. Clear join_group_request record
		_, err = db.DeleteGroupJoinRequest(groupId, memberId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not delete join request from database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}

		if action == "decline" {

			group, err = db.GetGroupById(group.Id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", fmt.Sprint(err))}
				sendResponse(w, resp)
				return
			}
			resp.Payload = group

			swMessage := types.WSMessage{
				Type:    REQUEST_TO_JOIN_GROUP_DECLINED,
				Payload: group,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(memberId, b)
			} else {
				fmt.Println(err)
			}
		}

		if action == "approve" {
			//Add member to group
			//1. Get members from group
			membersInterface := group.Members
			members := []int{}
			memberExists := false
			if membersInterface != nil {
				for _, m := range membersInterface.([]types.UserBasicInfo) {
					if m.Id == memberId {
						memberExists = true
					}
					members = append(members, m.Id)
				}
			}
			if !memberExists {
				members = append(members, memberId)
			}

			//2. Save members to group
			b, err := json.Marshal(members)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", members)}
				sendResponse(w, resp)
				return
			}
			num, err := db.SaveMembersToGroup(groupId, string(b))
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save members to database. %v", fmt.Sprint(err))}
				sendResponse(w, resp)
				return
			}
			resp.Payload = types.RowsAffected{RowsAffected: int(*num)}

			group, err = db.GetGroupById(group.Id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", fmt.Sprint(err))}
				sendResponse(w, resp)
				return
			}
			swMessage := types.WSMessage{
				Type:    REQUEST_TO_JOIN_GROUP_APPROVED,
				Payload: group,
			}
			b, err = json.Marshal(swMessage)
			if err == nil {
				notifyClient(memberId, b)
			} else {
				fmt.Println(err)
			}
		}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}

	sendResponse(w, resp)
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}

	if r.Method == "GET" {

	} else if r.Method == "POST" {

		//handle Image
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while parse multipart form %v", err)}
			sendResponse(w, resp)
			return
		}

		//Validate image
		fileName := ""

		file, _, err := r.FormFile("image")

		if err == nil {

			defer file.Close()

			err = makeDirectoryIfNotExists(IMAGES_DIRECTORY)

			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating directory %v", err)}
				sendResponse(w, resp)
				return
			}

			uuid := generateSessionId()

			tempFile, err := ioutil.TempFile(IMAGES_DIRECTORY, fmt.Sprintf("%v-*", uuid))

			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating temp file %v", err)}
				sendResponse(w, resp)
				return
			}

			defer tempFile.Close()

			fileBytes, err := ioutil.ReadAll(file)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while reading file %v", err)}
				sendResponse(w, resp)
				return
			}

			_, err = tempFile.Write(fileBytes)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while writing file %v", err)}
				sendResponse(w, resp)
				return
			}

			fileName = filepath.Base(tempFile.Name())

		} else {
			if !strings.Contains(err.Error(), "no such file") {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: image error %v", err)}
				sendResponse(w, resp)
				return
			}
		}

		postIdStr := strings.TrimSpace(r.FormValue("post_id"))
		postId, err := strconv.Atoi(postIdStr)
		if err != nil || postId < 1 {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", postIdStr)}
			sendResponse(w, resp)
			return
		}

		content := strings.TrimSpace(r.FormValue("content"))
		if content == "" || len(content) > 250 {
			resp.Error = &types.Error{Type: INVALID_COMMENT_FORMAT, Message: "Error: comment shoud be between 1 and 250 characters long"}
			sendResponse(w, resp)
			return
		}

		date := util.GetCurrentMilli()

		comment := types.Comment{
			Date:    date,
			PostId:  postId,
			Content: content,
			User:    user.Id,
			Image:   fileName}

		num, err := db.SaveComment(comment)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save comment to database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}
		resp.Payload = types.Inserted{Inserted: int(*num)}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
		sendResponse(w, resp)
	}

	sendResponse(w, resp)
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}
	if r.Method == "GET" {

		//Extract group id
		groupIdStr := ""
		params, ok := r.URL.Query()["group_id"]
		if ok && len(params[0]) > 0 {
			groupIdStr = params[0]
		}

		groupId, err := strconv.Atoi(groupIdStr)
		if err != nil || groupId < 1 {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", groupIdStr)}
			sendResponse(w, resp)
			return
		}

		//Check if user is a part of group
		group, err := db.GetGroupById(groupId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}
		isGroupMember := false
		if group.Creator.(types.UserBasicInfo).Id == user.Id {
			isGroupMember = true
		} else if group.Members != nil {
			for _, member := range group.Members.([]types.UserBasicInfo) {
				if member.Id == user.Id {
					isGroupMember = true
					break
				}
			}
		}

		if !isGroupMember {
			resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: could not get events from database. Must be the group creator or a group member"}
			sendResponse(w, resp)
			return
		}

		events, err := db.GetEvents(groupId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get events from database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}

		//Get members
		for index, event := range *events {
			if event.Members != nil && strings.TrimSpace(event.Members.(string)) != "" {
				b := []byte(event.Members.(string))
				var memberIds []int
				err = json.Unmarshal(b, &memberIds)
				if err != nil {
					resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", event.Members.(string))}
					sendResponse(w, resp)
				}

				members := []types.UserBasicInfo{}

				for _, memberId := range memberIds {
					u, err := db.GetUserById(memberId)
					if err != nil {
						resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get user from database. %v", fmt.Sprint(err))}
						sendResponse(w, resp)
						return
					}
					member := types.UserBasicInfo{}
					member.Id = u.Id
					member.Avatar = u.Avatar
					displayName := u.NickName
					if displayName == "" {
						displayName = u.FirstName + " " + u.LastName
					}
					member.DisplayName = displayName

					members = append(members, member)
				}
				(*events)[index].Members = members
			}
		}

		resp.Payload = events

	} else if r.Method == "POST" {

		title := strings.TrimSpace(r.FormValue("title"))
		if title == "" || len(title) > 50 {
			resp.Error = &types.Error{Type: INVALID_EVENT_TITLE_FORMAT, Message: "Error: event title shoud be between 1 and 50 characters long"}
			sendResponse(w, resp)
			return
		}
		description := strings.TrimSpace(r.FormValue("description"))
		if description == "" || len(description) > 250 {
			resp.Error = &types.Error{Type: INVALID_EVENT_DESCRIPTION_FORMAT, Message: "Error: event description shoud be between 1 and 250 characters long"}
			sendResponse(w, resp)
			return
		}

		eventDateStr := strings.TrimSpace(r.FormValue("event_date"))
		eventDate, err := strconv.Atoi(eventDateStr)
		if err != nil {
			resp.Error = &types.Error{Type: INVALID_DATE_FORMAT, Message: fmt.Sprintf("Error: could not parse date: %v", eventDateStr)}
			sendResponse(w, resp)
			return
		}

		members := strings.TrimSpace(r.FormValue("members"))

		groupIdStr := strings.TrimSpace(r.FormValue("group_id"))

		groupId, err := strconv.Atoi(groupIdStr)
		if err != nil || groupId < 1 {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", groupIdStr)}
			sendResponse(w, resp)
			return
		}
		//Check that user is allowed to create event (should be creator or member)
		group, err := db.GetGroupById(groupId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get group from database. %v", err)}
			sendResponse(w, resp)
			return
		}
		isAllowed := false
		if group.Creator.(types.UserBasicInfo).Id == user.Id {
			isAllowed = true
		} else if group.Members != nil {
			for _, m := range group.Members.([]types.UserBasicInfo) {
				if m.Id == user.Id {
					isAllowed = true
					break
				}
			}
		}

		if !isAllowed {
			resp.Error = &types.Error{Type: AUTHORIZATION, Message: "Error: cannot create event. Must be the group creator or a group member"}
			sendResponse(w, resp)
			return
		}

		createDate := util.GetCurrentMilli()

		event := types.Event{
			Creator:     user.Id,
			EventDate:   int64(eventDate),
			CreateDate:  createDate,
			Title:       title,
			Description: description,
			GroupId:     groupId,
			Image:       "",
			Members:     members}

		lastIndex, err := db.SaveEvent(event)

		event.Id = int(*lastIndex)

		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save event to database. %v", err)}
			sendResponse(w, resp)
			return
		}

		//Save Notification

		//Get all group members except event creator
		allMembers := []types.UserBasicInfo{}
		if group.Creator.(types.UserBasicInfo).Id != user.Id {
			allMembers = append(allMembers, group.Creator.(types.UserBasicInfo))
		}
		if group.Members != nil {
			for _, m := range group.Members.([]types.UserBasicInfo) {
				if m.Id != user.Id {
					allMembers = append(allMembers, m)
				}
			}
		}

		//Notify every user
		for _, m := range allMembers {
			//create and save notification
			n := types.Notification{
				Type:      NEW_EVENT_NOTIFICATION,
				Content:   "New event",
				Sender:    user.Id,
				Recipient: m.Id,
				Group:     groupId,
				Event:     event.Id}
			err = db.SaveNotification(n)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save notification to database. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Notify via ws
			eventCreator := types.UserBasicInfo{
				Id:     user.Id,
				Avatar: user.Avatar,
			}
			displayName := user.NickName
			if displayName == "" {
				displayName = user.FirstName + " " + user.LastName
			}
			eventCreator.DisplayName = displayName

			newEventNotification := types.NewEventNotification{
				Date:         util.GetCurrentMilli(),
				EventCreator: eventCreator,
				Group:        group,
				Event:        event,
			}
			swMessage := types.WSMessage{
				Type:    NEW_EVENT_NOTIFICATION,
				Payload: newEventNotification,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(m.Id, b)
			} else {
				fmt.Println(err)
			}

		}

		resp.Payload = types.Inserted{Inserted: int(*lastIndex)}

	} else if r.Method == "PATCH" {

		//Extract event id
		eventIdStr := ""
		if strings.Contains(r.URL.Path, "/events/") {
			eventIdStr = strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/events/"))
		}

		eventId, err := strconv.Atoi(eventIdStr)
		if err != nil || eventId < 1 {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", eventIdStr)}
			sendResponse(w, resp)
			return
		}

		attending := ""
		params, ok := r.URL.Query()["attending"]
		if ok && len(params[0]) > 0 {
			attending = strings.TrimSpace(params[0])
		}

		if attending != "true" && attending != "false" {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: fmt.Sprintf("Error: missing or invalid parameter: attending %v", attending)}
			sendResponse(w, resp)
			return
		}

		event, err := db.GetEventById(eventId)
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get event from database. %v", err)}
			sendResponse(w, resp)
			return
		}

		membersStr := strings.TrimSpace(event.Members.(string))
		if membersStr == "" {
			membersStr = "[]"
		}

		var members []int
		err = json.Unmarshal([]byte(membersStr), &members)
		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", membersStr)}
			sendResponse(w, resp)
			return
		}

		//remove member
		freshMembers := []int{}
		for _, memberId := range members {
			if memberId != user.Id {
				freshMembers = append(freshMembers, memberId)
			}
		}

		if attending == "true" {
			freshMembers = append(freshMembers, user.Id)
		}

		b, err := json.Marshal(freshMembers)
		if err != nil {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", freshMembers)}
			sendResponse(w, resp)
			return
		}

		row, err := db.SaveEventMembers(eventId, string(b))
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save event members to database. %v", err)}
			sendResponse(w, resp)
			return
		}

		resp.Payload = types.RowsAffected{RowsAffected: int(*row)}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}
	sendResponse(w, resp)
}

func chatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}

	recipientId := strings.TrimSpace(r.FormValue("recipient_id"))
	chatGroupId := strings.TrimSpace(r.FormValue("chat_group_id"))

	if r.Method == "GET" {
		if recipientId == "" && chatGroupId == "" {

			lastMessages := []types.ChatMessage{}

			//1. Get Private Messages
			privateMessages, err := db.GetPrivateMessagesByUserId(user.Id)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get private messages from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			lastMessages = append(lastMessages, *privateMessages...)

			//2. Get Group Messages
			chatGroupMessages, err := db.GetChatGroupMessagesByMemberId(user.Id)
			if err != nil {
				fmt.Println(err)
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get chat group messages from database. %v", err)}
				sendResponse(w, resp)
				return
			}
			lastMessages = append(lastMessages, *chatGroupMessages...)

			//3. Remove Doublicates
			/*
				lastMessagesFiltered := []types.ChatMessage{}
				for _, message := range lastMessages {
					isDublicate := false
					for i, m := range lastMessagesFiltered {
						// Check if dublicate Private Message
						if message.ChatGroup == nil && m.ChatGroup == nil {
							if (message.Sender.(int) == m.Sender.(int) && message.Recipient.(int) == m.Recipient.(int)) ||
								(message.Sender.(int) == m.Recipient.(int) && message.Recipient.(int) == m.Sender.(int)) {
								isDublicate = true
								//Check Date
								if isDublicate && message.Date > m.Date {
									lastMessagesFiltered[i] = message
								}
							}
						}
						// Check if dublicate Group Message
						if message.ChatGroup != nil && m.ChatGroup != nil {
							if message.ChatGroup.(int) == m.ChatGroup.(int) {
								isDublicate = true
								//Check Date
								if isDublicate && message.Date > m.Date {
									lastMessagesFiltered[i] = message
								}
							}
						}
					}
					if !isDublicate {
						lastMessagesFiltered = append(lastMessagesFiltered, message)
					}
				}
				lastMessages = lastMessagesFiltered
			*/
			//4. Order By Date
			isOrdered := false
			for !isOrdered {
				isOrdered = true
				for i := 0; i < len(lastMessages)-1; i++ {
					m1 := lastMessages[i]
					m2 := lastMessages[i+1]
					if m2.Date > m1.Date {
						lastMessages[i] = m2
						lastMessages[i+1] = m1
						isOrdered = false
						break
					}
				}
			}

			//5. Fill missing data
			for index, message := range lastMessages {
				lastMessages[index].Sender = ToUserBasicInfo(message.Sender.(int))
				if message.Recipient != nil {
					lastMessages[index].Recipient = ToUserBasicInfo(message.Recipient.(int))
				}
				if message.ChatGroup != nil {
					chatGroup, err := db.GetChatGroupById(message.ChatGroup.(int))
					if err != nil {
						resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get chat group from database. %v", err)}
						sendResponse(w, resp)
						return
					}
					lastMessages[index].ChatGroup = chatGroup
				}
			}

			resp.Payload = lastMessages

			// m1 := types.ChatMessage{
			// 	Id:        2,
			// 	ChatGroup: nil,
			// 	Content:   "Private Message to Alice",
			// 	Date:      1711978614079,
			// 	IsRead:    true,
			// 	ReadBy:    nil,
			// 	Recipient: types.UserBasicInfo{
			// 		Id:          1,
			// 		DisplayName: "alice",
			// 		Avatar:      "70e47853-cc1d-47fb-a6c0-b6412c9034a7-2660343401.gif"},
			// 	Sender: types.UserBasicInfo{
			// 		Id:          2,
			// 		DisplayName: "Bob Smyth",
			// 		Avatar:      ""}}

			// m2 := types.ChatMessage{
			// 	ChatGroup: types.Group{
			// 		Date:    1711982414711,
			// 		Id:      1,
			// 		Image:   "",
			// 		Members: "[2,1]",
			// 		Title:   "Room 1",
			// 	},
			// 	Content:   "Room 1 Message By Bob",
			// 	Date:      1711986635863,
			// 	Id:        5,
			// 	IsRead:    false,
			// 	ReadBy:    "[1]",
			// 	Recipient: nil,
			// 	Sender: types.UserBasicInfo{
			// 		Id:          2,
			// 		DisplayName: "Bob Smyth",
			// 		Avatar:      ""}}

			// messages := []types.ChatMessage{}
			// messages = append(messages, m1, m2)

		} else if recipientId != "" {
			//Handle Private Message
			recipientId, err := strconv.Atoi(recipientId)
			if err != nil || recipientId < 1 {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", recipientId)}
				sendResponse(w, resp)
				return
			}
			messages, err := db.GetPrivateMessages(user.Id, recipientId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get messages grom database. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Update Message users
			for index, message := range *messages {
				sender := ToUserBasicInfo(message.Sender.(int))
				recipient := ToUserBasicInfo(message.Recipient.(int))
				(*messages)[index].Sender = sender
				(*messages)[index].Recipient = recipient
			}
			resp.Payload = messages
		} else if chatGroupId != "" {
			//Handle Chat Group
			chatGroupId, err := strconv.Atoi(chatGroupId)
			if err != nil || chatGroupId < 1 {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", chatGroupId)}
				sendResponse(w, resp)
				return
			}
			messages, err := db.GetChatGroupMessages(chatGroupId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get chat group messages grom database. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Get Extra Data
			chatGroup, err := db.GetChatGroupById(chatGroupId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get chat group grom database. %v", err)}
				sendResponse(w, resp)
				return
			}
			for index, message := range *messages {
				(*messages)[index].ChatGroup = chatGroup
				(*messages)[index].Sender = ToUserBasicInfo(message.Sender.(int))
			}

			resp.Payload = messages
		}

	} else if r.Method == "POST" {
		content := strings.TrimSpace(r.FormValue("message"))
		if content == "" {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter in http request: message"}
			sendResponse(w, resp)
			return
		}

		if recipientId == "" && chatGroupId == "" {
			resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter in http request: recipient_id or chat_group_id"}
			sendResponse(w, resp)
			return
		}

		date := util.GetCurrentMilli()

		if recipientId != "" {
			//Handle private message
			recipientId, err := strconv.Atoi(recipientId)
			if err != nil || recipientId < 1 {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", recipientId)}
				sendResponse(w, resp)
				return
			}

			id, err := db.SavePrivateMessage(user.Id, recipientId, date, content)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save chat messages to database. %v", err)}
				sendResponse(w, resp)
				return
			}

			message := types.ChatMessage{
				Id:        *id,
				Sender:    ToUserBasicInfo(user.Id),
				Recipient: ToUserBasicInfo(recipientId),
				Date:      date,
				Content:   content,
				IsRead:    false,
			}

			resp.Payload = message

			//Update via ws
			swMessage := types.WSMessage{
				Type:    NEW_CHAT_MESSAGE,
				Payload: message,
			}
			b, err := json.Marshal(swMessage)
			if err == nil {
				notifyClient(recipientId, b)
			} else {
				fmt.Println(err)
			}

			// id, err = UpdateInbox(*id, user.Id, &recipientId, nil, date)
			// if err != nil {
			// 	resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not update inbox. %v", err)}
			// 	sendResponse(w, resp)
			// 	return
			// }

		} else if chatGroupId != "" {
			//Handle group message
			chatGroupId, err := strconv.Atoi(chatGroupId)
			if err != nil || chatGroupId < 1 {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", chatGroupId)}
				sendResponse(w, resp)
				return
			}

			id, err := db.SaveChatGroupMessage(user.Id, chatGroupId, date, content)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not save chat messages to database. %v", err)}
				sendResponse(w, resp)
				return
			}
			chatGroup, err := db.GetChatGroupById(chatGroupId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get chat group from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			message := types.ChatMessage{
				Id:        *id,
				Sender:    ToUserBasicInfo(user.Id),
				ChatGroup: chatGroup,
				Date:      date,
				Content:   content,
				IsRead:    false,
			}

			resp.Payload = message

			//Update Chat Group Members
			members := []int{}
			if chatGroup.Members == nil {
				chatGroup.Members = ""
			}
			err = json.Unmarshal([]byte(chatGroup.Members.(string)), &members)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: cannot parse chat group members. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Remove member if exists
			filtered := []int{}
			for _, memberId := range members {
				if memberId != user.Id {
					filtered = append(filtered, memberId)
				}
			}
			members = filtered
			//Add new Member
			members = append(members, user.Id)

			b, err := json.Marshal(members)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: cannot parse chat group members. %v", err)}
				sendResponse(w, resp)
				return
			}

			_, err = db.SaveChatGroupMembers(chatGroupId, string(b))
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot save chat room members. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Notify via ws
			for _, memberId := range members {
				swMessage := types.WSMessage{
					Type:    NEW_CHAT_MESSAGE,
					Payload: message}

				b, err := json.Marshal(swMessage)
				if err == nil {
					notifyClient(memberId, b)
				} else {
					fmt.Println(err)
				}
			}

		}

	} else if r.Method == "PATCH" {

		resp.Payload = types.Updated{Updated: 0}

		field := strings.TrimSpace(r.FormValue("field"))
		valueStr := strings.TrimSpace(r.FormValue("value"))
		personIdStr := strings.TrimSpace(r.FormValue("person_id"))
		chatGroupIdStr := strings.TrimSpace(r.FormValue("chat_group_id"))

		if field == "is_read" {

			if personIdStr == "" {
				resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter in http request: person_id"}
				sendResponse(w, resp)
				return
			}

			personId, err := strconv.Atoi(personIdStr)
			if err != nil || personId < 1 {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", personIdStr)}
				sendResponse(w, resp)
				return
			}

			if valueStr == "" {
				resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter in http request: value"}
				sendResponse(w, resp)
				return
			}

			value, err := strconv.ParseBool(valueStr)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", valueStr)}
				sendResponse(w, resp)
				return
			}

			num, err := db.UpdateIsRead(value, user.Id, personId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not update chat messages in database. %v", err)}
				sendResponse(w, resp)
				return
			}

			updated := types.Updated{
				Updated: int(*num),
			}
			resp.Payload = updated
		}

		if field == "read_by" {
			if valueStr == "" {
				resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter in http request: value"}
				sendResponse(w, resp)
				return
			}
			value, err := strconv.ParseBool(valueStr)
			if err != nil {
				fmt.Println("err 1 ", err)
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", valueStr)}
				sendResponse(w, resp)
				return
			}

			if chatGroupIdStr == "" {
				fmt.Println(err)
				resp.Error = &types.Error{Type: MISSING_PARAM, Message: "Error: missing parameter in http request: chat_group_id"}
				sendResponse(w, resp)
				return
			}

			chatGroupId, err := strconv.Atoi(chatGroupIdStr)
			if err != nil || chatGroupId < 1 {
				fmt.Println("err 2 ", err)
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", chatGroupIdStr)}
				sendResponse(w, resp)
				return
			}
			// Get messages from chat group
			messages, err := db.GetChatGroupMessages(chatGroupId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot get messages from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			for _, message := range *messages {

				if message.ReadBy == nil {
					message.ReadBy = "[]"
				}
				var readBys []int
				err := json.Unmarshal([]byte(message.ReadBy.(string)), &readBys)
				if err != nil {
					fmt.Println("err 3 ", err)
					resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", message.ReadBy)}
					sendResponse(w, resp)
					return
				}

				//Remove existing person
				filtered := []int{}
				for _, readBy := range readBys {
					if readBy != user.Id {
						filtered = append(filtered, readBy)
					}
				}
				readBys = filtered
				if value {
					filtered = append(readBys, user.Id)
				}
				readBys = filtered

				b, err := json.Marshal(readBys)
				if err != nil {
					fmt.Println("err 4 ", err)
					resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", filtered)}
					sendResponse(w, resp)
					return
				}

				num, err := db.UpdateReadBy(message.Id, string(b))
				if err != nil {
					resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot save read_by to database. %v", err)}
					sendResponse(w, resp)
					return
				}
				updated := types.Updated{
					Updated: int(*num),
				}
				resp.Payload = updated
			}
		}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}
	sendResponse(w, resp)
}

func chatGroupsHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.Response{Payload: nil, Error: nil}

	user, e := getUserFromRequest(r)
	if e != nil {
		resp.Error = e
		sendResponse(w, resp)
		return
	}
	resp.Payload = []types.ChatMessage{}

	if r.Method == "GET" {
		chatGroups, err := db.GetChatGroups()
		if err != nil {
			resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get chat groups from database. %v", fmt.Sprint(err))}
			sendResponse(w, resp)
			return
		}
		resp.Payload = chatGroups

	} else if r.Method == "POST" {

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			fmt.Println("muliparce ", err)
			resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while parse multipart form %v", err)}
			sendResponse(w, resp)
			return
		}

		title := strings.TrimSpace(r.FormValue("title"))

		if title == "" || len(title) > 50 {
			resp.Error = &types.Error{Type: INVALID_ROOM_CHAT_TITLE_FORMAT, Message: "Error: chat room title shoud be between 1 and 50 characters long"}
			sendResponse(w, resp)
			return
		}

		fileName := ""
		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			err = makeDirectoryIfNotExists(IMAGES_DIRECTORY)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating directory %v", err)}
				sendResponse(w, resp)
				return
			}
			uuid := generateSessionId()
			tempFile, err := os.CreateTemp(IMAGES_DIRECTORY, fmt.Sprintf("%v-*.gif", uuid))
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while creating temp file %v", err)}
				sendResponse(w, resp)
				return
			}
			defer tempFile.Close()
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while reading file %v", err)}
				sendResponse(w, resp)
				return
			}
			_, err = tempFile.Write(fileBytes)
			if err != nil {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: error while writing file %v", err)}
				sendResponse(w, resp)
				return
			}
			fileName = filepath.Base(tempFile.Name())

		} else {
			if !strings.Contains(err.Error(), "no such file") {
				resp.Error = &types.Error{Type: IMAGE_UPLOAD_ERROR, Message: fmt.Sprintf("Error: image error %v", err)}
				sendResponse(w, resp)
				return
			}
		}

		members := fmt.Sprintf("[%v]", user.Id)
		date := util.GetCurrentMilli()

		id, err := db.SaveChatGroup(date, title, fileName, members)
		if err != nil {
			errorStr := fmt.Sprintf("%v", err)
			if strings.Contains(errorStr, "UNIQUE constraint") {
				resp.Error = &types.Error{Type: INVALID_CHAT_ROOM, Message: "Error: chat room already exists"}
			} else {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot save chat room. %v", err)}
			}
		} else {
			resp.Payload = types.Inserted{Inserted: *id}
		}

	} else if r.Method == "PATCH" {
		resp.Payload = types.Updated{Updated: 0}

		//Get Chat Room id
		chatGroupsId := -1
		if strings.Contains(r.URL.Path, "/chatgroups/") {
			chatGroupsIdStr := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/chatgroups/"))
			if chatGroupsIdStr != "" {
				var err error
				chatGroupsId, err = strconv.Atoi(chatGroupsIdStr)
				if err != nil || chatGroupsId < 1 {
					resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could parse: %v", chatGroupsIdStr)}
					sendResponse(w, resp)
					return
				}
			}
		}
		if chatGroupsId < 1 {
			resp.Error = &types.Error{Type: PARSE_ERROR, Message: "Error: could parse chat group id"}
			sendResponse(w, resp)
			return
		}

		addMemberId := strings.TrimSpace(r.FormValue("add_member"))
		if addMemberId != "" {
			addMemberId, err := strconv.Atoi(addMemberId)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: could not parse: %v", addMemberId)}
				sendResponse(w, resp)
				return
			}

			chatGroup, err := db.GetChatGroupById(chatGroupsId)
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: could not get chat group from database. %v", err)}
				sendResponse(w, resp)
				return
			}

			if chatGroup.Members == nil {
				chatGroup.Members = "[]"
			}

			newMembers := []int{}

			err = json.Unmarshal([]byte(chatGroup.Members.(string)), &newMembers)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: cannot parse chat room members. %v", err)}
				sendResponse(w, resp)
				return
			}

			//Remove member if exists
			for index, memberId := range newMembers {
				if memberId == user.Id {
					newMembers = append(newMembers[:index], newMembers[index+1:]...)
					break
				}
			}

			//Add new Member
			newMembers = append(newMembers, user.Id)

			b, err := json.Marshal(newMembers)
			if err != nil {
				resp.Error = &types.Error{Type: PARSE_ERROR, Message: fmt.Sprintf("Error: cannot parse chat room members. %v", err)}
				sendResponse(w, resp)
				return
			}

			id, err := db.SaveChatGroupMembers(chatGroupsId, string(b))
			if err != nil {
				resp.Error = &types.Error{Type: DATABASE_ERROR, Message: fmt.Sprintf("Error: cannot save chat group members. %v", err)}
				sendResponse(w, resp)
				return
			}

			resp.Payload = types.Updated{Updated: *id}
		}

	} else {
		resp.Error = &types.Error{Type: WRONG_METHOD, Message: "Error: wrong http method"}
	}
	sendResponse(w, resp)
}
