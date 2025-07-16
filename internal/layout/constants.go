package layout

import "time"

// Layout constants - no more magic numbers!
const (
	// Spacing constants
	TitleSpacing     = 2  // Space after title
	HelpSpacing      = 2  // Space before help text
	ListPadding      = 2  // Additional padding for lists
	BorderPadding    = 1  // Padding inside borders
	
	// Input field constants
	MinInputWidth    = 20 // Minimum width for text inputs
	InputMargin      = 10 // Margin for input fields
	
	// Notification timing
	NotificationDuration = 2 * time.Second
	
	// List item spacing
	CheckboxSpacing  = 2  // Spaces between checkbox and title
	DescIndentation  = 4  // Indentation for descriptions
)

// CalculateListHeight calculates the appropriate height for lists
// based on window height and UI elements
func CalculateListHeight(windowHeight int) int {
	return windowHeight - TitleSpacing - HelpSpacing - ListPadding
}

// CalculateInputWidth calculates appropriate width for input fields
// based on window width
func CalculateInputWidth(windowWidth int) int {
	width := windowWidth - InputMargin
	if width < MinInputWidth {
		return MinInputWidth
	}
	return width
}