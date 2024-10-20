package menu

import (
	"math/rand"

	"github.com/charmbracelet/lipgloss"
)

var cyberpunkColors = []lipgloss.Color{
	lipgloss.Color("#6A5ACD"), // SlateBlue
	lipgloss.Color("#00BFFF"), // DeepSkyBlue
	lipgloss.Color("#FF69B4"), // HotPink
	lipgloss.Color("#FFD700"), // Gold
	lipgloss.Color("#32CD32"), // LimeGreen
	lipgloss.Color("#FF6347"), // Tomato
	lipgloss.Color("#4682B4"), // SteelBlue
	lipgloss.Color("#9370DB"), // MediumPurple
	lipgloss.Color("#7FFF00"), // Chartreuse
	lipgloss.Color("#FF1493"), // DeepPink
	lipgloss.Color("#00FA9A"), // MediumSpringGreen
	lipgloss.Color("#FF4500"), // OrangeRed
}

func randomCyberpunkColor() lipgloss.Color {
	return cyberpunkColors[rand.Intn(len(cyberpunkColors))]
}
