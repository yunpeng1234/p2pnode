package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"testing"
	"time"
)

func TestInput(t *testing.T) {

	// Start the first instance in a new terminal
	cmdOne := exec.Command("cmd.exe", "/C", "message -port 4000 -main")

	stdinOne, err := cmdOne.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}
	defer stdinOne.Close()

	stdOneOut, err := cmdOne.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	defer stdOneOut.Close()
	readerOne := bufio.NewReader(stdOneOut)

	if err := cmdOne.Start(); err != nil {
		t.Fatalf("Failed to start first instance: %v", err)
	}
	defer cmdOne.Process.Kill()

	// Sleep for a moment to ensure the first instance has started
	time.Sleep(2 * time.Second)

	// Start the second instance in a new terminal
	cmdTwo := exec.Command("cmd.exe", "/C", "message -port 4001")

	stdinTwo, err := cmdTwo.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}
	defer stdinTwo.Close()

	stdTwoOut, err := cmdTwo.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	defer stdTwoOut.Close()

	readerTwo := bufio.NewReader(stdTwoOut)

	if err := cmdTwo.Start(); err != nil {
		t.Fatalf("Failed to start second instance: %v", err)
	}
	defer cmdTwo.Process.Kill()

	// Sleep for a moment to ensure the second instance has started
	time.Sleep(2 * time.Second)

	io.WriteString(stdinTwo, "testing")
	scanner := bufio.NewScanner(readerOne)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Print("reading standard input:", err)
	}
	io.WriteString(stdinOne, "test")
	scanner = bufio.NewScanner(readerTwo)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Print("reading standard input:", err)
	}
	io.WriteString(stdinOne, "list peers")
	for {
		line, err := readerOne.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				// End of file reached
				break
			}
			fmt.Println("Error:", err)
			break
		}
		fmt.Print(line)
	}
}
