package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	infoStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	warnStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	labelStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	listBullet   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).SetString("  -")
)

func Success(label, value string) {
	fmt.Printf("%s %s\n", successStyle.Render(label), valueStyle.Render(value))
}

func Info(label, value string) {
	fmt.Printf("%s %s\n", infoStyle.Render(label), valueStyle.Render(value))
}

func Warn(msg string) {
	fmt.Println(warnStyle.Render(msg))
}

func ListItem(value string) {
	fmt.Printf("%s %s\n", listBullet, valueStyle.Render(value))
}

func Error(err error) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("Error:")+" "+err.Error())
}

func Detail(output string) {
	fmt.Printf("\n%s\n", labelStyle.Render(output))
}
