/** Default port for the application. */
const DEFAULT_PORT = 8080;

/** Application name. */
let appName = "myapp";

/**
 * AppConfig holds application configuration.
 * Supports multiple environments.
 */
export class AppConfig {
  host: string;
  port: number;

  constructor(host: string, port: number = DEFAULT_PORT) {
    this.host = host;
    this.port = port;
  }

  validate(): boolean {
    if (!this.host) {
      throw new Error("host is required");
    }
    if (this.port <= 0) {
      throw new Error("port must be positive");
    }
    return true;
  }

  get address(): string {
    return `${this.host}:${this.port}`;
  }
}

/**
 * AppFormatter interface for output formatting.
 */
interface AppFormatter {
  format(data: string): string;
  reset(): void;
}

/** Application status. */
enum AppStatus {
  Active = "active",
  Inactive = "inactive",
  Maintenance = "maintenance",
}

/** AppOptions type alias. */
type AppOptions = {
  host: string;
  port: number;
  logLevel: string;
};

/**
 * sayHelloApp is a standalone function.
 */
export function sayHelloApp(name: string): void {
  const greeting = `Hello, ${name}!`;
  console.log(greeting);
}

/** formatNameApp is an arrow function. */
export const formatNameApp = (name: string): string => {
  return name.trim().toUpperCase();
};

namespace AppUtils {
  export function helper(): void {}
}
