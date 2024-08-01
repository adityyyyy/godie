package main

import (
	"log"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var (
	ROWS, COLS       int
	offsetX, offsetY int
)

var text_buffer = [][]rune{
	{'h', 'e', 'l', 'l', 'o'},
	{'w', 'o', 'r', 'l', 'd'},
}

func display_text_buffer() {
	var row, col int

	for row = 0; row < ROWS; row++ {
		text_buffer_row := row + offsetY
		for col = 0; col < COLS; col++ {
			text_buffer_col := col + offsetX

			if text_buffer_row >= 0 && text_buffer_row < len(text_buffer) && text_buffer_col < len(text_buffer[text_buffer_row]) {
				if text_buffer[text_buffer_row][text_buffer_col] != '\t' {
					termbox.SetChar(col, row, text_buffer[text_buffer_row][text_buffer_col])
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
				}
			} else if row+offsetY > len(text_buffer) {
				termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
				termbox.SetChar(col, row, '\n')
			}
		}
	}
}

func print_message(col, row int, fg, bg termbox.Attribute, message string) {
	for _, ch := range message {
		termbox.SetCell(col, row, ch, fg, bg)

		col += runewidth.RuneWidth(ch)
	}
}

func run_editor() {
	err := termbox.Init()
	if err != nil {
		log.Fatal(err)
	}
	for {
		print_message(25, 11, termbox.ColorDefault, termbox.ColorDefault, "GO DIE")

		termbox.Flush()

		event := termbox.PollEvent()
		if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
			termbox.Close()
			break
		}
	}
}

func main() {
	run_editor()
}
