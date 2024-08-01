package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var (
	ROWS, COLS         int
	offsetX, offsetY   int
	currentX, currentY int
	mode               int
	source_file        string
	text_buffer        [][]rune
	undo_buffer        [][]rune
	copy_buffer        []rune
	modified           bool
)

func read_file(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		source_file = filename
		text_buffer = append(text_buffer, []rune{})
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	line_number := 0

	for scanner.Scan() {
		line := scanner.Text()
		text_buffer = append(text_buffer, []rune{})

		for i := 0; i < len(line); i++ {
			text_buffer[line_number] = append(text_buffer[line_number], rune(line[i]))
		}

		line_number++
	}

	if line_number == 0 {
		text_buffer = append(text_buffer, []rune{})
	}
}

func display_text_buffer() {
	var row, col int

	for row = 0; row < ROWS; row++ {
		text_buffer_row := row + offsetY
		for col = 0; col < COLS; col++ {
			text_buffer_col := col + offsetX

			if text_buffer_row >= 0 && text_buffer_row < len(text_buffer) &&
				text_buffer_col < len(text_buffer[text_buffer_row]) {
				if text_buffer[text_buffer_row][text_buffer_col] != '\t' {
					termbox.SetChar(col, row, text_buffer[text_buffer_row][text_buffer_col])
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
				}
			} else if row+offsetY > len(text_buffer)-1 {
				termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
				termbox.SetChar(col, row, rune('\n'))
			}
		}
	}
}

func display_status_bar() {
	var mode_status string
	var copy_status string
	var undo_status string
	var file_status string
	var cursor_status string

	if mode > 0 {
		mode_status = "EDIT: "
	} else {
		mode_status = "VIEW: "
	}

	filename_length := len(source_file)
	if filename_length > 8 {
		filename_length = 8
	}

	file_status = source_file[:filename_length] + " - " + strconv.Itoa(len(text_buffer)) + " lines"

	if modified {
		file_status += " modified"
	} else {
		file_status += " saved"
	}

	cursor_status = " Row-" + strconv.Itoa(currentY+1) + " Col-" + strconv.Itoa(currentX+1)

	if len(copy_buffer) > 0 {
		copy_status = " [Copy]"
	}

	if len(undo_buffer) > 0 {
		undo_status = " [Undo]"
	}

	used_space := len(
		mode_status,
	) + len(
		file_status,
	) + len(
		cursor_status,
	) + len(
		copy_status,
	) + len(
		undo_status,
	)

	spaces := strings.Repeat(" ", COLS-used_space)

	message := mode_status + file_status + copy_status + undo_status + spaces + cursor_status

	print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)
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

	if len(os.Args) > 1 {
		source_file = os.Args[1]
		read_file(source_file)
	} else {
		source_file = "out.txt"
		text_buffer = append(text_buffer, []rune{})
	}

	for {
		COLS, ROWS = termbox.Size()
		ROWS--

		if COLS < 80 {
			COLS = 80
		}

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		display_text_buffer()
		display_status_bar()
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
