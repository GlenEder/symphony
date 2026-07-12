/**
 * Tests for maestro/static/script.js shared utilities.
 * Run with: node __tests__/maestro-script-utils.test.js
 */

const fs = require("fs");
const path = require("path");

// Load the script.js file and evaluate it in a DOM-like environment
const scriptJs = fs.readFileSync(
	path.join(__dirname, "..", "maestro", "static", "script.js"),
	"utf8"
);

// Minimal DOM mock for escapeHtml
global.document = {
	createElement(tag) {
		return {
			tagName: tag.toUpperCase(),
			textContent: "",
			innerHTML: "",
			set textContent(v) {
				this._text = v;
				this.innerHTML = v
					.replace(/&/g, "&amp;")
					.replace(/</g, "&lt;")
					.replace(/>/g, "&gt;")
					.replace(/"/g, "&quot;")
					.replace(/'/g, "&#039;");
			},
			get textContent() {
				return this._text || "";
			}
		};
	}
};
global.document.querySelector = () => null; // updateAgentDot needs this

// Evaluate the script
eval(scriptJs);

let passed = 0;
let failed = 0;

function assert(condition, label) {
	if (condition) {
		console.log(`  PASS: ${label}`);
		passed++;
	} else {
		console.log(`  FAIL: ${label}`);
		failed++;
	}
}

console.log("=== escapeHtml ===");
assert(typeof escapeHtml === "function", "escapeHtml is a function");
assert(escapeHtml("<script>alert('xss')</script>") === "&lt;script&gt;alert(&#039;xss&#039;)&lt;/script&gt;", "escapes HTML tags and quotes");
assert(escapeHtml("hello & world") === "hello &amp; world", "escapes ampersands");
assert(escapeHtml("simple text") === "simple text", "passes through plain text");
assert(escapeHtml("") === "", "handles empty string");
assert(escapeHtml(null) === "", "handles null");
assert(escapeHtml(undefined) === "", "handles undefined");

console.log("\n=== updateAgentDot ===");
assert(typeof updateAgentDot === "function", "updateAgentDot is a function");
// Can't fully test DOM manipulation without jsdom, but function exists and handles gracefully

console.log(`\n${passed} passed, ${failed} failed`);
if (failed > 0) process.exit(1);
