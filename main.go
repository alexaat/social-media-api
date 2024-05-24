package main

import (
	"fmt"
	"log"
	"my-social-network/db/sqlite"
	"strings"

	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, err := sqlite.CreateDatabase()
	if err != nil {
		if !strings.Contains(err.Error(), "no change") {
			log.Fatal(err)
		}
	}

	fmt.Println("Startig server at port 8080")
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/signin", signinHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/signout", signoutHandler)
	http.HandleFunc("/image", imageHandler)
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/user/", userHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/posts/", postsHandler)
	http.HandleFunc("/followers", followersHandler)
	http.HandleFunc("/followers/", followersHandler)
	http.HandleFunc("/following", followingHandler)
	http.HandleFunc("/following/", followingHandler)
	http.HandleFunc("/notifications", notificationsHandler)
	http.HandleFunc("/chatmessages", chatMessagesHandler)
	http.HandleFunc("/chatmessages/", chatMessagesHandler)
	http.HandleFunc("/groups", groupsHandler)
	http.HandleFunc("/groups/", groupsHandler)
	http.HandleFunc("/comments", commentsHandler)
	http.HandleFunc("/events", eventsHandler)
	http.HandleFunc("/events/", eventsHandler)
	http.HandleFunc("/chatgroups", chatGroupsHandler)
	http.HandleFunc("/chatgroups/", chatGroupsHandler)

	http.HandleFunc("/ws", wsHandler)

	http.ListenAndServe(":8080", nil)

	defer db.Close()
}
