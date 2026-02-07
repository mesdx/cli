/** Default port for the application. */
const DEFAULT_PORT_JS = 8080;

/** Application name. */
let appNameJs = "myapp";

/**
 * AppConfigJs holds application configuration.
 */
class AppConfigJs {
  constructor(host, port = DEFAULT_PORT_JS) {
    this.host = host;
    this.port = port;
  }

  validate() {
    if (!this.host) {
      throw new Error("host is required");
    }
    if (this.port <= 0) {
      throw new Error("port must be positive");
    }
    return true;
  }

  get address() {
    return `${this.host}:${this.port}`;
  }
}

/**
 * sayHelloJs is a standalone function.
 */
function sayHelloJs(name) {
  const greeting = `Hello, ${name}!`;
  console.log(greeting);
  return greeting;
}

/** formatNameJs is an arrow function. */
const formatNameJs = (name) => {
  return name.trim().toUpperCase();
};

module.exports = { AppConfigJs, sayHelloJs, formatNameJs };
