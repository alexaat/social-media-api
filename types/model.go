package types

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Response struct {
	Payload interface{} `json:"payload"`
	Error   *Error      `json:"error"`
}

type Echo struct {
	User string `json:"user"`
}

type User struct {
	Id          int    `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	NickName    string `json:"nick_name"`
	DateOfBirth int64  `json:"date_of_birth"`
	Email       string `json:"email"`
	Password    string `json:"-"`
	AboutMe     string `json:"about_me"`
	Avatar      string `json:"avatar"`
	Privacy     string `json:"privacy"`
}

type UserBasicInfo struct {
	Id          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
}

type Session struct {
	SessionId string `json:"session_id"`
}

type Post struct {
	Id       int         `json:"id"`
	Date     int64       `json:"date"`
	Group    interface{} `json:"group"`
	Content  string      `json:"content"`
	Privacy  string      `json:"privacy"`
	Image    string      `json:"image"`
	User     User        `json:"user"`
	Comments []Comment   `json:"comments"`
}

type GroupPost struct {
	Id       int       `json:"id"`
	GroupId  int       `json:"group_id"`
	Date     int64     `json:"date"`
	Content  string    `json:"content"`
	Image    string    `json:"image"`
	User     User      `json:"user"`
	Comments []Comment `json:"comments"`
}

type HomePageData struct {
	User  *User   `json:"user"`
	Posts *[]Post `json:"posts"`
}

type ProfilePageData struct {
	User      *User   `json:"user"`
	Posts     *[]Post `json:"posts"`
	Followers *[]User `json:"followers"`
	Following *[]User `json:"following"`
}

type Followers struct {
	Following []int `json:"following"`
	Followers []int `json:"followers"`
}

type Follower struct {
	Follower *User `json:"follower"`
	Date     int64 `json:"date"`
	Approved bool  `json:"approved"`
}

type Following struct {
	Following *User `json:"following"`
	Date      int64 `json:"date"`
	Approved  bool  `json:"approved"`
}

type Notification struct {
	Id        int         `json:"id"`
	Date      int64       `json:"date"`
	Type      string      `json:"type"`
	Content   string      `json:"content"`
	Sender    interface{} `json:"sender"`
	Recipient interface{} `json:"recipient"`
	IsRead    bool        `json:"is_read"`
	Group     interface{} `json:"group"`
	Event     interface{} `json:"event"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PostId struct {
	PostId int `json:"post_id"`
}

type ChatMessage struct {
	Id      int         `json:"id"`
	Date    int64       `json:"date"`
	Content string      `json:"content"`
	Sender  interface{} `json:"sender"`
	// SenderDisplayName interface{} `json:"sender_display_name"`
	Recipient interface{} `json:"recipient"`
	ChatGroup interface{} `json:"chat_group"`
	IsRead    bool        `json:"is_read"`
	// IsRoom bool  `json:"is_room"`
	ReadBy interface{} `json:"read_by"`
}
type ChatMessageId struct {
	ChatMessageId int `json:"chat_message_id"`
}

type Updated struct {
	Updated int `json:"updated"`
}

type Inserted struct {
	Inserted int `json:"updated"`
}

type Group struct {
	Id                   int         `json:"id"`
	Date                 int64       `json:"date"`
	Title                string      `json:"title"`
	Description          string      `json:"description"`
	Image                string      `json:"image"`
	Creator              interface{} `json:"creator"`
	Members              interface{} `json:"members"`
	Invited              interface{} `json:"invited"`
	AwaitingJoinApproval bool        `json:"awaiting_join_approval"`
}

type GroupBasicInfo struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type GroupId struct {
	GroupId int `json:"group_id"`
}

type RowsAffected struct {
	RowsAffected int `json:"rows_affected"`
}

type RowsAffectedArray struct {
	RowsAffected []int `json:"rows_affected_array"`
}

type InviteToJoinGroup struct {
	Date    int64          `json:"date"`
	Inviter *UserBasicInfo `json:"inviter"`
	Group   Group          `json:"group"`
}

type AcceptJoinGroupInvite struct {
	Date   int64          `json:"date"`
	Member *UserBasicInfo `json:"member"`
	Group  Group          `json:"group"`
}

type DeclineJoinGroupInvite struct {
	Date   int64          `json:"date"`
	Member *UserBasicInfo `json:"member"`
	Group  Group          `json:"group"`
}

type LeaveGroup struct {
	Date   int64          `json:"date"`
	Member *UserBasicInfo `json:"member"`
	Group  Group          `json:"group"`
}

type RequestToJoinGroup struct {
	Date   int64          `json:"date"`
	Member *UserBasicInfo `json:"member"`
	Group  Group          `json:"group"`
}

type JoinRequest struct {
	Id       int
	Date     int64
	GroupId  int
	MemberId int
}

type Comment struct {
	Id      int         `json:"id"`
	Date    int64       `json:"date"`
	User    interface{} `json:"user"`
	PostId  int         `json:"post_id"`
	Content string      `json:"content"`
	Image   string      `json:"image"`
}

type Event struct {
	Id          int         `json:"id"`
	GroupId     int         `json:"group_id"`
	Creator     interface{} `json:"creator"`
	CreateDate  int64       `json:"create_date"`
	EventDate   int64       `json:"event_date"`
	Image       string      `json:"image"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Members     interface{} `json:"members"`
}

type NewEventNotification struct {
	Date         int64       `json:"date"`
	EventCreator interface{} `json:"event_creator"`
	Group        interface{} `json:"group"`
	Event        interface{} `json:"event"`
}

type ChatGroup struct {
	Id      int         `json:"id"`
	Date    int64       `json:"date"`
	Title   string      `json:"title"`
	Image   string      `json:"image"`
	Members interface{} `json:"members"`
}
