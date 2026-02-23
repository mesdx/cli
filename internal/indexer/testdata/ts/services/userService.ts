export const MAX_WORKERS = 4;

export function processUserData(userId: number, data: Record<string, string>): boolean {
    return Object.keys(data).length > 0;
}

export function validateEmail(email: string): boolean {
    return email.includes("@");
}

export const formatUserName = (first: string, last: string): string => {
    return `${first.trim()} ${last.trim()}`;
};

export class UserRepository {
    findById(userId: number): Record<string, string> | null {
        return { id: String(userId) };
    }

    save(user: Record<string, string>): boolean {
        return true;
    }
}
