package main

import "strings"

func shapeName(sides int) string {
	if sides < 3 {
		return "Invalid (need at least 3 sides)"
	}

	switch sides {
	case 3:
		return "Triangle"
	case 4:
		return "Square"
	case 5:
		return "Pentagon"
	default:
		return "Polygon"
	}
}

func drawShape(sides int) string {
	if sides < 3 {
		return "Cannot draw shape with less than 3 sides"
	}

	if sides == 3 {
		return "  *\n * *\n*****"
	}

	size := sides
	var b strings.Builder

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i == 0 || i == size-1 || j == 0 || j == size-1 {
				b.WriteString("*")
			} else {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}
