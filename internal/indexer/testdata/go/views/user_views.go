package views

import "services"

func HandleUpdate(userID int, data map[string]string) bool {
	result := services.ProcessUserData(userID, data)
	return result
}

func HandleValidate(email string) bool {
	return services.ValidateEmail(email)
}

func HandleFormat(first, last string) string {
	return services.FormatUserName(first, last)
}
