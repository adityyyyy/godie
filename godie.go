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

func write_file(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for row, line := range text_buffer {
		new_line := "\n"
		if row == len(text_buffer)-1 {
			new_line = ""
		}

		write_line := string(line) + new_line

		_, err := writer.WriteString(write_line)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}

	writer.Flush()
	modified = false
}

func insert_rune(event termbox.Event) {
	insert_runes := make([]rune, len(text_buffer[currentY])+1)
	copy(insert_runes[:currentX], text_buffer[currentY][:currentX])

	switch event.Key {
	case termbox.KeySpace:
		insert_runes[currentX] = rune(' ')

	case termbox.KeyTab:
		insert_runes[currentX] = rune(' ')

	default:
		insert_runes[currentX] = rune(event.Ch)
	}

	copy(insert_runes[currentX+1:], text_buffer[currentY][currentX:])

	text_buffer[currentY] = insert_runes
	currentX++
}

func delete_rune() {
	if currentX > 0 {
		currentX--
		delete_line := make([]rune, len(text_buffer[currentY])-1)
		copy(delete_line[:currentX], text_buffer[currentY][:currentX])
		copy(delete_line[currentX:], text_buffer[currentY][currentX+1:])
		text_buffer[currentY] = delete_line
	} else if currentY > 0 {
		append_line := make([]rune, len(text_buffer[currentY]))

		copy(append_line, text_buffer[currentY][currentX:])

		new_text_buffer := make([][]rune, len(text_buffer)-1)

		copy(new_text_buffer[:currentY], text_buffer[:currentY])
		copy(new_text_buffer[currentY:], text_buffer[currentY+1:])

		text_buffer = new_text_buffer
		currentY--
		currentX = len(text_buffer[currentY])

		insert_line := make([]rune, len(text_buffer[currentY])+len(append_line))

		copy(insert_line[:len(text_buffer[currentY])], text_buffer[currentY][:])
		copy(insert_line[len(text_buffer[currentY]):], append_line)

		text_buffer[currentY] = insert_line
	}
}

func insert_line() {
	right_line := make([]rune, len(text_buffer[currentY][currentX:]))
	left_line := make([]rune, len(text_buffer[currentY][:currentX]))
	new_text_buffer := make([][]rune, len(text_buffer)+1)

	copy(right_line, text_buffer[currentY][currentX:])
	copy(left_line, text_buffer[currentY][:currentX])
	copy(new_text_buffer, text_buffer[:currentY])

	new_text_buffer[currentY] = left_line
	currentY++
	currentX = 0
	new_text_buffer[currentY] = right_line

	copy(new_text_buffer[currentY+1:], text_buffer[currentY:])

	text_buffer = new_text_buffer
}

func scroll_text_buffer() {
	if currentY < offsetY {
		offsetY = currentY
	}

	if currentX < offsetX {
		offsetX = currentX
	}

	if currentY >= offsetY+ROWS {
		offsetY = currentY - ROWS + 1
	}

	if currentX >= offsetX+COLS {
		offsetX = offsetX - COLS + 1
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

func get_key() termbox.Event {
	var key_event termbox.Event
	switch event := termbox.PollEvent(); event.Type {
	case termbox.EventKey:
		key_event = event

	case termbox.EventError:
		panic(event.Err)
	}

	return key_event
}

func process_key_press() {
	key_event := get_key()

	if key_event.Key == termbox.KeyEsc {
		mode = 0
	} else if key_event.Ch != 0 {
		if mode == 1 {

			insert_rune(key_event)
			modified = true
		} else {
			switch key_event.Ch {
			case 'q':
				termbox.Close()
				os.Exit(0)

			case 'e':
				mode = 1

			case 'w':
				write_file(source_file)
			}
		}
	} else {
		switch key_event.Key {
		case termbox.KeyEnter:
			if mode == 1 {
				insert_line()
				modified = true
			}

		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if mode == 1 {
				delete_rune()
				modified = true
			} else {
				if currentX > 0 {
					currentX--
				} else if currentY > 0 {
					currentY--
					currentX = len(text_buffer[currentY])
				}
			}

		case termbox.KeyTab:
			if mode == 1 {
				for i := 0; i < 2; i++ {
					insert_rune(key_event)
				}
				modified = true
			}

		case termbox.KeySpace:
			if mode == 1 {
				insert_rune(key_event)
				modified = true
			}

		case termbox.KeyHome:
			currentX = 0

		case termbox.KeyEnd:
			currentX = len(text_buffer[currentY]) - 1

		case termbox.KeyPgup:
			if currentY-int(ROWS/4) >= 0 {
				currentY -= int(ROWS / 4)
			} else {
				currentY = 0
			}

		case termbox.KeyPgdn:
			if currentY+int(ROWS/4) < len(text_buffer) {
				currentY += int(ROWS / 4)
			} else {
				currentY = len(text_buffer) - 1
			}

		case termbox.KeyArrowUp:
			if currentY != 0 {
				currentY--
			}

		case termbox.KeyArrowDown:
			if currentY < len(text_buffer)-1 {
				currentY++
			}

		case termbox.KeyArrowLeft:
			if currentX > 0 {
				currentX--
			} else if currentY > 0 {
				currentY--
				currentX = len(text_buffer[currentY])
			}

		case termbox.KeyArrowRight:
			if currentX < len(text_buffer[currentY]) {
				currentX++
			} else if currentY < len(text_buffer)-1 {
				currentY++
				currentX = 0
			}
		}

		if currentX > len(text_buffer[currentY]) {
			currentX = len(text_buffer[currentY])
		}
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
		scroll_text_buffer()
		display_text_buffer()
		display_status_bar()
		termbox.SetCursor(currentX-offsetX, currentY-offsetY)
		termbox.Flush()
		process_key_press()
	}
}

func main() {
	run_editor()
}
