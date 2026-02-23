from services.user_service import process_user_data, validate_email, format_user_name


class UserViewSet:
    def update(self, request):
        user_id = 1
        data = {"name": "Alice"}
        result = process_user_data(user_id, data)
        return result

    def validate(self, request):
        email = request.get("email", "")
        return validate_email(email)


def standalone_handler(first, last):
    name = format_user_name(first, last)
    return name
