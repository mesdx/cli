package services

const MaxWorkers = 4

func ProcessUserData(userID int, data map[string]string) bool {
	if len(data) == 0 {
		return false
	}
	return true
}

func ValidateEmail(email string) bool {
	for _, ch := range email {
		if ch == '@' {
			return true
		}
	}
	return false
}

func FormatUserName(first, last string) string {
	return first + " " + last
}

type UserRepository struct{}

func (r *UserRepository) FindByID(userID int) map[string]string {
	return map[string]string{"id": "1"}
}

func (r *UserRepository) Save(user map[string]string) bool {
	return true
}
