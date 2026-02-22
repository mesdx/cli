import { processUserData, validateEmail, formatUserName } from "../services/userService";

export function handleUpdate(userId: number, data: Record<string, string>): boolean {
    const result = processUserData(userId, data);
    return result;
}

export function handleValidate(email: string): boolean {
    return validateEmail(email);
}

export function handleFormat(first: string, last: string): string {
    return formatUserName(first, last);
}
