package bot

type UserState struct {
	UserID        int64
	State         string
	CategoryID    int
	SubcategoryID int
	PhotoFileID   string
	Title         string
}

var UserStates = make(map[int64]UserState)

func SetUserState(userID int64, state string, categoryID int) {
	UserStates[userID] = UserState{
		UserID:     userID,
		State:      state,
		CategoryID: categoryID,
	}
}

func GetUserState(userID int64) (UserState, bool) {
	state, exists := UserStates[userID]
	return state, exists
}

func UpdateUserState(userID int64, updates UserState) {
	if state, exists := UserStates[userID]; exists {
		if updates.PhotoFileID != "" {
			state.PhotoFileID = updates.PhotoFileID
		}
		if updates.Title != "" {
			state.Title = updates.Title
		}
		if updates.State != "" {
			state.State = updates.State
		}
		UserStates[userID] = state
	}
}

func ClearUserState(userID int64) {
	delete(UserStates, userID)
}
