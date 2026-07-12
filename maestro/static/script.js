/**
 * Maestro — Shared Utility Functions
 *
 * These utilities are loaded in the <head> via base.html and are available
 * to all page-specific inline scripts.
 */

/**
 * Escape a string for safe innerHTML assignment.
 * Prevents XSS when interpolating untrusted text into HTML.
 */
function escapeHtml(str) {
	if (str == null) return "";
	const div = document.createElement("div");
	div.textContent = str;
	return div.innerHTML;
}

/**
 * Update the agent status dot in the UI.
 * @param {"listening"|"thinking"|"offline"|null} status
 */
function updateAgentDot(status) {
	const dot = document.querySelector(".agent-status-dot");
	if (!dot) return;
	dot.classList.remove("listening", "thinking", "offline");
	if (status === "thinking") {
		dot.classList.add("thinking");
		dot.title = "Agent is thinking…";
	} else if (status === "offline" || !status) {
		dot.classList.add("offline");
		dot.title = "Agent offline";
	} else {
		dot.classList.add("listening");
		dot.title = "Agent is listening";
	}
}
