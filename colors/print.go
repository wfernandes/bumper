package colors

func Red(s string) string {
	return "\033[38;5;202m" + s + "\033[0m"
}

func ExtraRed(s string) string {
	return "\033[38;5;222m" + s + "\033[0m"
}

func Green(s string) string {
	return "\033[38;5;82m" + s + "\033[0m"
}

func Blue(s string) string {
	return "\033[1;34m" + s + "\033[0m"
}

func Yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func Grey(s string) string {
	return "\033[38;5;242m" + s + "\033[0m"
}

func Prince(s string) string {
	return "\033[38;5;92m" + s + "\033[0m"
}
